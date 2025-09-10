package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"aur-update-checker/internal/logger"
)

// getAurPackageInfo 从AUR API获取软件包信息
func (s *AurService) getAurPackageInfo(packageName string) (*AurPackage, error) {
	// 构建AUR API URL
	url := fmt.Sprintf("https://aur.archlinux.org/rpc/?v=5&type=info&arg[]=%s", packageName)

	// 添加重试逻辑
	maxRetries := 3
	retryDelay := time.Second * 2
	var lastErr error

	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			logger.GlobalLogger.Debugf("重试第 %d 次获取AUR软件包信息: %s", retry, packageName)
			time.Sleep(retryDelay)
			// 每次重试增加延迟时间
			retryDelay *= 2
		}

		// 发送HTTP请求
		resp, err := http.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("请求AUR API失败: %v", err)
			logger.GlobalLogger.Debugf("请求AUR API失败 (尝试 %d/%d): %v", retry+1, maxRetries, err)
			continue
		}
		defer resp.Body.Close()

		// 检查响应状态码
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("AUR API返回错误状态码: %d", resp.StatusCode)
			logger.GlobalLogger.Debugf("AUR API返回错误状态码 (尝试 %d/%d): %d", retry+1, maxRetries, resp.StatusCode)
			continue
		}

		// 读取响应体
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("读取AUR API响应失败: %v", err)
			logger.GlobalLogger.Debugf("读取AUR API响应失败 (尝试 %d/%d): %v", retry+1, maxRetries, err)
			continue
		}

		// 如果成功读取响应体，继续处理
		if len(body) > 0 {
			// 解析JSON响应
			var aurResponse AurResponse
			if err := json.Unmarshal(body, &aurResponse); err != nil {
				lastErr = fmt.Errorf("解析AUR API响应失败: %v", err)
				logger.GlobalLogger.Debugf("解析AUR API响应失败 (尝试 %d/%d): %v", retry+1, maxRetries, err)
				continue
			}

			// 检查结果
			if aurResponse.Resultcount == 0 {
				lastErr = fmt.Errorf("在AUR中未找到软件包: %s", packageName)
				logger.GlobalLogger.Debugf("在AUR中未找到软件包 (尝试 %d/%d): %s", retry+1, maxRetries, packageName)
				continue
			}

			// 成功获取软件包信息，返回结果
			return &aurResponse.Results[0], nil
		} else {
			lastErr = fmt.Errorf("AUR API返回空响应体")
			logger.GlobalLogger.Debugf("AUR API返回空响应体 (尝试 %d/%d)", retry+1, maxRetries)
			continue
		}
	}

	// 所有重试都失败，返回最后一个错误
	return nil, lastErr
}

// getAurPackagesInfo 从AUR API批量获取软件包信息
func (s *AurService) getAurPackagesInfo(packageNames []string) ([]AurPackage, error) {
	// 如果没有软件包名称，直接返回空结果
	if len(packageNames) == 0 {
		return []AurPackage{}, nil
	}

	// 构建AUR API URL，支持批量查询
	// AUR RPC接口支持一次查询多个软件包，通过多个arg[]参数实现
	var urlBuilder strings.Builder
	urlBuilder.WriteString("https://aur.archlinux.org/rpc/?v=5&type=info")

	for _, name := range packageNames {
		urlBuilder.WriteString("&arg[]=")
		urlBuilder.WriteString(name)
	}

	url := urlBuilder.String()

	// 发送HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求AUR API失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AUR API返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取AUR API响应失败: %v", err)
	}

	// 解析JSON响应
	var aurResponse AurResponse
	if err := json.Unmarshal(body, &aurResponse); err != nil {
		return nil, fmt.Errorf("解析AUR API响应失败: %v", err)
	}

	// 返回结果
	return aurResponse.Results, nil
}
