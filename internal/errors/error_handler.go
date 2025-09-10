package errors

import (
	"fmt"
	"aur-update-checker/internal/logger"
	"time"
)

// ErrorCode 定义错误码类型
type ErrorCode int

const (
	// 系统级错误 (1000-1999)
	SystemError ErrorCode = 1000 + iota
	ConfigurationError
	TimeoutError

	// 网络请求错误 (2000-2999)
	NetworkError ErrorCode = 2000 + iota
	HTTPError
	NotFoundError
	PermissionError
	RateLimitError
	NetworkTimeoutError

	// 解析错误 (3000-3999)
	ParseError ErrorCode = 3000 + iota
	JSONParseError
	HTMLParseError
	VersionParseError

	// 检查器特定错误 (4000-4999)
	CheckerError ErrorCode = 4000 + iota
	GitHubError
	GitLabError
	NPMError
	PyPIError
	CurlError
	HTTPErrorChecker
	JSONErrorChecker
	PlaywrightError
	RedirectError
	GiteeError
)

// ErrorType 定义错误类型
type ErrorType struct {
	Code    ErrorCode
	Message string
}

// 预定义的错误类型
var ErrorTypes = map[ErrorCode]ErrorType{
	SystemError:         {SystemError, "系统错误"},
	ConfigurationError:  {ConfigurationError, "配置错误"},
	TimeoutError:        {TimeoutError, "超时错误"},
	NetworkError:        {NetworkError, "网络错误"},
	HTTPError:           {HTTPError, "HTTP错误"},
	NotFoundError:       {NotFoundError, "资源未找到"},
	PermissionError:     {PermissionError, "权限不足"},
	RateLimitError:      {RateLimitError, "请求频率限制"},
	NetworkTimeoutError: {NetworkTimeoutError, "网络超时"},
	ParseError:          {ParseError, "解析错误"},
	JSONParseError:      {JSONParseError, "JSON解析错误"},
	HTMLParseError:      {HTMLParseError, "HTML解析错误"},
	VersionParseError:   {VersionParseError, "版本解析错误"},
	CheckerError:        {CheckerError, "检查器错误"},
	GitHubError:         {GitHubError, "GitHub检查器错误"},
	GitLabError:         {GitLabError, "GitLab检查器错误"},
	NPMError:            {NPMError, "NPM检查器错误"},
	PyPIError:           {PyPIError, "PyPI检查器错误"},
	CurlError:           {CurlError, "Curl检查器错误"},
	HTTPErrorChecker:    {HTTPErrorChecker, "HTTP检查器错误"},
	JSONErrorChecker:    {JSONErrorChecker, "JSON检查器错误"},
	PlaywrightError:     {PlaywrightError, "Playwright检查器错误"},
	RedirectError:       {RedirectError, "重定向检查器错误"},
	GiteeError:          {GiteeError, "Gitee检查器错误"},
}

// AppError 应用程序错误结构
type AppError struct {
	Code        ErrorCode
	Message     string
	Details     string
	OriginalErr error
	Context     map[string]interface{}
	Retryable   bool
	MaxRetries  int
	Backoff     time.Duration
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.OriginalErr != nil {
		return fmt.Sprintf("[%d] %s: %s (%v)", e.Code, e.Message, e.Details, e.OriginalErr)
	}
	return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
}

// NewAppError 创建新的应用程序错误
func NewAppError(code ErrorCode, details string, originalErr error) *AppError {
	errorType, exists := ErrorTypes[code]
	if !exists {
		errorType = ErrorType{code, "未知错误"}
	}

	return &AppError{
		Code:        errorType.Code,
		Message:     errorType.Message,
		Details:     details,
		OriginalErr: originalErr,
		Context:     make(map[string]interface{}),
		Retryable:   false,
		MaxRetries:  0,
		Backoff:     0,
	}
}

// WithContext 添加上下文信息
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.Context[key] = value
	return e
}

// WithRetryable 设置错误是否可重试
func (e *AppError) WithRetryable(retryable bool) *AppError {
	e.Retryable = retryable
	return e
}

// WithMaxRetries 设置最大重试次数
func (e *AppError) WithMaxRetries(maxRetries int) *AppError {
	e.MaxRetries = maxRetries
	return e
}

// WithBackoff 设置退避时间
func (e *AppError) WithBackoff(backoff time.Duration) *AppError {
	e.Backoff = backoff
	return e
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	HandleError(err error) error
	IsRecoverable(err error) bool
	ShouldRetry(err error) bool
	GetRetryDelay(err error, attempt int) time.Duration
}

// DefaultErrorHandler 默认错误处理器
type DefaultErrorHandler struct {
	logger logger.Logger
}

// NewDefaultErrorHandler 创建默认错误处理器
func NewDefaultErrorHandler(logger logger.Logger) *DefaultErrorHandler {
	return &DefaultErrorHandler{logger: logger}
}

// HandleError 处理错误
func (h *DefaultErrorHandler) HandleError(err error) error {
	if appErr, ok := err.(*AppError); ok {
		h.logger.Errorf("处理应用错误: %v", appErr)
		if len(appErr.Context) > 0 {
			h.logger.Debugf("错误上下文: %v", appErr.Context)
		}
		return err
	}

	h.logger.Errorf("处理未知错误: %v", err)
	return err
}

// IsRecoverable 检查错误是否可恢复
func (h *DefaultErrorHandler) IsRecoverable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Retryable
	}
	return false
}

// ShouldRetry 检查是否应该重试
func (h *DefaultErrorHandler) ShouldRetry(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Retryable && appErr.MaxRetries > 0
	}
	return false
}

// GetRetryDelay 获取重试延迟
func (h *DefaultErrorHandler) GetRetryDelay(err error, attempt int) time.Duration {
	if appErr, ok := err.(*AppError); ok {
		if appErr.Backoff > 0 {
			// 指数退避
			return appErr.Backoff * time.Duration(attempt)
		}
	}
	// 默认延迟
	return time.Second * time.Duration(attempt)
}

// RetryableOperation 可重试操作类型
type RetryableOperation func() (interface{}, error)

// WithRetry 带重试的操作执行
func WithRetry(operation RetryableOperation, errorHandler ErrorHandler) (interface{}, error) {
	var lastErr error
	maxAttempts := 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := operation()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// 检查错误是否可重试
		if !errorHandler.ShouldRetry(err) {
			break
		}

		// 获取重试延迟
		delay := errorHandler.GetRetryDelay(err, attempt)

		// 记录重试日志
		errorHandler.HandleError(fmt.Errorf("第%d次尝试失败，%v后重试: %v", attempt, delay, err))

		// 等待延迟时间
		time.Sleep(delay)
	}

	return nil, lastErr
}

// 预定义的错误创建函数
func NewSystemError(details string, originalErr error) *AppError {
	return NewAppError(SystemError, details, originalErr)
}

func NewConfigurationError(details string, originalErr error) *AppError {
	return NewAppError(ConfigurationError, details, originalErr)
}

func NewTimeoutError(details string, originalErr error) *AppError {
	return NewAppError(TimeoutError, details, originalErr).WithRetryable(true).WithMaxRetries(3).WithBackoff(time.Second * 2)
}

func NewNetworkError(details string, originalErr error) *AppError {
	return NewAppError(NetworkError, details, originalErr).WithRetryable(true).WithMaxRetries(3).WithBackoff(time.Second)
}

func NewHTTPError(details string, originalErr error) *AppError {
	return NewAppError(HTTPError, details, originalErr)
}

func NewNotFoundError(details string, originalErr error) *AppError {
	return NewAppError(NotFoundError, details, originalErr)
}

func NewPermissionError(details string, originalErr error) *AppError {
	return NewAppError(PermissionError, details, originalErr)
}

func NewRateLimitError(details string, originalErr error) *AppError {
	return NewAppError(RateLimitError, details, originalErr).WithRetryable(true).WithMaxRetries(3).WithBackoff(time.Second * 5)
}

func NewNetworkTimeoutError(details string, originalErr error) *AppError {
	return NewAppError(NetworkTimeoutError, details, originalErr).WithRetryable(true).WithMaxRetries(3).WithBackoff(time.Second * 2)
}

func NewParseError(details string, originalErr error) *AppError {
	return NewAppError(ParseError, details, originalErr)
}

func NewJSONParseError(details string, originalErr error) *AppError {
	return NewAppError(JSONParseError, details, originalErr)
}

func NewHTMLParseError(details string, originalErr error) *AppError {
	return NewAppError(HTMLParseError, details, originalErr)
}

func NewVersionParseError(details string, originalErr error) *AppError {
	return NewAppError(VersionParseError, details, originalErr)
}

func NewGitHubError(details string, originalErr error) *AppError {
	return NewAppError(GitHubError, details, originalErr)
}

func NewGitLabError(details string, originalErr error) *AppError {
	return NewAppError(GitLabError, details, originalErr)
}

func NewNPMError(details string, originalErr error) *AppError {
	return NewAppError(NPMError, details, originalErr)
}

func NewPyPIError(details string, originalErr error) *AppError {
	return NewAppError(PyPIError, details, originalErr)
}

func NewCurlError(details string, originalErr error) *AppError {
	return NewAppError(CurlError, details, originalErr)
}

func NewHTTPCheckerError(details string, originalErr error) *AppError {
	return NewAppError(HTTPErrorChecker, details, originalErr)
}

func NewJSONCheckerError(details string, originalErr error) *AppError {
	return NewAppError(JSONErrorChecker, details, originalErr)
}

func NewPlaywrightError(details string, originalErr error) *AppError {
	return NewAppError(PlaywrightError, details, originalErr)
}

func NewRedirectError(details string, originalErr error) *AppError {
	return NewAppError(RedirectError, details, originalErr)
}

func NewGiteeError(details string, originalErr error) *AppError {
	return NewAppError(GiteeError, details, originalErr)
}
