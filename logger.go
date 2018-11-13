package wpgx

import "github.com/jackc/pgx"
import "github.com/golang/glog"

const logtpl = "wpgx: %s %+v"

type logger struct{}

func (l *logger) Log(level pgx.LogLevel, msg string, data map[string]interface{}) {
	switch level {
	case pgx.LogLevelNone:
		return
	case pgx.LogLevelError:
		glog.Errorf(logtpl, msg, data)
	case pgx.LogLevelWarn:
		glog.Warningf(logtpl, msg, data)
	default:
		glog.Infof(logtpl, msg, data)
	}
}
