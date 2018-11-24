package wpgx_test

import (
	"database/sql"
	"testing"

	"github.com/pkg/errors"
	"github.com/shestakovda/wpgx"
	"github.com/stretchr/testify/assert"
)

func TestDealer(t *testing.T) {
	db, err := wpgx.Connect(connStr)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	d, err := db.NewDealer()
	assert.NoError(t, err)
	assert.NotNil(t, d)
	defer func() {
		err = d.Jail(false)
		assert.NoError(t, err)

		err = d.Deal(nil, `SELECT 1;`)
		assert.Equal(t, wpgx.ErrConnClosed, errors.Cause(err))
	}()

	err = d.Deal(nil, `SELECT 1;`)
	assert.NoError(t, err)

	err = d.Deal(nil, `SELECT FROM WHERE;`)
	assert.EqualError(t, err, "executing query: ERROR: syntax error at or near \"WHERE\" (SQLSTATE 42601)")

	strings := make(wpgx.Strings, 0, 1)

	err = d.Deal(&strings, `SELECT FROM WHERE;`)
	assert.EqualError(t, err, "selecting data: ERROR: syntax error at or near \"WHERE\" (SQLSTATE 42601)")
	d.Jail(false)

	d, err = db.NewDealer()
	assert.NoError(t, err)
	assert.NotNil(t, d)

	err = d.Deal(&strings, `SELECT 'test';`)
	assert.NoError(t, err)
	assert.Len(t, strings, 1)
	assert.Equal(t, "test", strings[0])

	err = d.Deal(nil, `CREATE TABLE users (id serial PRIMARY KEY, name text);`)
	assert.NoError(t, err)

	sqlInsert, err := d.Prepare(`INSERT INTO users (name) VALUES ($1) RETURNING id;`, "name")
	assert.NoError(t, err)
	assert.NotEmpty(t, sqlInsert)

	strings = make(wpgx.Strings, 0, 1)
	u := &user{
		Name: "test",
	}
	err = d.Save(u, sqlInsert, &strings)
	assert.NoError(t, err)
	assert.Len(t, strings, 1)
	assert.Equal(t, "1", strings[0])

	nu := new(user)

	sqlSelect, err := d.Prepare(`SELECT * FROM users WHERE id = $1;`)
	assert.NoError(t, err)
	assert.NotEmpty(t, sqlSelect)

	err = d.Load(nu, sqlSelect, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, nu.ID)
	assert.Equal(t, "test", nu.Name)

}

type user struct {
	ID   int
	Name string
}

func (u *user) Extrude() wpgx.Translator {
	return &userModel{
		ID:   u.ID,
		Name: sql.NullString{Valid: u.Name != "", String: u.Name},
	}
}

func (u *user) Receive(item wpgx.Translator) error {
	model, ok := item.(*userModel)
	if !ok {
		return wpgx.ErrUnknownType
	}

	u.ID = model.ID
	if model.Name.Valid {
		u.Name = model.Name.String
	} else {
		u.Name = ""
	}
	return nil
}

type userModel struct {
	ID   int
	Name sql.NullString
}

func (m *userModel) Translate(name string) interface{} {
	switch name {
	case "id":
		return &m.ID
	case "name":
		return &m.Name
	}
	return nil
}
