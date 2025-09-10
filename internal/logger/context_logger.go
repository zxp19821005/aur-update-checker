package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// LogLevel 日志级别类型
type LogLevel = logrus.Level

// 日志级别常量
const (
	DebugLevel = logrus.DebugLevel
	InfoLevel  = logrus.InfoLevel
	WarnLevel  = logrus.WarnLevel
	ErrorLevel = logrus.ErrorLevel
	FatalLevel = logrus.FatalLevel
)

// LogEntry 日志条目
type LogEntry struct {
	Time    time.Time
	Level   LogLevel
	Message string
	File    string
	Line    int
	Fields  map[string]interface{}
	Error   error
}

// contextLogger 上下文日志记录器实现
type contextLogger struct {
	baseLogger *Logger
	fields    map[string]interface{}
	error     error
}

// NewContextLogger 创建新的上下文日志记录器
func NewContextLogger(baseLogger *Logger) ContextLogger {
	return &contextLogger{
		baseLogger: baseLogger,
		fields:    make(map[string]interface{}),
	}
}

// Debug 记录调试级别日志
func (l *contextLogger) Debug(ctx context.Context, msg string) {
	l.log(ctx, DebugLevel, msg)
}

// Info 记录信息级别日志
func (l *contextLogger) Info(ctx context.Context, msg string) {
	l.log(ctx, InfoLevel, msg)
}

// Warn 记录警告级别日志
func (l *contextLogger) Warn(ctx context.Context, msg string) {
	l.log(ctx, WarnLevel, msg)
}

// Error 记录错误级别日志
func (l *contextLogger) Error(ctx context.Context, msg string) {
	l.log(ctx, ErrorLevel, msg)
}

// Fatal 记录致命错误级别日志
func (l *contextLogger) Fatal(ctx context.Context, msg string) {
	l.log(ctx, FatalLevel, msg)
	os.Exit(1)
}

// Debugf 记录格式化调试级别日志
func (l *contextLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.logf(ctx, DebugLevel, format, args...)
}

// Infof 记录格式化信息级别日志
func (l *contextLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.logf(ctx, InfoLevel, format, args...)
}

// Warnf 记录格式化警告级别日志
func (l *contextLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.logf(ctx, WarnLevel, format, args...)
}

// Errorf 记录格式化错误级别日志
func (l *contextLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.logf(ctx, ErrorLevel, format, args...)
}

// Fatalf 记录格式化致命错误级别日志
func (l *contextLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	l.logf(ctx, FatalLevel, format, args...)
	os.Exit(1)
}

// WithFields 添加字段到上下文
func (l *contextLogger) WithFields(ctx context.Context, fields map[string]interface{}) context.Context {
	// 获取或创建上下文中的日志记录器
	logger := l.getLoggerFromContext(ctx)

	// 合并字段
	for k, v := range fields {
		logger.fields[k] = v
	}

	// 将更新后的日志记录器存回上下文
	return context.WithValue(ctx, loggerContextKey, logger)
}

// WithError 添加错误到上下文
func (l *contextLogger) WithError(ctx context.Context, err error) context.Context {
	// 获取或创建上下文中的日志记录器
	logger := l.getLoggerFromContext(ctx)

	// 设置错误
	logger.error = err

	// 将更新后的日志记录器存回上下文
	return context.WithValue(ctx, loggerContextKey, logger)
}

// getLoggerFromContext 从上下文中获取日志记录器，如果不存在则创建新的
func (l *contextLogger) getLoggerFromContext(ctx context.Context) *contextLogger {
	if logger, ok := ctx.Value(loggerContextKey).(*contextLogger); ok {
		return logger
	}

	// 创建新的日志记录器并复制当前字段
	newLogger := &contextLogger{
		baseLogger: l.baseLogger,
		fields:     make(map[string]interface{}),
		error:      l.error,
	}

	// 复制字段
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// log 记录日志
func (l *contextLogger) log(ctx context.Context, level LogLevel, msg string) {
	// 获取调用者信息
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "unknown"
		line = 0
	}

	// 构建日志条目
	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
		File:    filepath.Base(file),
		Line:    line,
		Fields:  l.fields,
		Error:   l.error,
	}

	// 获取请求ID（如果存在）
	if requestID := ctx.Value("requestID"); requestID != nil {
		if entry.Fields == nil {
			entry.Fields = make(map[string]interface{})
		}
		entry.Fields["requestID"] = requestID
	}

	// 获取用户ID（如果存在）
	if userID := ctx.Value("userID"); userID != nil {
		if entry.Fields == nil {
			entry.Fields = make(map[string]interface{})
		}
		entry.Fields["userID"] = userID
	}

	// 记录日志
	l.baseLogger.log(entry)
}

// logf 记录格式化日志
func (l *contextLogger) logf(ctx context.Context, level LogLevel, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.log(ctx, level, msg)
}

// 上下文键类型
type contextKey int

// 日志上下文键
const loggerContextKey contextKey = iota

// GetLoggerFromContext 从上下文中获取日志记录器
func GetLoggerFromContext(ctx context.Context) ContextLogger {
	// 如果上下文中没有日志记录器，则创建一个新的
	if logger, ok := ctx.Value(loggerContextKey).(ContextLogger); ok {
		return logger
	}
	return NewContextLogger(GlobalLogger)
}

// WithLogger 将日志记录器添加到上下文
func WithLogger(ctx context.Context, logger ContextLogger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

// WithRequestID 将请求ID添加到上下文并创建新的日志记录器
func WithRequestID(ctx context.Context, requestID string) context.Context {
	logger := GetLoggerFromContext(ctx)
	return logger.WithFields(ctx, map[string]interface{}{"requestID": requestID})
}

// WithUserID 将用户ID添加到上下文并创建新的日志记录器
func WithUserID(ctx context.Context, userID string) context.Context {
	logger := GetLoggerFromContext(ctx)
	return logger.WithFields(ctx, map[string]interface{}{"userID": userID})
}
