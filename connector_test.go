package wpgx_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/shestakovda/wpgx"
	"github.com/stretchr/testify/assert"
)

func TestConnector(t *testing.T) {
	db, err := wpgx.Connect("port=test")
	assert.EqualError(t, err, "parsing connection string: strconv.ParseUint: parsing \"test\": invalid syntax")

	db, err = wpgx.Connect("postgresql://***")
	assert.EqualError(t, err, "creating connection pool: dial tcp: lookup ***: no such host")

	db, err = wpgx.Connect(connStr, wpgx.ReservePath("./connector.go"))
	assert.EqualError(t, err, "applying connection options: reserve path is not a directory")

	db, err = wpgx.Connect(connStr, wpgx.ReservePath(reserve))
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	defer func() {
		err = db.Deal(nil, `DROP TABLE users;`)
		assert.NoError(t, err)

		db.Close()

		_, err = db.Cook(`SELECT 1;`)
		assert.Equal(t, wpgx.ErrConnClosed, errors.Cause(err))

		err = db.Deal(nil, `SELECT 1;`)
		assert.Equal(t, wpgx.ErrConnClosed, errors.Cause(err))

		err = db.Load(nil, `SELECT 1;`)
		assert.Equal(t, wpgx.ErrConnClosed, errors.Cause(err))

		err = db.Save(nil, `SELECT 1;`, nil)
		assert.Equal(t, wpgx.ErrConnClosed, errors.Cause(err))
	}()

	_, err = db.Cook(`SELECT 1;`)
	assert.NoError(t, err)

	_, err = db.Cook(`SELECT FROM WHERE;`)
	assert.EqualError(t, err, "preparing statement: ERROR: syntax error at or near \"WHERE\" (SQLSTATE 42601)")

	err = db.Jail(false)
	assert.NoError(t, err)

	err = db.Deal(nil, `SELECT 1;`)
	assert.NoError(t, err)

	err = db.Deal(nil, `SELECT FROM WHERE;`)
	assert.EqualError(t, err, "executing query: ERROR: syntax error at or near \"WHERE\" (SQLSTATE 42601)")

	err = db.Load(nil, `SELECT FROM WHERE;`)
	assert.EqualError(t, err, "selecting data: ERROR: syntax error at or near \"WHERE\" (SQLSTATE 42601)")

	strings := make(wpgx.Strings, 0, 1)

	err = db.Deal(&strings, `SELECT FROM WHERE;`)
	assert.EqualError(t, err, "selecting data: ERROR: syntax error at or near \"WHERE\" (SQLSTATE 42601)")

	err = db.Deal(&strings, `SELECT 'test';`)
	assert.NoError(t, err)
	assert.Len(t, strings, 1)
	assert.Equal(t, "test", strings[0])

	err = db.Deal(nil, `CREATE TABLE users (id serial PRIMARY KEY, name text not null);`)
	assert.NoError(t, err)

	sqlInsert, err := db.Cook(`INSERT INTO users (name) VALUES ($1) RETURNING id;`, "name")
	assert.NoError(t, err)
	assert.NotEmpty(t, sqlInsert)

	strings = make(wpgx.Strings, 0, 1)
	u := &user{
		Name: "test",
	}
	err = db.Save(u, sqlInsert, &strings)
	assert.NoError(t, err)
	assert.Len(t, strings, 1)
	assert.Equal(t, "1", strings[0])

	nu := new(user)

	sqlSelect, err := db.Cook(`SELECT * FROM users WHERE id = $1;`)
	assert.NoError(t, err)
	assert.NotEmpty(t, sqlSelect)

	err = db.Load(nu, sqlSelect, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, nu.ID)
	assert.Equal(t, "test", nu.Name)

	err = db.Save(new(user), sqlInsert, nil)
	assert.EqualError(t, err, "executing query: ERROR: null value in column \"name\" violates not-null constraint (SQLSTATE 23502)")

	dump, err := ioutil.ReadFile(filepath.Join(reserve, sqlInsert+".pgsql"))
	assert.NoError(t, err)
	assert.Equal(t, "INSERT INTO users (name) VALUES ($1) RETURNING id;", string(dump))

	dump, err = ioutil.ReadFile(filepath.Join(reserve, sqlInsert+"_c3b374cf5a76aa557c625f9eaa6e91f462b67fe2.json"))
	assert.NoError(t, err)
	assert.Equal(t, "{\n  \"name\": null\n}", string(dump))

}
