package checkers

import (
	"aur-update-checker/internal/checkers/common"
	"aur-update-checker/internal/logger"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	checkerInterfaces "aur-update-checker/internal/interfaces/checkers"
	version "aur-update-checker/internal/checkers/version"
)

// HttpChecker HTTP检查器，用于JS网页
type HttpChecker struct {
	*checkerInterfaces.BaseChecker
	client *http.Client
}

// NewHttpChecker 创建HTTP检查器
func NewHttpChecker() *HttpChecker {
	return &HttpChecker{
		BaseChecker: checkerInterfaces.NewBaseChecker("http"),
		client:      &http.Client{},
	}
}

// Check 实现检查器接口，通过浏览器方式获取页面内容并提取版本
func (c *HttpChecker) Check(ctx context.Context, url, versionExtractKey string) (string, error) {
	// 默认不检查测试版本
	return c.CheckWithOption(ctx, url, versionExtractKey, 0)
}

// CheckWithOption 实现检查器接口，根据选项检查上游版本
func (c *HttpChecker) CheckWithOption(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error) {
	logger.GlobalLogger.Infof("[HTTP检查器] 开始检查上游版本 - URL: %s, 提取键: %s, 检查测试版本: %d", url, versionExtractKey, checkTestVersion)

	if versionExtractKey == "" {
		logger.GlobalLogger.Errorf("[HTTP检查器] versionExtractKey为空")
		return "", fmt.Errorf("HTTP检查器需要提供versionExtractKey来定位版本信息")
	}

	// 获取页面内容
	logger.GlobalLogger.Debugf("[HTTP检查器] 获取页面内容: %s", url)
	content, err := c.fetchContent(ctx, url)
	if err != nil {
		logger.GlobalLogger.Errorf("[HTTP检查器] 获取页面内容失败: %v", err)
		return "", fmt.Errorf("获取页面内容失败: %v", err)
	}
	logger.GlobalLogger.Debugf("[HTTP检查器] 成功获取页面内容，长度: %d", len(content))

	// 提取版本
	logger.GlobalLogger.Debugf("[HTTP检查器] 开始提取版本，提取键: %s", versionExtractKey)
	version, err := c.extractVersion(content, versionExtractKey)
	if err != nil {
		logger.GlobalLogger.Errorf("[HTTP检查器] 从页面内容提取版本失败: %v", err)
		return "", fmt.Errorf("从页面内容提取版本失败: %v", err)
	}
	logger.GlobalLogger.Infof("[HTTP检查器] 成功提取版本: %s", version)

	// 规范化版本号，移除平台特定信息
	normalizedVersion := c.BaseChecker.NormalizeVersionWithOption(version, checkTestVersion)
	logger.GlobalLogger.Infof("[HTTP检查器] 版本规范化完成，最终版本: %s", normalizedVersion)
	return normalizedVersion, nil
}

// fetchContent 获取页面内容
// 注意：这里简化了实现，实际应该使用像playwright或chromedp这样的库来渲染JS页面
func (c *HttpChecker) fetchContent(ctx context.Context, url string) (string, error) {
	logger.GlobalLogger.Debugf("[HTTP检查器] 创建HTTP请求: %s", url)

	// 处理单页应用URL，去掉#后面的部分，因为服务器只返回基础HTML
	baseURL := url
	if idx := strings.Index(url, "#"); idx != -1 {
		baseURL = url[:idx]
		logger.GlobalLogger.Infof("[HTTP检查器] 检测到单页应用URL，使用基础URL: %s", baseURL)
	}

	logger.GlobalLogger.Debugf("[HTTP检查器] 实际请求URL: %s", baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		logger.GlobalLogger.Errorf("[HTTP检查器] 创建请求失败: %v", err)
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置User-Agent，模拟浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	logger.GlobalLogger.Debugf("[HTTP检查器] 设置请求头完成")

	logger.GlobalLogger.Debugf("[HTTP检查器] 发送HTTP请求")
	resp, err := c.client.Do(req)
	if err != nil {
		logger.GlobalLogger.Errorf("[HTTP检查器] 请求失败: %v", err)
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	logger.GlobalLogger.Debugf("[HTTP检查器] 收到HTTP响应，状态码: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		logger.GlobalLogger.Errorf("[HTTP检查器] 请求失败，状态码: %d", resp.StatusCode)
		return "", fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	logger.GlobalLogger.Debugf("[HTTP检查器] 读取响应体")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GlobalLogger.Errorf("[HTTP检查器] 读取响应体失败: %v", err)
		return "", fmt.Errorf("读取响应体失败: %v", err)
	}

	content := string(body)
	logger.GlobalLogger.Debugf("[HTTP检查器] 成功获取页面内容，长度: %d", len(content))

	// 检查是否是单页应用，并尝试从初始HTML中提取API数据
	isSPA := strings.Contains(content, "<div id=\"app\"") || 
	         strings.Contains(content, "angular") || 
	         strings.Contains(content, "react") || 
	         strings.Contains(content, "vue")

	logger.GlobalLogger.Debugf("[HTTP检查器] 检测是否为单页应用: %v", isSPA)

	if isSPA {
		logger.GlobalLogger.Infof("[HTTP检查器] 检测到可能是单页应用，尝试从HTML中提取数据")
		// 尝试从HTML中提取可能的API数据或版本信息
		if apiData, found := c.extractAPIDataFromHTML(content); found {
			logger.GlobalLogger.Infof("[HTTP检查器] 从HTML中提取到API数据")
			return apiData, nil
		}
		logger.GlobalLogger.Warnf("[HTTP检查器] 未从HTML中提取到API数据，返回原始HTML")
	}

	// 添加调试信息，检查内容中是否包含我们需要的键
	if strings.Contains(content, "Linux") {
		logger.GlobalLogger.Debugf("[HTTP检查器] 内容中包含'Linux'")
	}
	if strings.Contains(content, "信创") {
		logger.GlobalLogger.Debugf("[HTTP检查器] 内容中包含'信创'")
	}

	return content, nil
}

// extractVersion 从页面内容中提取版本
func (c *HttpChecker) extractVersion(content, versionExtractKey string) (string, error) {
	logger.GlobalLogger.Debugf("[HTTP检查器] 开始从页面内容提取版本，提取键: %s", versionExtractKey)

	// 查找所有匹配版本提取键的内容
	versions := c.findAllVersions(content, versionExtractKey)
	logger.GlobalLogger.Debugf("[HTTP检查器] 找到 %d 个匹配版本提取键的内容", len(versions))

	if len(versions) == 0 {
		logger.GlobalLogger.Warnf("[HTTP检查器] 未找到匹配版本提取键的内容: %s", versionExtractKey)
		return "", fmt.Errorf("未找到匹配版本提取键的内容")
	}

	// 如果有多个版本，取最新的
	latestVersion := c.getLatestVersion(versions)
	logger.GlobalLogger.Infof("[HTTP检查器] 从多个版本中选择最新版本: %s", latestVersion)
	return latestVersion, nil
}

// findAllVersions 查找所有匹配版本提取键的内容
func (c *HttpChecker) findAllVersions(content, key string) []string {
	var versions []string

	// 特殊处理复合键，如 "Linux&信创"
	if strings.Contains(key, "&") {
		logger.GlobalLogger.Debugf("[HTTP检查器] 检测到复合键: %s", key)
		keys := strings.Split(key, "&")
		if len(keys) >= 2 {
			logger.GlobalLogger.Debugf("[HTTP检查器] 分割后的键: %v", keys)
			// 查找所有键的组合
			combinedResults := c.findCombinedKeys(content, keys)
			logger.GlobalLogger.Debugf("[HTTP检查器] 复合键找到的结果数: %d", len(combinedResults))
			return combinedResults
		}
	}

	// 查找key在content中的所有位置
	logger.GlobalLogger.Debugf("[HTTP检查器] 开始查找键: %s", key)
	index := 0
	matchCount := 0
	for {
		pos := strings.Index(content[index:], key)
		if pos == -1 {
			logger.GlobalLogger.Debugf("[HTTP检查器] 未找到更多匹配")
			break
		}
		matchCount++

		// 计算绝对位置
		absPos := index + pos

		// 提取前后50个字符
		start := absPos - 50
		if start < 0 {
			start = 0
		}
		end := absPos + len(key) + 50
		if end > len(content) {
			end = len(content)
		}

		extract := content[start:end]
		versions = append(versions, extract)
		logger.GlobalLogger.Debugf("[HTTP检查器] 找到匹配 #%d: %s", matchCount, extract)

		// 移动到下一个位置继续搜索
		index = absPos + len(key)
	}

	logger.GlobalLogger.Debugf("[HTTP检查器] 总共找到 %d 个匹配", matchCount)
	return versions
}

// findCombinedKeys 查找多个键的组合
// 使用公共函数 common.FindCombinedKeys 实现
func (c *HttpChecker) findCombinedKeys(content string, keys []string) []string {
	logger.GlobalLogger.Debugf("[HTTP检查器] 调用公共函数查找复合键")
	return common.FindCombinedKeys(content, keys)
}

// getLatestVersion 从多个版本中获取最新的
func (c *HttpChecker) getLatestVersion(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	if len(versions) == 1 {
		// 尝试从单个版本字符串中提取版本号
		return c.extractVersionFromString(versions[0])
	}

	// 多个版本，尝试提取版本号并比较
	var versionNumbers []string
	for _, v := range versions {
		num := c.extractVersionFromString(v)
		if num != "" {
			versionNumbers = append(versionNumbers, num)
		}
	}

	if len(versionNumbers) == 0 {
		// 如果无法提取版本号，返回第一个匹配
		return versions[0]
	}

	// 简单比较版本号，返回最大的
	latest := versionNumbers[0]
	for _, v := range versionNumbers[1:] {
		if c.compareVersions(v, latest) > 0 {
			latest = v
		}
	}

	return latest
}

// extractVersionFromString 从字符串中提取版本号
// 使用公共函数 common.ExtractVersionFromString 实现
func (c *HttpChecker) extractVersionFromString(s string) string {
	logger.GlobalLogger.Debugf("[HTTP检查器] 调用公共函数提取版本号")
	return common.ExtractVersionFromString(s)
}

// compareVersions 比较两个版本号
// 使用公共函数 version.NewVersionComparator().CompareVersions 实现
func (c *HttpChecker) compareVersions(v1, v2 string) int {
	comparator := version.NewVersionComparator()
	return comparator.CompareVersions(v1, v2)
}

// extractAPIDataFromHTML 从HTML中提取可能的API数据或版本信息
func (c *HttpChecker) extractAPIDataFromHTML(html string) (string, bool) {
	// 尝试提取包含版本信息的script标签内容
	// 1. 查找JSON数据
	re := regexp.MustCompile(`<script[^>]*>\s*window\.__INITIAL_STATE__\s*=\s*({.+?});\s*</script>`)
	matches := re.FindStringSubmatch(html)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[HTTP检查器] 找到window.__INITIAL_STATE__数据")
		return matches[1], true
	}

	// 2. 查找其他可能的JSON数据
	re = regexp.MustCompile(`<script[^>]*>\s*var\s+\w+\s*=\s*({.+?});\s*</script>`)
	matches = re.FindStringSubmatch(html)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[HTTP检查器] 找到可能的JSON数据")
		return matches[1], true
	}

	// 3. 查找API端点
	re = regexp.MustCompile(`apiUrl\s*[:=]\s*["']([^"']+)["']`)
	matches = re.FindStringSubmatch(html)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[HTTP检查器] 找到API端点: %s", matches[1])
		return matches[1], true
	}

	// 4. 查找版本号
	re = regexp.MustCompile(`version\s*[:=]\s*["'](\d+\.\d+\.\d+)["']`)
	matches = re.FindStringSubmatch(html)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[HTTP检查器] 找到版本号: %s", matches[1])
		return matches[1], true
	}

	// 5. 查找下载链接中的版本号
	re = regexp.MustCompile(`download[_-]?url[^>]*>.*?(\d+\.\d+\.\d+).*?</`)
	matches = re.FindStringSubmatch(html)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[HTTP检查器] 从下载链接中找到版本号: %s", matches[1])
		return matches[1], true
	}

	// 6. 查找Linux和信创相关信息，并尝试提取附近的版本号
	re = regexp.MustCompile(`(Linux|信创)[^<]{0,100}(\d+\.\d+\.\d+)`)
	matches = re.FindStringSubmatch(html)
	if len(matches) >= 3 {
		logger.GlobalLogger.Debugf("[HTTP检查器] 找到Linux/信创相关版本号: %s", matches[2])
		return matches[2], true
	}

	// 7. 查找Linux和信创相关信息
	re = regexp.MustCompile(`(Linux|信创)[^<]*([\d\.]+)`)
	matches = re.FindStringSubmatch(html)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[HTTP检查器] 找到Linux/信创相关信息: %s", matches[0])
		return matches[0], true
	}

	return "", false
}

// CheckWithVersionRef 带选项和版本引用地检查上游版本
func (c *HttpChecker) CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	// 简单地调用CheckWithOption方法，忽略versionRef参数
	return c.CheckWithOption(ctx, url, versionExtractKey, checkTestVersion)
}
