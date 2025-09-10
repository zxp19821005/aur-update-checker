package checkers

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

// generateAsyncCheckID 生成异步检查ID
func generateAsyncCheckID(url, versionExtractKey string, checkTestVersion int) string {
	// 使用URL、版本提取键和测试版本标志生成唯一ID
	data := fmt.Sprintf("%s-%s-%d-%d", url, versionExtractKey, checkTestVersion, time.Now().UnixNano())
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))[:16]
}

// isCriticalError 判断是否为严重错误
func isCriticalError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// 网络错误通常不是严重错误，可以重试
	if isNetworkError(err) {
		return false
	}

	// 以下错误类型被认为是严重错误，不应重试
	criticalErrors := []string{
		"invalid URL",
		"unsupported protocol scheme",
		"no such host",
		"context canceled",
		"timeout",
	}

	for _, ce := range criticalErrors {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(ce)) {
			return true
		}
	}

	return false
}

// isNetworkError 判断是否为网络错误
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// 常见的网络错误标识
	networkErrors := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"network is unreachable",
		"no route to host",
		"temporary failure",
		"deadline exceeded",
		"operation timed out",
		"request canceled",
		"client timeout",
		"TLS handshake",
	}

	for _, ne := range networkErrors {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(ne)) {
			return true
		}
	}

	// 检查是否为特定的网络错误类型
	if _, ok := err.(net.Error); ok {
		return true
	}

	// 检查URL错误
	if urlErr, ok := err.(*url.Error); ok {
		return isNetworkError(urlErr.Err)
	}

	return false
}
