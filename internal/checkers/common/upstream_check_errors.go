package common

import (
	"fmt"
	"net/url"
	"time"
	"aur-update-checker/internal/errors"
)

// CheckerError 检查器错误类型，包装了统一的AppError
// 所有检查器特定的错误都应该使用此类型，它继承自errors包中的AppError
type CheckerError struct {
	*errors.AppError
	URL string
}

// Error 实现error接口
func (e *CheckerError) Error() string {
	if e.AppError != nil {
		return fmt.Sprintf("%s (URL: %s)", e.AppError.Error(), e.URL)
	}
	return fmt.Sprintf("未知错误 (URL: %s)", e.URL)
}

// NewCheckerError 创建检查器错误
func NewCheckerError(errorType errors.ErrorCode, message, url string, details error) *CheckerError {
	appErr := errors.NewAppError(errorType, message, details)
	return &CheckerError{
		AppError: appErr,
		URL:      url,
	}
}

// NewNetworkError 创建网络错误
func NewNetworkError(url string, details error) *CheckerError {
	appErr := errors.NewNetworkError("网络请求失败", details)
	return &CheckerError{
		AppError: appErr,
		URL:      url,
	}
}

// NewParseError 创建解析错误
func NewParseError(url string, details error) *CheckerError {
	appErr := errors.NewParseError("解析响应失败", details)
	return &CheckerError{
		AppError: appErr,
		URL:      url,
	}
}

// NewFormatError 创建格式错误
func NewFormatError(url, message string) *CheckerError {
	appErr := errors.NewConfigurationError(message, nil)
	return &CheckerError{
		AppError: appErr,
		URL:      url,
	}
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(url string) *CheckerError {
	appErr := errors.NewNotFoundError("未找到所需信息", nil)
	return &CheckerError{
		AppError: appErr,
		URL:      url,
	}
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(url string) *CheckerError {
	appErr := errors.NewTimeoutError("请求超时", nil)
	return &CheckerError{
		AppError: appErr,
		URL:      url,
	}
}

// NewPermissionError 创建权限错误
func NewPermissionError(url string) *CheckerError {
	appErr := errors.NewPermissionError("权限不足", nil)
	return &CheckerError{
		AppError: appErr,
		URL:      url,
	}
}

// NewUnsupportedError 创建不支持错误
func NewUnsupportedError(url, message string) *CheckerError {
	appErr := errors.NewConfigurationError(message, nil)
	return &CheckerError{
		AppError: appErr,
		URL:      url,
	}
}

// ValidateURL 验证URL格式是否正确
func ValidateURL(urlStr string) (*url.URL, error) {
	if urlStr == "" {
		return nil, NewFormatError("", "URL不能为空")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, NewFormatError(urlStr, fmt.Sprintf("URL格式无效: %v", err))
	}

	// 检查协议
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, NewFormatError(urlStr, "只支持http和https协议")
	}

	// 检查主机名
	if parsedURL.Hostname() == "" {
		return nil, NewFormatError(urlStr, "URL中缺少主机名")
	}

	return parsedURL, nil
}

// IsNetworkError 检查是否为网络错误
func IsNetworkError(err error) bool {
	if checkerErr, ok := err.(*CheckerError); ok {
		return checkerErr.AppError.Code == errors.NetworkError
	}
	return false
}

// IsParseError 检查是否为解析错误
func IsParseError(err error) bool {
	if checkerErr, ok := err.(*CheckerError); ok {
		return checkerErr.AppError.Code == errors.ParseError
	}
	return false
}

// IsFormatError 检查是否为格式错误
func IsFormatError(err error) bool {
	if checkerErr, ok := err.(*CheckerError); ok {
		return checkerErr.AppError.Code == errors.ConfigurationError
	}
	return false
}

// IsNotFoundError 检查是否为未找到错误
func IsNotFoundError(err error) bool {
	if checkerErr, ok := err.(*CheckerError); ok {
		return checkerErr.AppError.Code == errors.NotFoundError
	}
	return false
}

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	if checkerErr, ok := err.(*CheckerError); ok {
		return checkerErr.AppError.Code == errors.TimeoutError ||
		       checkerErr.AppError.Code == errors.NetworkTimeoutError
	}
	return false
}

// IsPermissionError 检查是否为权限错误
func IsPermissionError(err error) bool {
	if checkerErr, ok := err.(*CheckerError); ok {
		return checkerErr.AppError.Code == errors.PermissionError
	}
	return false
}

// IsUnsupportedError 检查是否为不支持错误
func IsUnsupportedError(err error) bool {
	if checkerErr, ok := err.(*CheckerError); ok {
		return checkerErr.AppError.Code == errors.ConfigurationError
	}
	return false
}

// IsRetryableError 检查错误是否可重试
func IsRetryableError(err error) bool {
	if checkerErr, ok := err.(*CheckerError); ok {
		return checkerErr.AppError.Retryable
	}
	return false
}

// GetErrorContext 获取错误的上下文信息
func GetErrorContext(err error) map[string]interface{} {
	if checkerErr, ok := err.(*CheckerError); ok {
		return checkerErr.AppError.Context
	}
	return nil
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) errors.ErrorCode {
	if checkerErr, ok := err.(*CheckerError); ok {
		return checkerErr.AppError.Code
	}
	return errors.SystemError
}

// WithRetryable 设置错误是否可重试
func WithRetryable(err error, retryable bool) error {
	if checkerErr, ok := err.(*CheckerError); ok {
		checkerErr.AppError.WithRetryable(retryable)
		return err
	}
	return err
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(err error, maxRetries int) error {
	if checkerErr, ok := err.(*CheckerError); ok {
		checkerErr.AppError.WithMaxRetries(maxRetries)
		return err
	}
	return err
}

// WithBackoff 设置退避时间
func WithBackoff(err error, backoff time.Duration) error {
	if checkerErr, ok := err.(*CheckerError); ok {
		checkerErr.AppError.WithBackoff(backoff)
		return err
	}
	return err
}

// WithContext 添加上下文信息
func WithContext(err error, key string, value interface{}) error {
	if checkerErr, ok := err.(*CheckerError); ok {
		checkerErr.AppError.WithContext(key, value)
		return err
	}
	return err
}
