package errors

import (
	"aur-update-checker/internal/logger"
	"time"
)

// ErrorHandlerFactory 错误处理器工厂
type ErrorHandlerFactory struct {
	logger logger.Logger
}

// NewErrorHandlerFactory 创建错误处理器工厂
func NewErrorHandlerFactory(logger logger.Logger) *ErrorHandlerFactory {
	return &ErrorHandlerFactory{
		logger: logger,
	}
}

// CreateDefaultErrorHandler 创建默认错误处理器
func (f *ErrorHandlerFactory) CreateDefaultErrorHandler() ErrorHandler {
	return NewDefaultErrorHandler(f.logger)
}

// CreateRetryHandler 创建重试处理器
func (f *ErrorHandlerFactory) CreateRetryHandler(config RetryConfig) *RetryHandler {
	errorHandler := f.CreateDefaultErrorHandler()
	retryHandler := NewRetryHandler(config, errorHandler, f.logger)
	return retryHandler.WithDefaultConditions()
}

// CreateRetryHandlerWithCustomConditions 创建带有自定义条件的重试处理器
func (f *ErrorHandlerFactory) CreateRetryHandlerWithCustomConditions(config RetryConfig, conditions []RetryCondition) *RetryHandler {
	errorHandler := f.CreateDefaultErrorHandler()
	retryHandler := NewRetryHandler(config, errorHandler, f.logger)

	// 添加默认条件
	retryHandler.WithDefaultConditions()

	// 添加自定义条件
	for _, condition := range conditions {
		retryHandler.AddCondition(condition)
	}

	return retryHandler
}

// CreateMiddleware 创建错误处理中间件
func (f *ErrorHandlerFactory) CreateMiddleware() *Middleware {
	errorHandler := f.CreateDefaultErrorHandler()
	return NewMiddleware(errorHandler, f.logger)
}

// 预定义的重试配置
var (
	// FastRetryConfig 快速重试配置，适用于临时性错误
	FastRetryConfig = RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Jitter:      true,
		Multiplier:  1.5,
	}

	// StandardRetryConfig 标准重试配置，适用于一般性错误
	StandardRetryConfig = RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   1 * time.Second,
		MaxDelay:    30 * time.Second,
		Jitter:      true,
		Multiplier:  2.0,
	}

	// AggressiveRetryConfig 激进重试配置，适用于重要但可能不稳定的操作
	AggressiveRetryConfig = RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   1 * time.Second,
		MaxDelay:    60 * time.Second,
		Jitter:      true,
		Multiplier:  2.0,
	}

	// NetworkRetryConfig 网络重试配置，适用于网络请求
	NetworkRetryConfig = RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   1 * time.Second,
		MaxDelay:    30 * time.Second,
		Jitter:      true,
		Multiplier:  2.0,
	}

	// LongRunningRetryConfig 长时间运行操作的重试配置
	LongRunningRetryConfig = RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   5 * time.Second,
		MaxDelay:    300 * time.Second,
		Jitter:      true,
		Multiplier:  1.5,
	}
)

// CreateFastRetryHandler 创建快速重试处理器
func (f *ErrorHandlerFactory) CreateFastRetryHandler() *RetryHandler {
	return f.CreateRetryHandler(FastRetryConfig)
}

// CreateStandardRetryHandler 创建标准重试处理器
func (f *ErrorHandlerFactory) CreateStandardRetryHandler() *RetryHandler {
	return f.CreateRetryHandler(StandardRetryConfig)
}

// CreateAggressiveRetryHandler 创建激进重试处理器
func (f *ErrorHandlerFactory) CreateAggressiveRetryHandler() *RetryHandler {
	return f.CreateRetryHandler(AggressiveRetryConfig)
}

// CreateNetworkRetryHandler 创建网络重试处理器
func (f *ErrorHandlerFactory) CreateNetworkRetryHandler() *RetryHandler {
	return f.CreateRetryHandler(NetworkRetryConfig)
}

// CreateLongRunningRetryHandler 创建长时间运行操作的重试处理器
func (f *ErrorHandlerFactory) CreateLongRunningRetryHandler() *RetryHandler {
	return f.CreateRetryHandler(LongRunningRetryConfig)
}
