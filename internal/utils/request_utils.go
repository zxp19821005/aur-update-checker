package utils

import (
	"fmt"
	"io"
	"net/http"
	"time"
	"aur-update-checker/internal/logger"
)

// HTTPClient 自定义HTTP客户端
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient 创建自定义HTTP客户端
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Get 发送GET请求
func (c *HTTPClient) Get(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.GlobalLogger.Errorf("[http] 创建请求失败: %v", err)
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "AUR-Update-Checker/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		logger.GlobalLogger.Errorf("[http] 请求失败: %v", err)
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.GlobalLogger.Errorf("[http] HTTP错误: %s", resp.Status)
		return "", fmt.Errorf("HTTP错误: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GlobalLogger.Errorf("[http] 读取响应失败: %v", err)
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	return string(body), nil
}

// GetWithRetry 带重试的GET请求
func (c *HTTPClient) GetWithRetry(url string, maxRetries int) (string, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		body, err := c.Get(url)
		if err == nil {
			return body, nil
		}

		lastErr = err

		// 如果不是最后一次重试，则等待一段时间
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	logger.GlobalLogger.Errorf("[http] 请求失败，已重试%d次: %v", maxRetries, lastErr)
	return "", fmt.Errorf("请求失败，已重试%d次: %v", maxRetries, lastErr)
}

// DefaultHTTPClient 默认HTTP客户端
var DefaultHTTPClient = NewHTTPClient(30 * time.Second)

// Get 使用默认客户端发送GET请求
func Get(url string) (string, error) {
	return DefaultHTTPClient.Get(url)
}

// GetWithRetry 使用默认客户端发送带重试的GET请求
func GetWithRetry(url string, maxRetries int) (string, error) {
	return DefaultHTTPClient.GetWithRetry(url, maxRetries)
}
