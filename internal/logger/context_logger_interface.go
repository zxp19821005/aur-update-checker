package logger

import (
	"context"
)

// ContextLogger 上下文感知的日志记录器接口
type ContextLogger interface {
	Debug(ctx context.Context, msg string)
	Info(ctx context.Context, msg string)
	Warn(ctx context.Context, msg string)
	Error(ctx context.Context, msg string)
	Fatal(ctx context.Context, msg string)
	Debugf(ctx context.Context, format string, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Fatalf(ctx context.Context, format string, args ...interface{})
	WithFields(ctx context.Context, fields map[string]interface{}) context.Context
	WithError(ctx context.Context, err error) context.Context
}
