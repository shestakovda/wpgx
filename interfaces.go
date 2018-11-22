package wpgx

// Translator helps to isolate model from loading procedure
//
// Translate is intended to associate structure fields with their names
type Translator interface {
	Translate(name string) interface{}
}

// Shaper helps to make database model from business model and vice versa
//
// Extrude makes a database model from business data
//
// Receive fills the data from a database model
type Shaper interface {
	Extrude() Translator
	Receive(Translator) error
}

// Collector is a generic collection. It can use lists, maps or channels inside
//
// NewItem prepares a translator entity for future loads
//
// Collect puts a new translator item into the inside collection
type Collector interface {
	NewItem() Shaper
	Collect(item Shaper) error
}

/******************************************************************************/

// IListItem - Интерфейс для работы с элементом данных в БД
type IListItem interface {

	// Получение указателя на поле элемента, которое предназначено данному имени
	Placeholder(name string) interface{}
}

// IList - Интерфейс для сканирования списка данных из БД в структуры
type IList interface {

	// Создание нового элемента
	NewItem() IListItem

	// Добавление элемента
	Append(IListItem) error
}

// ITx - Интерфейс транзакции БД
type ITx interface {

	// Выборка из БД
	Select(result IList, key string, args ...interface{}) error

	// Выборка из БД единичной сущности
	SelectOne(result IListItem, key string, args ...interface{}) error

	// Абстрактная выборка
	RawSelect(key string, args ...interface{}) (list RawList, err error)

	// Выполнение запроса в БД
	Exec(key string, args ...interface{}) error

	// Сохранение объекта в БД (без резервирования)
	Save(src IListItem, key string, result IList) error

	// Завершение транзакции. Если ок - то коммит
	Close(ok bool) error
}

// IConn - Интерфейс для работы с БД
type IConn interface {

	// Проверка на валидность
	IsNull() bool

	// Установка подключения к серверу
	Connect(connStr string) error

	// Установка пути для резервных копий
	SetReservePath(path string) error

	// Регистрация запроса. Возвращает ключ для его выполнения
	Register(text string, names []string) string

	// Подготовка всех зарегистрированных для этого подключения запросов
	PrepareQueries() error

	// Абстрактная выборка
	RawSelect(key string, args ...interface{}) (list RawList, err error)

	// Выборка из БД
	Select(result IList, key string, args ...interface{}) error

	// Выборка из БД единичной сущности
	SelectOne(result IListItem, key string, args ...interface{}) error

	// Выполнение запроса в БД (режим автокоммита)
	Exec(key string, args ...interface{}) error

	// Сохранение объекта в БД (режим автокоммита с резервированием при неудаче)
	Save(src IListItem, key string, result IList) error

	// Начало новой транзакции
	Tx() ITx

	// Закрытие подключения и освобождение ресурсов
	// Если есть связанные открытые транзакции - по ним будет Rollback
	Close()
}
