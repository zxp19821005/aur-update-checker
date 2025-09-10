package errors

import (
	"context"
	"math"
	"math/rand"
	"time"
	"aur-update-checker/internal/logger"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts int           // 最大尝试次数
	BaseDelay   time.Duration // 基础延迟时间
	MaxDelay    time.Duration // 最大延迟时间
	Jitter      bool          // 是否添加随机抖动
	Multiplier  float64       // 延迟乘数
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	BaseDelay:   1 * time.Second,
	MaxDelay:    30 * time.Second,
	Jitter:      true,
	Multiplier:  2.0,
}

// RetryCondition 重试条件函数类型
type RetryCondition func(error) bool

// RetryHandler 重试处理器
type RetryHandler struct {
	config        RetryConfig
	conditions    []RetryCondition
	errorHandler  ErrorHandler
	logger        logger.Logger
}

// NewRetryHandler 创建重试处理器
func NewRetryHandler(config RetryConfig, errorHandler ErrorHandler, logger logger.Logger) *RetryHandler {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = DefaultRetryConfig.MaxAttempts
	}
	if config.BaseDelay <= 0 {
		config.BaseDelay = DefaultRetryConfig.BaseDelay
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = DefaultRetryConfig.MaxDelay
	}
	if config.Multiplier <= 0 {
		config.Multiplier = DefaultRetryConfig.Multiplier
	}

	return &RetryHandler{
		config:       config,
		conditions:   make([]RetryCondition, 0),
		errorHandler: errorHandler,
		logger:       logger,
	}
}

// AddCondition 添加重试条件
func (r *RetryHandler) AddCondition(condition RetryCondition) *RetryHandler {
	r.conditions = append(r.conditions, condition)
	return r
}

// WithDefaultConditions 添加默认的重试条件
func (r *RetryHandler) WithDefaultConditions() *RetryHandler {
	// 网络错误可重试
	r.AddCondition(func(err error) bool {
		return IsNetworkError(err)
	})

	// 超时错误可重试
	r.AddCondition(func(err error) bool {
		return IsTimeoutError(err)
	})

	// 速率限制错误可重试
	r.AddCondition(func(err error) bool {
		if appErr, ok := err.(*AppError); ok {
			return appErr.Code == RateLimitError
		}
		return false
	})

	// 临时性HTTP错误可重试
	r.AddCondition(func(err error) bool {
		if appErr, ok := err.(*AppError); ok {
			return appErr.Code == HTTPError && isTemporaryHTTPError(appErr.OriginalErr)
		}
		return false
	})

	return r
}

// isTemporaryHTTPError 检查是否为临时性HTTP错误
func isTemporaryHTTPError(err error) bool {
	// 处理 nil 情况
	if err == nil {
		return false
	}

	// 这里可以根据实际需求判断哪些HTTP错误是临时性的
	// 例如：5xx错误通常是服务器端临时性问题
	return true
}

// Execute 执行带有重试的操作
func (r *RetryHandler) Execute(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		// 检查上下文是否已取消
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 执行操作
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// 处理错误
		handledErr := r.errorHandler.HandleError(err)
		if handledErr != err {
			lastErr = handledErr
		}

		// 检查是否应该重试
		if !r.shouldRetry(handledErr) || attempt == r.config.MaxAttempts {
			break
		}

		// 计算延迟时间
		delay := r.calculateDelay(attempt)

		// 记录重试日志
		r.logger.Infof("第%d次尝试失败，%v后重试: %v", attempt, delay, handledErr)

		// 等待延迟时间
		select {
		case <-time.After(delay):
			// 继续下一次尝试
		case <-ctx.Done():
			// 上下文被取消
			return ctx.Err()
		}
	}

	return lastErr
}

// shouldRetry 检查是否应该重试
func (r *RetryHandler) shouldRetry(err error) bool {
	// 检查错误是否可重试
	if appErr, ok := err.(*AppError); ok {
		if !appErr.Retryable {
			return false
		}
	}

	// 检查自定义条件
	for _, condition := range r.conditions {
		if condition(err) {
			return true
		}
	}

	return false
}

// calculateDelay 计算延迟时间
func (r *RetryHandler) calculateDelay(attempt int) time.Duration {
	// 指数退避算法
	delay := float64(r.config.BaseDelay) * math.Pow(r.config.Multiplier, float64(attempt-1))

	// 限制最大延迟时间
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	// 添加随机抖动
	if r.config.Jitter {
		// 添加0.8到1.2之间的随机抖动
		jitter := 0.8 + rand.Float64()*0.4
		delay *= jitter
	}

	return time.Duration(delay)
}

// ExecuteWithResult 执行带有重试的操作，并返回结果
func (r *RetryHandler) ExecuteWithResult(ctx context.Context, operation func() (interface{}, error)) (interface{}, error) {
	var lastErr error
	var result interface{}

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		// 检查上下文是否已取消
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// 执行操作
		res, err := operation()
		if err == nil {
			return res, nil
		}

		lastErr = err
		result = res

		// 处理错误
		handledErr := r.errorHandler.HandleError(err)
		if handledErr != err {
			lastErr = handledErr
		}

		// 检查是否应该重试
		if !r.shouldRetry(handledErr) || attempt == r.config.MaxAttempts {
			break
		}

		// 计算延迟时间
		delay := r.calculateDelay(attempt)

		// 记录重试日志
		r.logger.Infof("第%d次尝试失败，%v后重试: %v", attempt, delay, handledErr)

		// 等待延迟时间
		select {
		case <-time.After(delay):
			// 继续下一次尝试
		case <-ctx.Done():
			// 上下文被取消
			return nil, ctx.Err()
		}
	}

	return result, lastErr
}

// IsNetworkError 检查是否为网络错误
func IsNetworkError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == NetworkError || appErr.Code == NetworkTimeoutError
	}
	return false
}

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == TimeoutError || appErr.Code == NetworkTimeoutError
	}
	return false
}
