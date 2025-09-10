package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

// CustomFormatter 自定义日志格式化器
type CustomFormatter struct {
	TimestampFormat string
	DisableColors   bool
}

// Format 格式化日志记录
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 格式化时间戳
	timestamp := entry.Time.Format(f.TimestampFormat)

	// 格式化日志级别
	level := strings.ToUpper(entry.Level.String())

	// 构建消息，确保以换行符结尾
	message := entry.Message
	if !strings.HasSuffix(message, "\n") {
		message = message + "\n"
	}

	// 构建最终的日志格式
	formatted := fmt.Sprintf("[%s] %s %s", level, timestamp, message)

	return []byte(formatted), nil
}

// EnsureAppDataDir 确保应用数据目录存在并返回目录路径
func EnsureAppDataDir() (string, error) {
	var appDataDir string

	// Linux系统: ~/.config或$XDG_CONFIG_HOME
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		appDataDir = xdgConfig
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("无法获取用户主目录: %v", err)
		}
		appDataDir = filepath.Join(home, ".config")
	}

	// 创建应用特定的目录
	appDir := filepath.Join(appDataDir, "aur-update-checker")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", fmt.Errorf("创建应用目录失败: %v", err)
	}

	return appDir, nil
}

// GlobalLogger 全局日志实例
var GlobalLogger *Logger

// Logger 日志结构体
type Logger struct {
	*logrus.Logger
	environment string // 运行环境: development, production, testing
	enableColor  bool   // 是否启用彩色输出
}

// InitLogger 初始化日志系统
func InitLogger() *Logger {
	// 创建日志实例
	log := logrus.New()

	// 获取运行环境，默认为development
	env := strings.ToLower(os.Getenv("AUR_ENV"))
	if env == "" {
		env = "development"
	}

	// 根据环境设置日志级别
	var logLevel logrus.Level
	switch env {
	case "production":
		logLevel = logrus.ErrorLevel // 修改为ErrorLevel，确保错误日志能正确显示为ERROR级别
	case "testing":
		logLevel = logrus.WarnLevel
	default: // development
		logLevel = logrus.DebugLevel
	}

	// 设置日志级别
	log.SetLevel(logLevel)

	// 获取是否启用彩色输出，默认启用
	enableColor := strings.ToLower(os.Getenv("AUR_LOG_COLOR")) != "false"

	// 根据环境设置日志格式
	if env == "production" {
		// 生产环境使用JSON格式，便于日志分析
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		// 开发和测试环境使用自定义文本格式，便于阅读
		log.SetFormatter(&CustomFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			DisableColors:   !enableColor,
		})
	}

	// 获取应用数据目录
	appDir, err := EnsureAppDataDir()
	if err != nil {
		fmt.Printf("获取应用数据目录失败: %v", err)
		os.Exit(1)
	}

	// 日志文件路径
	logPath := filepath.Join(appDir, "logs")
	if err := os.MkdirAll(logPath, 0755); err != nil {
		fmt.Printf("创建日志目录失败: %v", err)
		os.Exit(1)
	}

	// 按日期创建日志文件
	currentTime := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logPath, fmt.Sprintf("aur-update-checker-%s.log", currentTime))

	// 打开日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("打开日志文件失败: %v", err)
		os.Exit(1)
	}

	// 设置日志输出
	if env == "production" {
		// 生产环境只输出到文件
		log.SetOutput(file)
	} else {
		// 开发和测试环境只输出到文件，因为我们会通过 LogWithColor 方法输出到控制台
		log.SetOutput(file)
	}

	// 创建Logger实例
	logger := &Logger{
		Logger:      log,
		environment: env,
		enableColor: enableColor,
	}

	// 设置全局日志实例
	GlobalLogger = logger

	// 记录初始化信息
	logger.WithFields(logrus.Fields{
		"environment": env,
		"log_level":   logLevel.String(),
		"color":       enableColor,
	}).Info("日志系统初始化完成")

	return logger
}

// LogWithColor 带颜色的日志输出
func (l *Logger) LogWithColor(level logrus.Level, msg string) {
	// 如果是生产环境，不输出彩色日志到控制台
	if l.environment == "production" {
		return
	}

	// 确保消息以换行符结尾
	if !strings.HasSuffix(msg, "\n") {
		msg = msg + "\n"
	}

	// 格式化时间戳
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 根据级别设置颜色并输出
	switch level {
	case logrus.DebugLevel:
		color.New(color.FgCyan).Printf("[DEBUG] %s %s", timestamp, msg)
	case logrus.InfoLevel:
		color.New(color.FgGreen).Printf("[INFO] %s %s", timestamp, msg)
	case logrus.WarnLevel:
		color.New(color.FgYellow).Printf("[WARN] %s %s", timestamp, msg)
	case logrus.ErrorLevel:
		color.New(color.FgRed).Printf("[ERROR] %s %s", timestamp, msg)
	case logrus.FatalLevel:
		color.New(color.FgMagenta).Printf("[FATAL] %s %s", timestamp, msg)
	case logrus.PanicLevel:
		color.New(color.FgHiRed).Printf("[PANIC] %s %s", timestamp, msg)
	default:
		fmt.Printf("[UNKNOWN] %s %s", timestamp, msg)
	}
}

// WithFields 添加结构化字段
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// WithError 添加错误信息
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// Debug 带颜色的调试日志
func (l *Logger) Debug(args ...interface{}) {
	msg := fmt.Sprint(args...)
	// 开发和测试环境输出到控制台
	if l.environment != "production" {
		l.LogWithColor(logrus.DebugLevel, msg)
	}
	// 所有环境都输出到文件
	l.Logger.Debug(msg)
}

// Debugf 带格式的调试日志
func (l *Logger) Debugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// 开发和测试环境输出到控制台
	if l.environment != "production" {
		l.LogWithColor(logrus.DebugLevel, msg)
	}
	// 所有环境都输出到文件
	l.Logger.Debug(msg)
}

// Info 带颜色的信息日志
func (l *Logger) Info(args ...interface{}) {
	msg := fmt.Sprint(args...)
	// 开发和测试环境输出到控制台
	if l.environment != "production" {
		l.LogWithColor(logrus.InfoLevel, msg)
	}
	// 所有环境都输出到文件
	l.Logger.Info(msg)
}

// Infof 带格式的信息日志
func (l *Logger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// 开发和测试环境输出到控制台
	if l.environment != "production" {
		l.LogWithColor(logrus.InfoLevel, msg)
	}
	// 所有环境都输出到文件
	l.Logger.Info(msg)
}

// Warn 带颜色的警告日志
func (l *Logger) Warn(args ...interface{}) {
	msg := fmt.Sprint(args...)
	// 开发和测试环境输出到控制台
	if l.environment != "production" {
		l.LogWithColor(logrus.WarnLevel, msg)
	}
	// 所有环境都输出到文件
	l.Logger.Warn(msg)
}

// Warnf 带格式的警告日志
func (l *Logger) Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// 开发和测试环境输出到控制台
	if l.environment != "production" {
		l.LogWithColor(logrus.WarnLevel, msg)
	}
	// 所有环境都输出到文件
	l.Logger.Warn(msg)
}

// Error 带颜色的错误日志
func (l *Logger) Error(args ...interface{}) {
	msg := fmt.Sprint(args...)
	// 开发和测试环境输出到控制台
	if l.environment != "production" {
		l.LogWithColor(logrus.ErrorLevel, msg)
	}
	// 所有环境都输出到文件
	l.Logger.Error(msg)
}

// Errorf 带格式的错误日志
func (l *Logger) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// 开发和测试环境输出到控制台
	if l.environment != "production" {
		l.LogWithColor(logrus.ErrorLevel, msg)
	}
	// 所有环境都输出到文件
	l.Logger.Error(msg)
}

// Fatal 带颜色的致命错误日志
func (l *Logger) Fatal(args ...interface{}) {
	msg := fmt.Sprint(args...)
	// 开发和测试环境输出到控制台
	if l.environment != "production" {
		l.LogWithColor(logrus.FatalLevel, msg)
	}
	// 所有环境都输出到文件
	l.Logger.Fatal(msg)
	os.Exit(1)
}

// Fatalf 带格式的致命错误日志
func (l *Logger) Fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// 开发和测试环境输出到控制台
	if l.environment != "production" {
		l.LogWithColor(logrus.FatalLevel, msg)
	}
	// 所有环境都输出到文件
	l.Logger.Fatal(msg)
	os.Exit(1)
}

// Panic 带颜色的恐慌日志
func (l *Logger) Panic(args ...interface{}) {
	msg := fmt.Sprint(args...)
	// 开发和测试环境输出到控制台
	if l.environment != "production" {
		l.LogWithColor(logrus.PanicLevel, msg)
	}
	// 所有环境都输出到文件
	l.Logger.Panic(msg)
	panic(msg)
}

// Panicf 带格式的恐慌日志
func (l *Logger) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// 开发和测试环境输出到控制台
	if l.environment != "production" {
		l.LogWithColor(logrus.PanicLevel, msg)
	}
	// 所有环境都输出到文件
	l.Logger.Panic(msg)
	panic(msg)
}

// log 记录日志条目
func (l *Logger) log(entry LogEntry) {
	// 创建带有字段的日志条目
	logEntry := l.WithFields(logrus.Fields(entry.Fields))

	// 如果有错误，添加错误信息
	if entry.Error != nil {
		logEntry = logEntry.WithError(entry.Error)
	}

	// 添加文件和行号信息
	logEntry = logEntry.WithFields(logrus.Fields{
		"file": entry.File,
		"line": entry.Line,
	})

	// 根据日志级别记录日志
	switch entry.Level {
	case logrus.DebugLevel:
		logEntry.Debug(entry.Message)
	case logrus.InfoLevel:
		logEntry.Info(entry.Message)
	case logrus.WarnLevel:
		logEntry.Warn(entry.Message)
	case logrus.ErrorLevel:
		logEntry.Error(entry.Message)
	case logrus.FatalLevel:
		logEntry.Fatal(entry.Message)
	case logrus.PanicLevel:
		logEntry.Panic(entry.Message)
	default:
		logEntry.Info(entry.Message)
	}
}
