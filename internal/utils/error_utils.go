package utils

import (
	"fmt"
	"aur-update-checker/internal/errors"
	"aur-update-checker/internal/logger"
)

// ErrorHandler 统一错误处理器，封装了errors包中的DefaultErrorHandler
type ErrorHandler struct {
	logger      *logger.Logger
	errorHandler errors.ErrorHandler
}

// NewErrorHandler 创建新的错误处理器
func NewErrorHandler(logger *logger.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger:      logger,
		errorHandler: errors.NewDefaultErrorHandler(*logger),
	}
}

// HandleError 统一处理错误，记录日志并返回处理后的错误
func (h *ErrorHandler) HandleError(err error, context ...string) error {
	if err == nil {
		return nil
	}

	// 使用 errors 包中的错误处理器处理错误
	h.errorHandler.HandleError(err)

	// 记录错误日志
	if len(context) > 0 {
		h.logger.Errorf("%s: %v", context[0], err)
	} else {
		h.logger.Errorf("发生错误: %v", err)
	}

	return err
}

// HandleErrorWithLevel 根据错误级别处理错误
func (h *ErrorHandler) HandleErrorWithLevel(err error, level string, context ...string) error {
	if err == nil {
		return nil
	}

	// 根据级别记录日志
	switch level {
	case "debug":
		if len(context) > 0 {
			h.logger.Debugf("%s: %v", context[0], err)
		} else {
			h.logger.Debugf("调试错误: %v", err)
		}
	case "info":
		if len(context) > 0 {
			h.logger.Infof("%s: %v", context[0], err)
		} else {
			h.logger.Infof("信息错误: %v", err)
		}
	case "warn":
		if len(context) > 0 {
			h.logger.Warnf("%s: %v", context[0], err)
		} else {
			h.logger.Warnf("警告错误: %v", err)
		}
	default:
		// 默认使用错误级别
		return h.HandleError(err, context...)
	}

	return err
}

// WrapError 包装错误，添加上下文信息
func (h *ErrorHandler) WrapError(err error, message string) error {
	if err == nil {
		return nil
	}

	// 如果是AppError，添加上下文
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr.WithContext("wrapped_message", message)
	}

	// 如果是普通错误，创建新的AppError
	return errors.NewSystemError(fmt.Sprintf("%s: %v", message, err), err)
}

// IsRetryableError 检查错误是否可重试
func (h *ErrorHandler) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 使用 errors 包中的错误处理器检查错误是否可重试
	return h.errorHandler.ShouldRetry(err)
}

// GetErrorDetails 获取错误详细信息
func (h *ErrorHandler) GetErrorDetails(err error) string {
	if err == nil {
		return ""
	}

	if appErr, ok := err.(*errors.AppError); ok {
		details := fmt.Sprintf("[%d] %s: %s", appErr.Code, appErr.Message, appErr.Details)
		if appErr.OriginalErr != nil {
			details += fmt.Sprintf(" (原始错误: %v)", appErr.OriginalErr)
		}
		if len(appErr.Context) > 0 {
			details += fmt.Sprintf(" (上下文: %v)", appErr.Context)
		}
		return details
	}

	return err.Error()
}

// 全局错误处理器实例
var GlobalErrorHandler *ErrorHandler

// InitGlobalErrorHandler 初始化全局错误处理器
func InitGlobalErrorHandler(logger *logger.Logger) {
	GlobalErrorHandler = NewErrorHandler(logger)
}

// HandleError 全局错误处理函数
func HandleError(err error, context ...string) error {
	if GlobalErrorHandler == nil {
		// 如果全局错误处理器未初始化，使用默认日志记录
		if len(context) > 0 {
			logger.GlobalLogger.Errorf("%s: %v", context[0], err)
		} else {
			logger.GlobalLogger.Errorf("发生错误: %v", err)
		}
		return err
	}
	return GlobalErrorHandler.HandleError(err, context...)
}

// HandleErrorWithLevel 全局带级别的错误处理函数
func HandleErrorWithLevel(err error, level string, context ...string) error {
	if GlobalErrorHandler == nil {
		// 如果全局错误处理器未初始化，使用默认日志记录
		switch level {
		case "debug":
			if len(context) > 0 {
				logger.GlobalLogger.Debugf("%s: %v", context[0], err)
			} else {
				logger.GlobalLogger.Debugf("调试错误: %v", err)
			}
		case "info":
			if len(context) > 0 {
				logger.GlobalLogger.Infof("%s: %v", context[0], err)
			} else {
				logger.GlobalLogger.Infof("信息错误: %v", err)
			}
		case "warn":
			if len(context) > 0 {
				logger.GlobalLogger.Warnf("%s: %v", context[0], err)
			} else {
				logger.GlobalLogger.Warnf("警告错误: %v", err)
			}
		default:
			if len(context) > 0 {
				logger.GlobalLogger.Errorf("%s: %v", context[0], err)
			} else {
				logger.GlobalLogger.Errorf("发生错误: %v", err)
			}
		}
		return err
	}
	return GlobalErrorHandler.HandleErrorWithLevel(err, level, context...)
}

// WrapError 全局错误包装函数
func WrapError(err error, message string) error {
	if GlobalErrorHandler == nil {
		return fmt.Errorf("%s: %v", message, err)
	}
	return GlobalErrorHandler.WrapError(err, message)
}

// IsRetryableError 全局检查错误是否可重试函数
func IsRetryableError(err error) bool {
	if GlobalErrorHandler == nil {
		return false
	}
	return GlobalErrorHandler.IsRetryableError(err)
}

// GetErrorDetails 全局获取错误详细信息函数
func GetErrorDetails(err error) string {
	if GlobalErrorHandler == nil {
		if err == nil {
			return ""
		}
		return err.Error()
	}
	return GlobalErrorHandler.GetErrorDetails(err)
}
