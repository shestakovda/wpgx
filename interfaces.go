package wpgx

// Translator helps to isolate struct from loading procedure
//
// Translate is intended to associate structure fields with their names
type Translator interface {
	Translate(name string) interface{}
}

// Collector is a generic collection. It can use lists, maps or channels inside
//
// NewItem prepares a translator entity for future loads
//
// Collect puts a new translator item into the inside collection
type Collector interface {
	NewItem() Translator
	Collect(item Translator)
}

// Dealer is an active subject, like an opened transaction, for query performing
//
// Deal! It executes query and loads result into a data collector. Pass nil when no result needed
//
// Load gets just one item from the database. When no collection needed
//
// Save inserts item into database. Result may need for getting new ID or properties
//
// Jail (aka Close) ends a transaction with commit or rollback respective to the flag
type Dealer interface {
	Deal(result Collector, query string, args ...interface{}) error

	Load(item Translator, query string, args ...interface{}) error
	Save(item Translator, query string, result Collector) error

	Jail(commit bool) error
}

// Connector is the main database connection manager
// As a Dealer it can execute queries in a default transaction
//
// NewDealer spawns new dealer on the street. It needs to be jailed (closed)
//
// Reserve path is for saving args of failed queries. Useful for debug or data restore
//
// Register saves query for further preparation. Must be called before Connect
//
// Connect method initialize a new connection pool with uri in a connection string format
//
// Close closes all free dealers with rollback
type Connector interface {
	Dealer

	NewDealer() Dealer

	Reserve(path string) error

	Register(query string, names []string) string

	Connect(uri string) error

	Close()
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
