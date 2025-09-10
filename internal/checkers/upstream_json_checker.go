package checkers

import (
	"aur-update-checker/internal/logger"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	checkerInterfaces "aur-update-checker/internal/interfaces/checkers"
)

// JsonChecker JSON检查器
type JsonChecker struct {
	*checkerInterfaces.BaseChecker
	client *http.Client
}

// NewJsonChecker 创建JSON检查器
func NewJsonChecker() *JsonChecker {
	return &JsonChecker{
		BaseChecker: checkerInterfaces.NewBaseChecker("json"),
		client:      &http.Client{},
	}
}

// Check 实现检查器接口，从JSON文件中提取版本
func (c *JsonChecker) Check(ctx context.Context, url, versionExtractKey string) (string, error) {
	// 默认不检查测试版本
	return c.CheckWithOption(ctx, url, versionExtractKey, 0)
}

// CheckWithOption 实现检查器接口，根据选项从JSON文件中提取版本
func (c *JsonChecker) CheckWithOption(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error) {
	// 默认不使用版本引用
	return c.CheckWithVersionRef(ctx, url, versionExtractKey, "", checkTestVersion)
}

// CheckWithVersionRef 实现检查器接口，根据选项和版本引用从JSON文件中提取版本
func (c *JsonChecker) CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	if versionExtractKey == "" {
		logger.GlobalLogger.Errorf("[json] JSON检查器需要提供versionExtractKey来定位版本信息")
		return "", fmt.Errorf("JSON检查器需要提供versionExtractKey来定位版本信息")
	}

	// 获取JSON文件内容
	jsonData, err := c.fetchJSON(ctx, url)
	if err != nil {
		logger.GlobalLogger.Errorf("[json] 获取JSON文件失败: %v", err)
		return "", fmt.Errorf("获取JSON文件失败: %v", err)
	}

	// 解析JSON路径
	paths := strings.Split(versionExtractKey, ".")

	// 从JSON数据中提取版本
	version, err := c.extractVersionFromJSON(jsonData, paths)
	if err != nil {
		logger.GlobalLogger.Errorf("[json] 从JSON中提取版本失败: %v", err)
		return "", fmt.Errorf("从JSON中提取版本失败: %v", err)
	}

	// 如果提供了版本引用，使用版本引用来优化版本提取
	if versionRef != "" {
		logger.GlobalLogger.Debugf("[json] 使用版本引用 %s 来优化版本提取", versionRef)
		// 这里可以添加使用版本引用优化版本提取的逻辑
		// 例如，如果版本引用是 a.b.c，我们可以确保提取的版本号符合这个格式
	}

	// 规范化版本号，移除平台特定信息
	normalizedVersion := c.BaseChecker.NormalizeVersionWithOption(version, checkTestVersion)
	return normalizedVersion, nil
}

// fetchJSON 获取JSON文件内容
func (c *JsonChecker) fetchJSON(ctx context.Context, url string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logger.GlobalLogger.Errorf("[json] 创建请求失败: %v", err)
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		logger.GlobalLogger.Errorf("[json] 请求失败: %v", err)
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.GlobalLogger.Errorf("[json] 请求失败，状态码: %d", resp.StatusCode)
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GlobalLogger.Errorf("[json] 读取响应体失败: %v", err)
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.GlobalLogger.Errorf("[json] 解析JSON失败: %v", err)
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return result, nil
}

// extractVersionFromJSON 从JSON数据中提取版本
func (c *JsonChecker) extractVersionFromJSON(data map[string]interface{}, paths []string) (string, error) {
	var current interface{} = data

	// 遍历路径
	for i, path := range paths {
		switch v := current.(type) {
		case map[string]interface{}:
			if next, ok := v[path]; ok {
				current = next
			} else {
				logger.GlobalLogger.Errorf("[json] 路径 '%s' 不存在于JSON中", strings.Join(paths[:i+1], "."))
				return "", fmt.Errorf("路径 '%s' 不存在于JSON中", strings.Join(paths[:i+1], "."))
			}
		default:
			// 如果不是最后一个路径元素，但当前值不是map，则无法继续
			if i < len(paths)-1 {
				logger.GlobalLogger.Errorf("[json] 路径 '%s' 不是对象，无法继续", strings.Join(paths[:i+1], "."))
				return "", fmt.Errorf("路径 '%s' 不是对象，无法继续", strings.Join(paths[:i+1], "."))
			}
		}
	}

	// 处理最终值
	switch v := current.(type) {
	case string:
		return v, nil
	case float64:
		// JSON数字默认解析为float64
		return fmt.Sprintf("%v", v), nil
	case map[string]interface{}, []interface{}:
		// 如果是对象或数组，需要进一步处理
		// 这里简化处理，将整个结构转换为JSON字符串
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			logger.GlobalLogger.Errorf("[json] 将复杂结构转换为字符串失败: %v", err)
			return "", fmt.Errorf("将复杂结构转换为字符串失败: %v", err)
		}
		return string(jsonBytes), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}
