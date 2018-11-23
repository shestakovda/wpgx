package wpgx

import "testing"
import "github.com/jackc/pgx"

func TestLogger(t *testing.T) {
	const msg = "test %+v"
	var data = map[string]interface{}{"key": "value"}

	l := new(logger)
	l.Log(pgx.LogLevelNone, msg, data)
	l.Log(pgx.LogLevelError, msg, data)
	l.Log(pgx.LogLevelWarn, msg, data)
	l.Log(pgx.LogLevelInfo, msg, data)
	l.Log(pgx.LogLevelDebug, msg, data)
}
