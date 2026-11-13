package config

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/sirupsen/logrus"
)

type pgxLogrus struct {
	log *logrus.Logger
}

func (l pgxLogrus) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	entry := l.log.WithContext(ctx).WithFields(logrus.Fields(data))

	switch level {
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		entry.Debug(msg)
	case tracelog.LogLevelInfo:
		entry.Info(msg)
	case tracelog.LogLevelWarn:
		entry.Warn(msg)
	case tracelog.LogLevelError:
		entry.Error(msg)
	}
}

func NewDB(config *ConfigEnv, log *logrus.Logger) *sql.DB {
	connConfig, err := pgx.ParseConfig(config.DB_DSN)
	if err != nil {
		panic(err)
	}

	if log.IsLevelEnabled(logrus.DebugLevel) {
		connConfig.Tracer = &tracelog.TraceLog{
			Logger:   pgxLogrus{log: log},
			LogLevel: tracelog.LogLevelDebug,
		}
	}

	db := stdlib.OpenDB(*connConfig)

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(60 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		panic(err)
	}

	return db
}
