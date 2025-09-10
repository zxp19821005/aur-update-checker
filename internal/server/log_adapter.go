package server

import "aur-update-checker/internal/logger"

// LogProviderAdapter 适配器，将logger.Logger转换为interfaces.LoggerProvider
type LogProviderAdapter struct {
	logger *logger.Logger
}

// 实现interfaces.LoggerProvider接口
func (a *LogProviderAdapter) Info(msg string) {
	a.logger.Info(msg)
}

func (a *LogProviderAdapter) Debug(msg string) {
	a.logger.Debug(msg)
}

func (a *LogProviderAdapter) Warn(msg string) {
	a.logger.Warn(msg)
}

func (a *LogProviderAdapter) Error(msg string) {
	a.logger.Error(msg)
}

func (a *LogProviderAdapter) Infof(format string, args ...interface{}) {
	a.logger.Infof(format, args...)
}

func (a *LogProviderAdapter) Debugf(format string, args ...interface{}) {
	a.logger.Debugf(format, args...)
}

func (a *LogProviderAdapter) Warnf(format string, args ...interface{}) {
	a.logger.Warnf(format, args...)
}

func (a *LogProviderAdapter) Errorf(format string, args ...interface{}) {
	a.logger.Errorf(format, args...)
}
