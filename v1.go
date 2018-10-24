package wpgx

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/glog"
	"gopkg.in/jackc/pgx.v2"
)

// Менеджер подключений
var poolV1 = &connMgrV1{conns: make(map[string]*v1, 3)}

// V1 - Получение объекта подключения из пула. Гарантирует возврат объекта.
// Это нужно для упрощения кода использования. Ошибку следует перехватывать в дочерних методах
func V1(name string) IConn {
	// Получаем объект. Может быть пустым
	conn := poolV1.get(name)

	// Если не повезло - создаем новый с этим именем и добавляем
	if conn == nil {
		conn = newV1(name)
		poolV1.set(name, conn)
	}

	// В любом случае возвращаем объект
	return conn
}

// ConnectV1 - Создание и регистрация нового подключения к БД 1-й версии
// connStr - строка подключения
// reservePath - каталог для сохранения запросов и их резервных данных
// isDefault - считать ли это подключение по-умолчанию
func ConnectV1(name, connStr, reservePath string, isDefault bool) (err error) {
	// Если это подключение по-умолчанию - делаем имя пустым
	if isDefault {
		name = ""
	}

	// Получаем объект подключения. Используется еще старое имя по-умолчанию.
	conn := V1(name)

	// В любом случае, подключение нужно закрыть. Если было новое - ничего страшного
	conn.Close()

	// Проверяем, нашего ли типа этот объект. Если нет - замещаем новым
	if _, ok := conn.(*v1); !ok {
		cv1 := newV1(name)
		poolV1.set(name, cv1)
		conn = cv1
	}

	// Устанавливаем путь до резерва
	if err = conn.SetReservePath(reservePath); err != nil {
		return
	}

	// Устанавливаем новое подключение
	if err = conn.Connect(connStr); err != nil {
		return
	}

	// Подготавливаем все запросы подключения
	return conn.PrepareQueries()
}

// ClearV1 - Очистка пула подключений 1-й версии
func ClearV1() { poolV1.clear() }

// Менеджер подключений - для управления глобальным хранилищем подключений и запросов
type connMgrV1 struct {
	sync.RWMutex
	conns map[string]*v1
}

// Получение подключения из пула
func (c *connMgrV1) get(name string) *v1 {
	c.RLock()
	defer c.RUnlock()

	return c.conns[name]
}

// Сохранение подключения в пул
func (c *connMgrV1) set(name string, conn *v1) {
	c.Lock()
	defer c.Unlock()

	c.conns[name] = conn
}

// Очистка пула
func (c *connMgrV1) clear() {
	c.Lock()
	defer c.Unlock()

	for i := range c.conns {
		c.conns[i].Close()
	}

	c.conns = make(map[string]*v1, 3)
}

// Базовая структура подключения к БД
// Обязана реализовывать IConn
type v1 struct {
	sync.RWMutex

	name        string
	connStr     string
	reservePath string

	pool    *pgx.ConnPool
	queries map[string]*qv1
}

// Создание объекта подключения 1-й версии
func newV1(name string) *v1 { return &v1{name: name, queries: make(map[string]*qv1, 32)} }

// Проверка на валидность
func (v *v1) IsNull() bool { return v.pool == nil }

// Установка подключения к серверу
func (v *v1) Connect(connStr string) (err error) {
	var cfg pgx.ConnConfig

	if cfg, err = pgx.ParseConnectionString(connStr); err != nil {
		return
	}

	if v.pool, err = pgx.NewConnPool(pgx.ConnPoolConfig{MaxConnections: 256, ConnConfig: cfg}); err != nil {
		return
	}

	return nil
}

// Установка пути для резервных копий
func (v *v1) SetReservePath(path string) (err error) {
	if path, err = filepath.Abs(path); err != nil {
		return
	}

	var info os.FileInfo

	if info, err = os.Stat(path); err != nil {
		return
	}

	if !info.IsDir() {
		return fmt.Errorf("Файл `%s` не является каталогом", path)
	}

	var file *os.File
	test := filepath.Join(path, "write.test")

	if file, err = os.Create(test); err != nil {
		return
	}

	if err = file.Close(); err != nil {
		return
	}

	if err = os.RemoveAll(test); err != nil {
		return
	}

	v.reservePath = path
	return nil
}

// Регистрация запроса
func (v *v1) Register(text string, names []string) string {
	q := &qv1{text: text, names: names}

	sum := sha1.Sum([]byte(text))
	key := hex.EncodeToString(sum[:])

	v.queries[key] = q
	return key
}

// Подготовка всех зарегистрированных для этого подключения запросов
func (v *v1) PrepareQueries() (err error) {
	for key := range v.queries {
		if err = v.queries[key].Prepare(v, key); err != nil {
			return fmt.Errorf("Ошибка подготовки запроса `%s` [%+v]", v.queries[key].text, err)
		}
	}

	return nil
}

// RawSelect - Выборка из базы одного абстрактного объекта
func (v *v1) RawSelect(key string, args ...interface{}) (list RawList, err error) {
	if v.pool == nil {
		err = fmt.Errorf("Отсутствует подключение к серверу БД `%s`", v.name)
		return
	}

	defer func() {
		if err != nil {
			glog.Warningf("Запрос `%s`: ошибка SQL - %s. Параметры:\n%+v", key, err, args)
		}
	}()

	var rows *pgx.Rows

	if rows, err = v.pool.Query(key, args...); err != nil {
		return
	}

	defer rows.Close()

	// Колонки можно получить один раз
	names := rows.FieldDescriptions()
	list = make(RawList, 0, 128)

	// Получение данных
	for rows.Next() {
		strs := make([]String, len(names))
		places := make([]interface{}, len(names))
		for i := range strs {
			places[i] = &strs[i]
		}

		// Сканируем данные из источника
		if err = rows.Scan(places...); err != nil {
			return
		}

		// Объединяем с колонками
		obj := make(map[string]string, len(names))
		for i := range names {
			if strs[i].NullString.Valid {
				obj[names[i].Name] = strs[i].NullString.String
			}
		}

		// Добавляем в множество.
		list = append(list, obj)
	}

	// Если во время итерации были ошибки, они должны быть обработаны тут. Саму ошибку также палим в лог
	return list, rows.Err()
}

// Выборка из БД
func (v *v1) Select(result IList, key string, args ...interface{}) (err error) {
	if v.pool == nil {
		return fmt.Errorf("Отсутствует подключение к серверу БД `%s`", v.name)
	}

	defer func() {
		if err != nil {
			glog.Warningf("Запрос `%s`: ошибка SQL - %s. Параметры:\n%+v", key, err, args)
		}
	}()

	var rows *pgx.Rows

	if rows, err = v.pool.Query(key, args...); err != nil {
		return err
	}

	return loadRowsV1(result, rows)
}

// Выборка из БД единичной сущности
func (v *v1) SelectOne(result IListItem, key string, args ...interface{}) (err error) {
	if v.pool == nil {
		return fmt.Errorf("отсутствует подключение к серверу БД `%s`", v.name)
	}

	defer func() {
		if err != nil {
			glog.Warningf("запрос `%s`: ошибка SQL - %s. Параметры:\n%+v", key, err, args)
		}
	}()

	var rows *pgx.Rows

	if rows, err = v.pool.Query(key, args...); err != nil {
		return err
	}

	return loadRowV1(result, rows)
}

// Выполнение запроса в БД (режим автокоммита)
func (v *v1) Exec(key string, args ...interface{}) (err error) {
	if v.pool == nil {
		return fmt.Errorf("Отсутствует подключение к серверу БД `%s`", v.name)
	}

	defer func() {
		if err != nil {
			glog.Warningf("Запрос `%s`: ошибка SQL - %s. Параметры:\n%+v", key, err, args)
		}
	}()

	if _, err = v.pool.Exec(key, args...); err != nil {
		return
	}

	return nil
}

// Сохранение объекта в БД (режим автокоммита с резервированием при неудаче)
func (v *v1) Save(item IListItem, key string, result IList) (err error) {
	if v.pool == nil {
		return fmt.Errorf("Отсутствует подключение к серверу БД `%s`", v.name)
	}

	var args []interface{}

	for i := range v.queries[key].names {
		args = append(args, item.Placeholder(v.queries[key].names[i]))
	}

	// В случае ошибки резервируем аргументы запроса.
	// Имя получаемого файла состоит из ключа запроса и хеша аргументов
	defer func() {
		if err == nil {
			return
		}

		text, ex := json.MarshalIndent(args, "", "  ")
		if ex != nil {
			glog.Errorf("Ошибка резервирования данных: `%s`\n%+v", ex, args)
		}

		sum := sha1.Sum([]byte(text))
		hash := hex.EncodeToString(sum[:])

		if ex = ioutil.WriteFile(filepath.Join(v.reservePath, key+"_"+hash+".json"), text, 0755); ex != nil {
			glog.Errorf("Ошибка резервирования данных: `%s`\n%+v", ex, args)
		}
	}()

	if result == nil {
		return v.Exec(key, args...)
	}
	return v.Select(result, key, args...)
}

// Создание нового объекта транзакции
func (v *v1) Tx() ITx { return &txv1{db: v} }

// Закрытие подключения
func (v *v1) Close() {
	v.Lock()
	defer v.Unlock()

	// Закрытие подготовленных запросов
	for i := range v.queries {
		v.queries[i].Close()
	}

	// Закрытие подключения
	if v.pool != nil {
		v.pool.Close()
		v.pool = nil
	}
}

// Базовая структура транзакции БД
// Обязана реализовать ITx
type txv1 struct {
	db *v1
	tx *pgx.Tx
}

// RawSelect - Выборка из базы одного абстрактного объекта
func (t *txv1) RawSelect(key string, args ...interface{}) (list RawList, err error) {
	if t.db == nil || t.db.pool == nil {
		err = fmt.Errorf("отсутствует подключение к серверу БД `%s`", t.db.name)
		return
	}
	defer func() {
		if err != nil {
			glog.Warningf("запрос `%s`: ошибка SQL - %s. Параметры:\n%+v", key, err, args)
		}
	}()

	// Если транзакция еще не активна - начинаем новую
	if t.tx == nil {
		if t.tx, err = t.db.pool.Begin(); err != nil {
			return
		}
	}

	var rows *pgx.Rows

	if rows, err = t.tx.Query(key, args...); err != nil {
		return
	}

	defer rows.Close()

	// Колонки можно получить один раз
	names := rows.FieldDescriptions()
	list = make(RawList, 0, 128)

	// Получение данных
	for rows.Next() {
		strs := make([]String, len(names))
		places := make([]interface{}, len(names))
		for i := range strs {
			places[i] = &strs[i]
		}

		// Сканируем данные из источника
		if err = rows.Scan(places...); err != nil {
			return
		}

		// Объединяем с колонками
		obj := make(map[string]string, len(names))
		for i := range names {
			if strs[i].NullString.Valid {
				obj[names[i].Name] = strs[i].NullString.String
			}
		}

		// Добавляем в множество.
		list = append(list, obj)
	}

	// Если во время итерации были ошибки, они должны быть обработаны тут. Саму ошибку также палим в лог
	return list, rows.Err()
}

// Выборка из БД
func (t *txv1) Select(result IList, key string, args ...interface{}) (err error) {
	if t.db == nil || t.db.pool == nil {
		return fmt.Errorf("Отсутствует подключение к серверу БД `%s`", t.db.name)
	}

	defer func() {
		if err != nil {
			glog.Warningf("Запрос `%s`: ошибка SQL - %s. Параметры:\n%+v", key, err, args)
		}
	}()

	// Если транзакция еще не активна - начинаем новую
	if t.tx == nil {
		if t.tx, err = t.db.pool.Begin(); err != nil {
			return
		}
	}

	var rows *pgx.Rows

	if rows, err = t.tx.Query(key, args...); err != nil {
		return err
	}

	return loadRowsV1(result, rows)
}

// Выборка из БД единичной сущности
func (t *txv1) SelectOne(result IListItem, key string, args ...interface{}) (err error) {
	if t.db == nil || t.db.pool == nil {
		return fmt.Errorf("отсутствует подключение к серверу БД `%s`", t.db.name)
	}

	defer func() {
		if err != nil {
			glog.Warningf("запрос `%s`: ошибка SQL - %s. Параметры:\n%+v", key, err, args)
		}
	}()

	// Если транзакция еще не активна - начинаем новую
	if t.tx == nil {
		if t.tx, err = t.db.pool.Begin(); err != nil {
			return
		}
	}

	var rows *pgx.Rows

	if rows, err = t.tx.Query(key, args...); err != nil {
		return err
	}

	return loadRowV1(result, rows)
}

// Выполнение запроса в БД
func (t *txv1) Exec(key string, args ...interface{}) (err error) {
	if t.db == nil || t.db.pool == nil {
		return fmt.Errorf("Отсутствует подключение к серверу БД `%s`", t.db.name)
	}

	defer func() {
		if err != nil {
			glog.Warningf("Запрос `%s`: ошибка SQL - %s. Параметры:\n%+v", key, err, args)
		}
	}()

	// Если транзакция еще не активна - начинаем новую
	if t.tx == nil {
		if t.tx, err = t.db.pool.Begin(); err != nil {
			return
		}
	}

	if _, err = t.tx.Exec(key, args...); err != nil {
		return
	}

	return nil
}

// Сохранение объекта в БД (без резервирования)
func (t *txv1) Save(item IListItem, key string, result IList) (err error) {
	if t.db == nil || t.db.pool == nil {
		return fmt.Errorf("Отсутствует подключение к серверу БД `%s`", t.db.name)
	}

	var args []interface{}

	for i := range t.db.queries[key].names {
		args = append(args, item.Placeholder(t.db.queries[key].names[i]))
	}

	if result == nil {
		return t.Exec(key, args...)
	}
	return t.Select(result, key, args...)
}

// Завершение транзакции
func (t *txv1) Close(ok bool) error {
	// Если транзакция и не была активна - можно закрывать без проблем
	if t.tx == nil {
		return nil
	}

	// Если нужен явный откат - делаем
	if !ok {
		return t.tx.Rollback()
	}

	// На всякий случай - откат. Затем пробуем коммит
	defer t.tx.Rollback()
	return t.tx.Commit()
}

// Базовая структура запроса к БД
type qv1 struct {
	text  string
	stmt  *pgx.PreparedStatement
	names []string
}

// Подготовка запроса
func (q *qv1) Prepare(v *v1, key string) (err error) {
	if v.pool == nil {
		return fmt.Errorf("Отсутствует подключение к БД `%s`", v.name)
	}

	if q.stmt, err = v.pool.Prepare(key, q.text); err != nil {
		return
	}

	if v.reservePath != "" {
		if err = ioutil.WriteFile(filepath.Join(v.reservePath, key+".sql"), []byte(q.text), 0755); err != nil {
			return
		}
	}

	return nil
}

// Закрытие ресурсов и освобождение запроса
func (q *qv1) Close() {
	if q.stmt != nil {
		q.stmt = nil
	}
}

func loadRowsV1(result IList, rows *pgx.Rows) (err error) {
	defer rows.Close()

	// Колонки можно получить один раз
	names := rows.FieldDescriptions()
	places := make([]interface{}, len(names))

	// Получение данных
	for rows.Next() {
		// Формируем очередной экземпляр и список приемников
		item := result.NewItem()

		// Формируем список для получения данных. Можно было вынести в интерфейс, но
		// делать это одинаково и тут, и в реализации интерфейса - так что лучше тут
		for i := range names {
			places[i] = item.Placeholder(names[i].Name)
		}

		// Сканируем данные из источника
		if err = rows.Scan(places...); err != nil {
			return
		}

		// Добавляем в множество.
		// Считаем, что корректность добавления - проблема коллекции
		if err = result.Append(item); err != nil {
			return
		}
	}

	// Если во время итерации были ошибки, они должны быть обработаны тут. Саму ошибку также палим в лог
	return rows.Err()
}

func loadRowV1(result IListItem, rows *pgx.Rows) (err error) {
	defer rows.Close()

	// Колонки можно получить один раз
	names := rows.FieldDescriptions()
	places := make([]interface{}, len(names))

	// Получение данных
	for rows.Next() {
		// Формируем очередной экземпляр и список приемников
		// Формируем список для получения данных. Можно было вынести в интерфейс, но
		// делать это одинаково и тут, и в реализации интерфейса - так что лучше тут
		for i := range names {
			places[i] = result.Placeholder(names[i].Name)
		}

		// Сканируем данные из источника
		if err = rows.Scan(places...); err != nil {
			return
		}
		// если единичная выборка - всегда выбираем первую строку - на остальные пофиг
		break
	}

	// Если во время итерации были ошибки, они должны быть обработаны тут. Саму ошибку также палим в лог
	return rows.Err()
}
