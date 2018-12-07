package wpgx_test

import (
	"testing"

	"github.com/shestakovda/wpgx"
	"github.com/stretchr/testify/assert"
)

func TestStrings(t *testing.T) {
	db, err := wpgx.Connect(connStr)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	key, err := db.Cook(`
SELECT 
    t::text 
FROM generate_series('2018-01-01'::timestamp, '2019-01-01', '1 day') AS t;
    `)
	assert.NoError(t, err)
	assert.NotEmpty(t, key)

	strings := make(wpgx.Strings, 0, 366)
	assert.Len(t, strings, 0)

	err = db.Deal(&strings, key)
	assert.NoError(t, err)
	assert.Len(t, strings, 366)

	assert.Equal(t, "2018-02-01 00:00:00", strings[31])

	err = strings.Collect(nil)
	assert.Equal(t, wpgx.ErrUnknownType, err)
}
