package wpgx_test

import (
	"testing"

	"github.com/shestakovda/wpgx"
	"github.com/stretchr/testify/assert"
)

func TestInts(t *testing.T) {
	db, err := wpgx.Connect(connStr)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()

	key, err := db.Cook(`
SELECT 
    extract(epoch from t)::int
FROM generate_series('2018-01-01'::timestamp, '2019-01-01', '1 day') AS t;
    `)
	assert.NoError(t, err)
	assert.NotEmpty(t, key)

	ints := make(wpgx.Ints, 0, 366)
	assert.Len(t, ints, 0)

	err = db.Deal(&ints, key)
	assert.NoError(t, err)
	assert.Len(t, ints, 366)

	assert.Equal(t, 1517443200, ints[31])

	err = ints.Collect(nil)
	assert.Equal(t, wpgx.ErrUnknownType, err)
}
