package wpgx

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
