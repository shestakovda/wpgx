package wpgx_test

import "testing"
import "github.com/shestakovda/wpgx"
import "github.com/stretchr/testify/assert"

func TestRawList(t *testing.T) {
	db, err := wpgx.Connect(connStr)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	key, err := db.Prepare(`
SELECT 
    'name: ' || t::text AS name,
    'text: ' || t::text AS text
FROM generate_series('2018-01-01'::timestamp, '2019-01-01', '1 day') AS t;
    `, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, key)

	list := make(wpgx.RawList, 0, 366)
	assert.Len(t, list, 0)

	err = db.Deal(&list, key)
	assert.NoError(t, err)
	assert.Len(t, list, 366)

	assert.Equal(t, map[string]string{
		"name": "name: 2018-02-01 00:00:00",
		"text": "text: 2018-02-01 00:00:00",
	}, list[31])

	err = list.Collect(nil)
	assert.Equal(t, wpgx.ErrUnknownType, err)
}
