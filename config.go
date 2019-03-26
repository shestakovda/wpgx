package wpgx

import (
	"os"
	"path/filepath"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
)

// Config is just a pgx.ConnPoolConfig with some extra options
type Config struct {
	ReservePath string
	pgx.ConnPoolConfig
}

// PoolSize is a config helper to set pgx.ConnPoolConfig.MaxConnections field
func PoolSize(size int) func(*Config) error {
	return func(cfg *Config) error {
		// from pgx: max simultaneous connections to use, must be at least 2
		if size < 2 {
			size = 2
		}
		cfg.ConnPoolConfig.MaxConnections = size
		return nil
	}
}

// LogLevel is a config helper for use glog as logger
// godoc: https://godoc.org/github.com/golang/glog
func LogLevel(lvl int) func(*Config) error {
	return func(cfg *Config) error {
		if lvl < 0 {
			lvl = pgx.LogLevelNone
		}
		cfg.ConnPoolConfig.ConnConfig.Logger = new(logger)
		cfg.ConnPoolConfig.ConnConfig.LogLevel = pgx.LogLevel(lvl)
		return nil
	}
}

// ReservePath is a config helper to set catalog for saving
// prepared sql files and uncommitted object data as json files
func ReservePath(possible string) func(*Config) error {
	return func(cfg *Config) (err error) {
		if possible == "" {
			return
		}

		var path string

		if path, err = filepath.Abs(possible); err != nil {
			return errors.Wrap(err, "checking reserve path")
		}

		var info os.FileInfo

		if info, err = os.Stat(path); err != nil {
			return errors.Wrap(err, "testing reserve path")
		}

		if !info.IsDir() {
			return errors.New("reserve path is not a directory")
		}

		cfg.ReservePath = path
		return
	}
}
