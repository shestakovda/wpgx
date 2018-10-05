package wpgx

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	// "fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const reservePath = "./test_files/clear"

var connStr string

func TestMain(m *testing.M) {
	// Строка подключения к тестовой БД
	connStrBytes, err := ioutil.ReadFile("./test.db.conn")
	if err != nil {
		log.Fatal(err)
	}
	connStr = string(connStrBytes)

	if err = os.RemoveAll(reservePath); err != nil {
		log.Fatal(err)
	}
	if err = os.MkdirAll(reservePath, 0755); err != nil {
		log.Fatal(err)
	}

	db := V1("")

	dropTable := db.Register(`DROP TABLE IF EXISTS test_db;`, nil)
	createTable := db.Register(`
        CREATE TABLE test_db
        (
            num INTEGER,
            id  TEXT,
            c2  TEXT,
            c3  TEXT,
            c4  TEXT,
            c5  TEXT,
            c6  TEXT,
            c7  TEXT,
            c8  TEXT,
            c9  TEXT
        );
    `, nil)

	if err = ConnectV1("", connStr, reservePath, true); err != nil {
		log.Fatal(err)
	}

	if err = db.Exec(dropTable); err != nil {
		log.Fatal(err)
	}

	if err = db.Exec(createTable); err != nil {
		log.Fatal(err)
	}

	res := m.Run()

	ClearV1()

	os.Exit(res)
}

type testListV1 struct {
	items []*testItemV1
	index map[string]*testItemV1
}

func newTestListV1(capacity int) *testListV1 {
	return &testListV1{
		items: make([]*testItemV1, 0, capacity),
		index: make(map[string]*testItemV1, capacity),
	}
}

func (t *testListV1) NewItem() IListItem {
	return new(testItemV1)
}

func (t *testListV1) Append(item IListItem) error {
	switch k := item.(type) {
	case *testItemV1:
		{
			t.items = append(t.items, k)
			t.index[k.ID] = k
		}
	}
	return nil
}

type testItemV1 struct {
	ID  string `db:"id"`
	Num int    `db:"num"`
	C1  string `db:"c1"`
	C2  string `db:"c2"`
	C3  string `db:"c3"`
	C4  string `db:"c4"`
	C5  string `db:"c5"`
	C6  string `db:"c6"`
	C7  string `db:"c7"`
	C8  string `db:"c8"`
	C9  string `db:"c9"`
}

func (t *testItemV1) Placeholder(name string) interface{} {
	switch name {
	case "id":
		return &t.ID
	case "num":
		return &t.Num
	case "c1":
		return &t.C1
	case "c2":
		return &t.C2
	case "c3":
		return &t.C3
	case "c4":
		return &t.C4
	case "c5":
		return &t.C5
	case "c6":
		return &t.C6
	case "c7":
		return &t.C7
	case "c8":
		return &t.C8
	case "c9":
		return &t.C9
	}
	return nil
}

type testCrashV1 struct {
	ID  string `db:"id"`
	Num int    `db:"num"`
	C1  string `db:"c1"`
}

func (t testCrashV1) Placeholder(name string) (holder interface{}) {
	switch name {
	case "id":
		return &t.ID
	case "num":
		return &t.Num
	case "c1":
		return &t.C1
	}
	return
}

type testEntityV1 struct {
	ID  string `db:"id"`
	Num int    `db:"num"`
	C1  string `db:"c1"`
	C2  string `db:"c2"`
	C3  string `db:"c3"`
	C4  string `db:"c4"`
	C5  string `db:"c5"`
	C6  string `db:"c6"`
	C7  string `db:"c7"`
	C8  string `db:"c8"`
	C9  string `db:"c9"`
}

func (t testEntityV1) Placeholder(name string) (holder interface{}) {
	switch name {
	case "id":
		return &t.ID
	case "num":
		return &t.Num
	case "c1":
		return &t.C1
	case "c2":
		return &t.C2
	case "c3":
		return &t.C3
	case "c4":
		return &t.C4
	case "c5":
		return &t.C5
	case "c6":
		return &t.C6
	case "c7":
		return &t.C7
	case "c8":
		return &t.C8
	case "c9":
		return &t.C9
	}
	return
}

func Test_V1(t *testing.T) {
	// Проверяем что второй вызов функции добавления подключения не вернул ошибку
	err := ConnectV1("", connStr, reservePath, true)
	assert.NoError(t, err)
	err = ConnectV1("", connStr, reservePath, true)
	assert.NoError(t, err)

	// Проверяем, что всегда можем что-то получить
	db := V1("test")
	assert.NotNil(t, db)
	db = V1("")
	assert.NotNil(t, db)

	// Проверка строки подключения
	err = db.Connect("postgres://test:lol@localhost:42140/purpurpur")
	assert.EqualError(t, err, "dial tcp 127.0.0.1:42140: connect: connection refused")
	err = db.Connect(connStr)
	assert.NoError(t, err)

	// Проверка пути сохранения резерва
	err = db.SetReservePath("-")
	assert.EqualError(t, err, "stat /go/src/operator/wpgx/-: no such file or directory")
	err = db.SetReservePath(reservePath)
	assert.NoError(t, err)

	// проверяем что сохраняется шаблон запроса
	key := db.Register("SELECT 1 WHERE 1 = $1", nil)
	err = ConnectV1("", connStr, reservePath, true)
	assert.NoError(t, err)

	data, err := ioutil.ReadFile(filepath.Join(reservePath, key+".sql"))
	assert.NoError(t, err)
	assert.Equal(t, "SELECT 1 WHERE 1 = $1", string(data))
}

func Test_V1_Select(t *testing.T) {
	crash1 := V1("crash1")
	crash2 := V1("crash2")
	test := V1("test")

	crash1.Register("SELECT from", nil)
	key0 := crash2.Register("SELECT null as ID", nil)
	key1 := test.Register("SELECT 1 as num, 'test1' as id;", nil)
	key2 := test.Register("SELECT 2 as num, 'test2' as ID;", nil)
	key3 := test.Register("SELECT 3 as num, 'test3' as ID;", nil)
	key4 := test.Register("SELECT 4;", nil)
	key5 := test.Register("SELECT WHERE true = false;", nil)

	err := ConnectV1("crash1", connStr, reservePath, false)
	assert.EqualError(t, err, "Ошибка подготовки запроса `SELECT from` [ERROR: syntax error at end of input (SQLSTATE 42601)]")

	err = ConnectV1("crash2", connStr, reservePath, false)
	assert.NoError(t, err)

	err = ConnectV1("test", connStr, reservePath, false)
	assert.NoError(t, err)

	list := newTestListV1(4)

	assert.EqualError(t, crash2.Select(list, key0), "can't scan into dest[0]: Cannot decode null into string")
	assert.Nil(t, test.Select(list, key1))
	assert.Nil(t, test.Select(list, key2))
	assert.Nil(t, test.Select(list, key3))
	assert.Nil(t, test.Select(list, key4))
	assert.Nil(t, test.Select(list, key5))
	assert.Len(t, list.items, 4)
	assert.Len(t, list.index, 4)

	assert.Equal(t, 2, list.index["test2"].Num)
	assert.Equal(t, 0, list.items[3].Num)
}

func Test_V1_Save(t *testing.T) {
	db := V1("")

	keySelect := db.Register("SELECT id, num, c2 FROM test_db WHERE id=$1", []string{"id"})
	key_only_save := db.Register(`
        INSERT INTO test_db (
            id,
            num,
            c2
        ) VALUES (
            $1,
            $2,
            $3
        )`,
		[]string{"id", "num", "c2"},
	)
	key_update := db.Register(`
        UPDATE test_db SET
            id = $1,
            num = $2,
            c2 = $3
        WHERE id = $1`,
		[]string{"id", "num", "c2"},
	)
	key_save_and_return := db.Register(`
        INSERT INTO test_db (
            id,
            num,
            c2
        ) VALUES (
            $1,
            $2,
            $3
        ) RETURNING
            id, num, c2`,
		[]string{"id", "num", "c2"},
	)

	err := ConnectV1("", connStr, reservePath, true)
	assert.NoError(t, err)

	//Делаем запись в бд
	model := testEntityV1{ID: "firstnah", Num: 1, C2: "C1"}
	err = db.Save(model, key_only_save, nil)
	assert.NoError(t, err)

	//Проверяем наличие записи
	savelist := newTestListV1(1)
	assert.NoError(t, db.Select(savelist, keySelect, model.ID))
	assert.Equal(t, 1, len(savelist.items))
	assert.Equal(t, "firstnah", savelist.items[0].ID)
	assert.Equal(t, 1, savelist.items[0].Num)
	assert.Equal(t, "C1", savelist.items[0].C2)

	//Обновляем запись
	model.C2 = "C111"
	err = db.Save(model, key_update, nil)
	assert.NoError(t, err)

	//Проверяем что запись обновилась
	savelist1 := newTestListV1(1)
	assert.NoError(t, db.Select(savelist1, keySelect, model.ID))
	assert.Equal(t, 1, len(savelist1.items))
	assert.Equal(t, "firstnah", savelist1.items[0].ID)
	assert.Equal(t, 1, savelist1.items[0].Num)
	assert.Equal(t, "C111", savelist1.items[0].C2)

	//Делаем новую запись в бд и возвращаем полученный результат (Returning)
	result := newTestListV1(1)
	model = testEntityV1{ID: "secondnah", Num: 2, C2: "C2"}
	err = db.Save(model, key_save_and_return, result)
	assert.NoError(t, err)

	//Проверяем что вернулось
	assert.Equal(t, 1, len(result.items))
	assert.Equal(t, "secondnah", result.items[0].ID)
	assert.Equal(t, 2, result.items[0].Num)
	assert.Equal(t, "C2", result.items[0].C2)

	//Проверяем наличие записи в бд
	savelist2 := newTestListV1(1)
	assert.NoError(t, db.Select(savelist2, keySelect, model.ID))
	assert.Equal(t, 1, len(savelist2.items))
	assert.Equal(t, "secondnah", savelist2.items[0].ID)
	assert.Equal(t, 2, savelist2.items[0].Num)
	assert.Equal(t, "C2", savelist2.items[0].C2)
}

func Test_V1_CrashSave(t *testing.T) {
	db := V1("")
	sql := `INSERT INTO test_db (
            id,
            num
        ) VALUES (
            $1,
            $2
        )`
	key := db.Register(sql, []string{"num", "id"})

	err := ConnectV1("", connStr, reservePath, true)
	assert.NoError(t, err)

	//Делаем запись в бд
	model := testCrashV1{ID: "firstnah", Num: 5}
	err = db.Save(model, key, nil)
	assert.Equal(t, "cannot encode int8 into oid 25", err.Error())

	var args []interface{}
	args = append(args, 5)
	args = append(args, "firstnah")
	text, err := json.MarshalIndent(args, "", "  ")
	hash := sha1.Sum([]byte(text))
	keyArgs := hex.EncodeToString(hash[:])
	data, err := ioutil.ReadFile(filepath.Join(reservePath, key+"_"+keyArgs+".json"))
	assert.NoError(t, err)
	assert.Equal(t, `[
  5,
  "firstnah"
]`, string(data))

}

func Test_V1_Register(t *testing.T) {
	// Проверяем что нет обьекта подключения
	assert.Nil(t, poolV1.conns["register_test"])
	key := V1("register_test").Register("SELECT 5", []string{"num"})

	//проверяем что обьект подключения добавился, плюс проверяем что добавился запрос
	assert.NotNil(t, poolV1.conns["register_test"])
	assert.NotNil(t, poolV1.conns["register_test"].queries[key])
	assert.Equal(t, "SELECT 5", poolV1.conns["register_test"].queries[key].text)
	assert.Equal(t, []string{"num"}, poolV1.conns["register_test"].queries[key].names)

	//добавляем еще один запрос
	key1 := V1("register_test").Register("SELECT 6", nil)

	//проверяем что старые данные не стерлись
	assert.NotNil(t, poolV1.conns["register_test"].queries[key])
	assert.Equal(t, "SELECT 5", poolV1.conns["register_test"].queries[key].text)
	assert.Equal(t, []string{"num"}, poolV1.conns["register_test"].queries[key].names)

	//проверяем налия новых данных
	assert.NotNil(t, poolV1.conns["register_test"].queries[key1])
	assert.Equal(t, "SELECT 6", poolV1.conns["register_test"].queries[key1].text)
	assert.Equal(t, []string(nil), poolV1.conns["register_test"].queries[key1].names)
}

func Test_ClearV1(t *testing.T) {
	err := ConnectV1("test_clear", connStr, reservePath, false)
	assert.NoError(t, err)

	// Проверяем, что есть подключение
	assert.NotNil(t, poolV1.conns["test_clear"].pool)

	//Закрываем все
	ClearV1()

	//проверяем отсутствие подключения
	assert.Nil(t, poolV1.conns["test_clear"])
}

func Benchmark_DBv2_Select(b *testing.B) {
	db := V1("")

	var err error

	if err = db.Exec(`
DROP TABLE IF EXISTS test_db;
SELECT 1 as num, t::text as ID, t::text as c2, t::text as c3, t::text as c4,
t::text as c5, t::text as c6, t::text as c7, t::text as c8, t::text as c9 INTO test_db
FROM generate_series('2008-01-01'::timestamp, '2010-01-01', '1 hour') t;
    `); err != nil {
		b.Fatal(err)
	}

	key1 := db.Register(`SELECT * FROM test_db LIMIT 1;`, nil)
	key2 := db.Register(`SELECT * FROM test_db;`, nil)

	err = ConnectV1("", connStr, reservePath, true)
	assert.NoError(b, err)

	list := newTestListV1(100000)
	b.Run("t1/v2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err = db.Select(list, key1); err != nil {
				b.Fatal(err)
			}
		}
	})

	list = newTestListV1(100000)
	b.Run("t2/v2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err = db.Select(list, key2); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Test_StringList_v2(t *testing.T) {
	err := ConnectV1("", connStr, reservePath, true)
	assert.NoError(t, err)

	names := make(Strings, 0, 100)

	assert.NoError(t, V1("").Select(&names, `SELECT t::text FROM generate_series('2008-01-01'::timestamp, '2008-01-02', '1 hour') t;`))
	assert.Len(t, names, 25)
}

func TestDBv2_TX(t *testing.T) {
	db := V1("")

	keySave := db.Register(`
        INSERT INTO test_db (
            id,
            num,
            c2
        ) VALUES (
            $1,
            $2,
            $3
        )`,
		[]string{"id", "num", "c2"},
	)
	keySelect := db.Register(`SELECT id, num, c2 FROM test_db WHERE id = $1`, []string{"id"})

	err := ConnectV1("", connStr, reservePath, true)
	assert.NoError(t, err)

	tx := db.Tx()

	names1 := make(Strings, 0, 100)
	names2 := make(Strings, 0, 100)

	assert.NoError(t, tx.Select(&names1, `SELECT t::text FROM generate_series('2008-01-01'::timestamp, '2008-01-02', '1 hour') t;`))
	assert.NoError(t, tx.Select(&names2, `SELECT t::text FROM generate_series('2008-01-01'::timestamp, '2008-01-03', '1 hour') t;`))
	assert.Len(t, names1, 25)
	assert.Len(t, names2, 49)
	assert.NoError(t, tx.Close(false))
	err = tx.Select(&names1, `SELECT t::text FROM generate_series('2008-01-01'::timestamp, '2008-01-02', '1 hour') t;`)
	assert.EqualError(t, err, "tx is closed")

	tx = db.Tx()

	model := testEntityV1{ID: "firsttx", Num: 1, C2: "C1"}
	assert.NoError(t, tx.Save(model, keySave, nil))

	models := newTestListV1(0)
	assert.NoError(t, tx.Select(models, keySelect, "firsttx"))
	assert.NoError(t, tx.Close(false))

	assert.Equal(t, 1, len(models.items))
}
