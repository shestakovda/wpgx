package wpgx_test

import "testing"
import "github.com/shestakovda/wpgx"
import "github.com/stretchr/testify/assert"

func TestStrings(t *testing.T) {
	const sqlName = "select_strings"

	db, err := wpgx.Connect(connStr)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	key, err := db.Prepare(`
SELECT 
    t::text 
FROM generate_series('2018-01-01'::timestamp, '2019-01-01', '1 day') AS t;
    `, nil)
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
