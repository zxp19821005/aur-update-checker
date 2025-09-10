package errors

import (
	"context"
	"net/http"
	"time"
	"aur-update-checker/internal/logger"
)

// Middleware 错误处理中间件
type Middleware struct {
	errorHandler ErrorHandler
	logger       logger.Logger
}

// NewMiddleware 创建错误处理中间件
func NewMiddleware(errorHandler ErrorHandler, logger logger.Logger) *Middleware {
	return &Middleware{
		errorHandler: errorHandler,
		logger:       logger,
	}
}

// HTTPMiddleware HTTP错误处理中间件
func (m *Middleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.logger.Errorf("HTTP请求处理中发生panic: %v", err)

				// 将panic转换为AppError
				appErr := NewSystemError("HTTP请求处理中发生panic", err.(error))

				// 处理错误
				handledErr := m.errorHandler.HandleError(appErr)

				// 返回错误响应
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"` + handledErr.Error() + `"}`))
			}
		}()

		// 创建带有错误上下文的请求上下文
		ctx := withErrorContext(r.Context())

		// 调用下一个处理器
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// 错误上下文类型
type errorContextKey struct{}

// withErrorContext 创建带有错误上下文的上下文
func withErrorContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, errorContextKey{}, make([]*AppError, 0))
}

// AddErrorToContext 将错误添加到上下文中
func AddErrorToContext(ctx context.Context, err *AppError) context.Context {
	errors, ok := ctx.Value(errorContextKey{}).([]*AppError)
	if !ok {
		errors = make([]*AppError, 0)
	}
	errors = append(errors, err)
	return context.WithValue(ctx, errorContextKey{}, errors)
}

// GetErrorsFromContext 从上下文中获取错误
func GetErrorsFromContext(ctx context.Context) []*AppError {
	errors, ok := ctx.Value(errorContextKey{}).([]*AppError)
	if !ok {
		return make([]*AppError, 0)
	}
	return errors
}

// HasErrors 检查上下文中是否有错误
func HasErrors(ctx context.Context) bool {
	errors := GetErrorsFromContext(ctx)
	return len(errors) > 0
}

// OperationWithContext 带错误上下文的操作执行
func (m *Middleware) OperationWithContext(ctx context.Context, operationName string, operation func() error) error {
	var err error

	// 执行操作
	defer func() {
		if r := recover(); r != nil {
			m.logger.Errorf("操作 %s 中发生panic: %v", operationName, r)
			err = NewSystemError(operationName+"中发生panic", r.(error))
		}
	}()

	err = operation()

	// 如果有错误，处理它
	if err != nil {
		var appErr *AppError
		var ok bool

		// 检查是否已经是AppError
		if appErr, ok = err.(*AppError); !ok {
			// 如果不是，转换为AppError
			appErr = NewSystemError(operationName+"失败", err)
		}

		// 处理错误
		handledErr := m.errorHandler.HandleError(appErr)

		// 将错误添加到上下文中
		_ = AddErrorToContext(ctx, appErr)

		return handledErr
	}

	return nil
}

// RetryableOperationWithContext 带错误上下文和重试的操作执行
func (m *Middleware) RetryableOperationWithContext(ctx context.Context, operationName string, operation func() error) error {
	var lastErr error
	maxAttempts := 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// 执行操作
		err := m.OperationWithContext(ctx, operationName, operation)
		if err == nil {
			return nil
		}

		lastErr = err

		// 检查错误是否可重试
		if !m.errorHandler.ShouldRetry(err) {
			break
		}

		// 获取重试延迟
		delay := m.errorHandler.GetRetryDelay(err, attempt)

		// 记录重试日志
		m.logger.Infof("操作 %s 第%d次尝试失败，%v后重试: %v", operationName, attempt, delay, err)

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
