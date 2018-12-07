package wpgx_test

import (
	"path/filepath"
	"testing"

	"github.com/jackc/pgx"
	"github.com/shestakovda/wpgx"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg := new(wpgx.Config)

	assert.NoError(t, wpgx.PoolSize(0)(cfg))
	assert.Equal(t, 2, cfg.ConnPoolConfig.MaxConnections)

	assert.NoError(t, wpgx.PoolSize(123)(cfg))
	assert.Equal(t, 123, cfg.ConnPoolConfig.MaxConnections)

	assert.NoError(t, wpgx.LogLevel(-1)(cfg))
	assert.Equal(t, pgx.LogLevelNone, cfg.ConnPoolConfig.ConnConfig.LogLevel)

	assert.NoError(t, wpgx.LogLevel(pgx.LogLevelInfo)(cfg))
	assert.Equal(t, pgx.LogLevelInfo, cfg.ConnPoolConfig.ConnConfig.LogLevel)

	assert.NoError(t, wpgx.ReservePath("")(cfg))
	assert.Equal(t, "", cfg.ReservePath)

	abs, err := filepath.Abs(reserve)
	assert.NoError(t, err)
	assert.NoError(t, wpgx.ReservePath(reserve)(cfg))
	assert.Equal(t, abs, cfg.ReservePath)

	abs, err = filepath.Abs("./test.test")
	assert.NoError(t, err)
	err = wpgx.ReservePath("./test.test")(cfg)
	assert.EqualError(t, err, "testing reserve path: stat "+abs+": no such file or directory")

	err = wpgx.ReservePath("./config_test.go")(cfg)
	assert.EqualError(t, err, "reserve path is not a directory")
}
