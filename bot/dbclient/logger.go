package dbclient

import (
	"context"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogAdapter struct {
	logger *zap.Logger
}

var _ pgx.Logger = (*LogAdapter)(nil)

func NewLogAdapter(logger *zap.Logger) *LogAdapter {
	return &LogAdapter{
		logger: logger,
	}
}

func (l *LogAdapter) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	l.logger.Log(pgxLevelToZapLevel(level), msg, toZapFields(data)...)
}

func pgxLevelToZapLevel(level pgx.LogLevel) zapcore.Level {
	switch level {
	case pgx.LogLevelTrace:
		return zapcore.DebugLevel
	case pgx.LogLevelDebug:
		return zapcore.DebugLevel
	case pgx.LogLevelInfo:
		return zapcore.InfoLevel
	case pgx.LogLevelWarn:
		return zapcore.WarnLevel
	case pgx.LogLevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func toZapFields(data map[string]interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(data))
	for key, value := range data {
		fields = append(fields, zap.Any(key, value))
	}
	return fields
}
