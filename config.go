package wpgx

import "os"
import "path/filepath"
import "github.com/jackc/pgx"
import "github.com/pkg/errors"

type Config struct {
	pgx.ConnPoolConfig
	ReservePath string
}

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

func LogLevel(lvl int) func(*Config) error {
	return func(cfg *Config) error {
		cfg.ConnPoolConfig.ConnConfig.Logger = new(logger)
		cfg.ConnPoolConfig.ConnConfig.LogLevel = lvl
		return nil
	}
}

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
