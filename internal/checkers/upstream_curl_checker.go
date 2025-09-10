package checkers

import (
	"aur-update-checker/internal/logger"
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"aur-update-checker/internal/checkers/common"
	checkers "aur-update-checker/internal/interfaces/checkers"
	versionProcessor "aur-update-checker/internal/checkers/version"
)

// CurlChecker curl检查器，用于非JS网页
type CurlChecker struct {
	*checkers.BaseChecker
	client *http.Client
}

// NewCurlChecker 创建curl检查器
func NewCurlChecker() *CurlChecker {
	// 创建具有连接池和合理超时设置的HTTP客户端
	transport := &http.Transport{
		MaxIdleConns:        100,               // 最大空闲连接数
		IdleConnTimeout:     90 * time.Second,  // 空闲连接超时时间
		DisableCompression:  false,             // 启用压缩
		MaxIdleConnsPerHost: 10,                // 每个主机的最大空闲连接数
		DisableKeepAlives:   false,             // 启用keep-alive
		MaxConnsPerHost:     100,               // 每个主机的最大连接数
		// 跳过SSL证书验证，解决证书过期问题
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,  // 总体请求超时时间
	}

	checker := &CurlChecker{
		BaseChecker: checkers.NewBaseChecker("curl"),
		client:      client,
	}

	return checker
}

// Check 实现检查器接口，通过curl方式获取页面内容并提取版本
func (c *CurlChecker) Check(ctx context.Context, url, versionExtractKey string) (string, error) {
	// 默认不检查测试版本，不使用版本引用
	return c.CheckWithVersionRef(ctx, url, versionExtractKey, "", 0)
}

// CheckWithOption 实现检查器接口，根据选项通过curl方式获取页面内容并提取版本
func (c *CurlChecker) CheckWithOption(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error) {
	// 默认不使用版本引用
	return c.CheckWithVersionRef(ctx, url, versionExtractKey, "", checkTestVersion)
}

// CheckWithVersionRef 实现检查器接口，根据选项和版本引用通过curl方式获取页面内容并提取版本
func (c *CurlChecker) CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	logger.GlobalLogger.Debugf("[curl] 开始检查版本 - URL: %s, 版本提取键: %s, 版本引用: %s, 检查测试版本: %d", url, versionExtractKey, versionRef, checkTestVersion)

	if versionExtractKey == "" {
		logger.GlobalLogger.Error("[curl] 版本提取键为空")
		return "", fmt.Errorf("curl检查器需要提供versionExtractKey来定位版本信息")
	}

	// 1. 使用curl获取上游URL的内容
	logger.GlobalLogger.Debugf("[curl] 正在获取页面内容...")
	content, err := c.fetchContent(ctx, url)
	if err != nil {
		errMsg := fmt.Errorf("获取页面内容失败: %v", err)
		logger.GlobalLogger.Errorf("[curl] %v", errMsg)
		return "", errMsg
	}
	logger.GlobalLogger.Debugf("[curl] 成功获取页面内容，长度: %d 字符", len(content))

	// 2. 使用版本提取关键字，提取版本提取关键字前后100个字符
	logger.GlobalLogger.Debugf("[curl] 使用版本提取关键字 '%s' 提取上下文", versionExtractKey)
	var contexts []string

	// 定义测试版本标识符
	testVersionPatterns := []string{"-alpha", "-beta", "-dev", "-rc", "-test", "-preview", "-pre"}
	
	// 查找所有版本提取关键字出现的位置
	logger.GlobalLogger.Debugf("[curl] 正在搜索版本提取关键字 '%s' 在内容中的位置", versionExtractKey)
	// 尝试直接使用versionExtractKey作为正则表达式
	keyPositions := regexp.MustCompile(versionExtractKey).FindAllStringIndex(content, -1)
	// 如果没有找到，尝试使用转义后的版本提取关键字
	if len(keyPositions) == 0 {
		logger.GlobalLogger.Debugf("[curl] 使用正则表达式未找到版本提取关键字，尝试使用转义后的字符串")
		keyPositions = regexp.MustCompile(regexp.QuoteMeta(versionExtractKey)).FindAllStringIndex(content, -1)
	}
	logger.GlobalLogger.Debugf("[curl] 找到 %d 个版本提取关键字位置", len(keyPositions))
	
	for i, pos := range keyPositions {
		logger.GlobalLogger.Debugf("[curl] 处理第 %d 个关键字位置: [%d, %d]", i+1, pos[0], pos[1])
		start := pos[0] - 100
		if start < 0 {
			start = 0
		}
		end := pos[1] + 100
		if end > len(content) {
			end = len(content)
		}
		context := content[start:end]
		
		// 如果不检查测试版本，检查上下文是否包含测试版本标识符
		if checkTestVersion == 0 {
			containsTestVersion := false
			for _, pattern := range testVersionPatterns {
				if strings.Contains(strings.ToLower(context), pattern) {
					containsTestVersion = true
					break
				}
			}
			
			if containsTestVersion {
				logger.GlobalLogger.Debugf("[curl] 上下文 %d 包含测试版本标识符，跳过", i+1)
				continue
			}
		}
		
		contexts = append(contexts, context)
		// 限制日志长度，避免日志过长
		logContext := context
		if len(logContext) > 200 {
			logContext = logContext[:200] + "..."
		}
		logger.GlobalLogger.Debugf("[curl] 提取到上下文 %d: %s", i+1, logContext)
	}
	
	if len(contexts) == 0 {
		errMsg := fmt.Errorf("在内容中未找到版本提取关键字 '%s'", versionExtractKey)
		logger.GlobalLogger.Errorf("[curl] %v", errMsg)
		return "", errMsg
	}

	// 3. 参考UpstreamVersionRef，使用extractVersionFromString从版本提取关键字前后100个字符尝试提取版本
	logger.GlobalLogger.Debugf("[curl] 开始从 %d 个上下文中提取版本号", len(contexts))
	var versions []string
	for i, context := range contexts {
		// 限制日志长度，避免日志过长
		logContext := context
		if len(logContext) > 200 {
			logContext = logContext[:200] + "..."
		}
		logger.GlobalLogger.Debugf("[curl] 从上下文 %d 中提取版本号: %s", i+1, logContext)
		version := c.extractVersionFromString(context)
		if version != "" {
			logger.GlobalLogger.Debugf("[curl] 从上下文 %d 中提取到版本: %s", i+1, version)
			versions = append(versions, version)
		} else {
			logger.GlobalLogger.Debugf("[curl] 从上下文 %d 中未能提取到版本号", i+1)
		}
	}

	if len(versions) == 0 {
		errMsg := fmt.Errorf("无法从提取的上下文中解析出有效的版本号")
		logger.GlobalLogger.Errorf("[curl] %v", errMsg)
		return "", errMsg
	}
	logger.GlobalLogger.Debugf("[curl] 共提取到 %d 个版本号: %v", len(versions), versions)

	// 4. 调用getLatestVersion，判定最新版本
	latestVersion := c.getLatestVersion(versions, checkTestVersion)
	if latestVersion == "" {
		errMsg := fmt.Errorf("无法从提取的版本中确定最新版本")
		logger.GlobalLogger.Errorf("[curl] %v", errMsg)
		return "", errMsg
	}
	logger.GlobalLogger.Debugf("[curl] 选择最新版本: %s", latestVersion)

	// 如果提供了版本引用，使用版本引用来优化版本提取
	if versionRef != "" {
		logger.GlobalLogger.Debugf("[curl] 使用版本引用 %s 来优化版本提取", versionRef)
		// 分析版本引用格式，确定版本号的结构
		// 例如，如果版本引用是 a.b.c，表示期望的版本号格式是 x.y.z
		versionFormat := c.analyzeVersionFormat(versionRef)
		logger.GlobalLogger.Debugf("[curl] 分析得到的版本格式: %s", versionFormat)

		// 根据版本格式筛选版本号
		filteredVersions := c.filterVersionsByFormat(versions, versionFormat)
		logger.GlobalLogger.Debugf("[curl] 根据版本格式筛选后的版本: %v", filteredVersions)

		if len(filteredVersions) > 0 {
			// 使用筛选后的版本重新选择最新版本
			latestVersion = c.getLatestVersion(filteredVersions, checkTestVersion)
			logger.GlobalLogger.Debugf("[curl] 使用版本引用筛选后的最新版本: %s", latestVersion)
		}
	}

	// 规范化版本号，移除平台特定信息
	normalizedVersion := c.BaseChecker.NormalizeVersionWithOption(latestVersion, checkTestVersion)
	logger.GlobalLogger.Debugf("[curl] 规范化后的版本: %s", normalizedVersion)
	return normalizedVersion, nil
}

// extractVersionFromString 从字符串中提取版本号
func (c *CurlChecker) extractVersionFromString(s string) string {
	logger.GlobalLogger.Debugf("[curl] 调用公共函数提取版本号")
	return common.ExtractVersionFromString(s)
}

// getLatestVersion 从多个版本中获取最新的版本
// analyzeVersionFormat 分析版本引用的格式
// 例如，如果版本引用是 a.b.c，返回 "x.y.z" 表示期望的版本号格式
func (c *CurlChecker) analyzeVersionFormat(versionRef string) string {
	// 统计版本引用中的点号数量，确定版本号的格式
	dotCount := strings.Count(versionRef, ".")

	// 根据点号数量确定格式
	switch dotCount {
	case 0:
		return "x" // 单个数字，如 1
	case 1:
		return "x.y" // 两个数字，如 1.2
	case 2:
		return "x.y.z" // 三个数字，如 1.2.3
	case 3:
		return "x.y.z.w" // 四个数字，如 1.2.3.4
	default:
		// 如果点号数量超过3，使用通用格式
		return "x.y.z+"
	}
}

// filterVersionsByFormat 根据版本格式筛选版本号
func (c *CurlChecker) filterVersionsByFormat(versions []string, format string) []string {
	var filtered []string

	for _, version := range versions {
		// 检查版本号是否符合指定的格式
		if c.versionMatchesFormat(version, format) {
			filtered = append(filtered, version)
		}
	}
	return filtered
}

// versionMatchesFormat 检查版本号是否符合指定的格式
func (c *CurlChecker) versionMatchesFormat(version, format string) bool {
	// 统计版本号中的点号数量
	dotCount := strings.Count(version, ".")
	// 根据格式检查点号数量
	switch format {
	case "x":
		return dotCount == 0
	case "x.y":
		return dotCount == 1
	case "x.y.z":
		return dotCount == 2
	case "x.y.z.w":
		return dotCount == 3
	case "x.y.z+":
		return dotCount >= 2
	default:
		// 默认情况下，接受所有格式
		return true
	}
}

func (c *CurlChecker) getLatestVersion(versions []string, checkTestVersion int) string {
	if len(versions) == 0 {
		return ""
	}

	// 去重
	uniqueVersions := make(map[string]bool)
	var dedupedVersions []string
	for _, v := range versions {
		if !uniqueVersions[v] {
			uniqueVersions[v] = true
			dedupedVersions = append(dedupedVersions, v)
		}
	}

	// 如果不检查测试版本，过滤掉包含测试版本标识的版本号
	if checkTestVersion == 0 {
		var filteredVersions []string
		for _, v := range dedupedVersions {
			// 检查是否包含测试版本标识，如alpha、beta、rc等
			isTestVersion := strings.Contains(strings.ToLower(v), "alpha") ||
				strings.Contains(strings.ToLower(v), "beta") ||
				strings.Contains(strings.ToLower(v), "rc") ||
				strings.Contains(strings.ToLower(v), "test") ||
				strings.Contains(strings.ToLower(v), "dev")

			if !isTestVersion {
				filteredVersions = append(filteredVersions, v)
			}
		}

		// 如果过滤后还有版本，使用过滤后的版本列表
		if len(filteredVersions) > 0 {
			dedupedVersions = filteredVersions
		}
	}

	// 简单比较版本号，返回最大的
	latest := dedupedVersions[0]
	comparator := versionProcessor.NewVersionComparator()
	for _, v := range dedupedVersions[1:] {
		if comparator.CompareVersions(v, latest) > 0 {
			latest = v
		}
	}

	return latest
}

// fetchContent 获取页面内容
func (c *CurlChecker) fetchContent(ctx context.Context, url string) (string, error) {
	// 验证URL格式
	_, err := common.ValidateURL(url)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		errMsg := fmt.Errorf("创建请求失败: %v", err)
		logger.GlobalLogger.Errorf("[curl] %v", errMsg)
		return "", common.NewNetworkError(url, errMsg)
	}

	// 设置通用的User-Agent
	req.Header.Set("User-Agent", "curl/8.15.0")

	// 设置基本请求头
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")

	// 对于特定网站添加额外的请求头
	if strings.Contains(url, "sourceforge.net") {
		req.Header.Set("Referer", "https://sourceforge.net/")
	}

	logger.GlobalLogger.Debugf("[curl] 发送请求到: %s", url)
	logger.GlobalLogger.Debugf("[curl] 使用User-Agent: %s", "curl/8.15.0")

	resp, err := c.client.Do(req)
	if err != nil {
		// 检查是否为超时错误
		if err == context.DeadlineExceeded {
			logger.GlobalLogger.Errorf("[curl] 请求超时: %v", err)
			return "", common.NewTimeoutError(url)
		}
		errMsg := fmt.Errorf("请求失败: %v", err)
		logger.GlobalLogger.Errorf("[curl] %v", errMsg)
		return "", common.NewNetworkError(url, errMsg)
	}
	defer resp.Body.Close()

	logger.GlobalLogger.Debugf("[curl] 收到响应，状态码: %d", resp.StatusCode)
	logger.GlobalLogger.Debugf("[curl] 响应头: %v", resp.Header)

	// 处理非200状态码
	if resp.StatusCode != http.StatusOK {
		logger.GlobalLogger.Warnf("[curl] 收到非200状态码: %d，尝试使用浏览器User-Agent重试", resp.StatusCode)
		// 尝试使用浏览器User-Agent重试
		req2, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			errMsg := fmt.Errorf("创建重试请求失败: %v", err)
			logger.GlobalLogger.Errorf("[curl] %v", errMsg)
			return "", common.NewNetworkError(url, errMsg)
		}

		// 设置浏览器User-Agent和基本请求头
		req2.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		req2.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req2.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req2.Header.Set("Accept-Encoding", "gzip, deflate")
		req2.Header.Set("Connection", "keep-alive")

		// 对于特定网站添加额外的请求头
		if strings.Contains(url, "sourceforge.net") {
			req2.Header.Set("Referer", "https://sourceforge.net/")
		}

		resp2, err := c.client.Do(req2)
		if err != nil {
			// 检查是否为超时错误
			if err == context.DeadlineExceeded {
				logger.GlobalLogger.Errorf("[curl] 重试请求超时: %v", err)
				return "", common.NewTimeoutError(url)
			}
			errMsg := fmt.Errorf("重试请求失败: %v", err)
			logger.GlobalLogger.Errorf("[curl] %v", errMsg)
			return "", common.NewNetworkError(url, errMsg)
		}
		defer resp2.Body.Close()

		if resp2.StatusCode != http.StatusOK {
			// 根据状态码返回不同的错误类型
			switch resp2.StatusCode {
			case http.StatusNotFound:
				return "", common.NewNotFoundError(url)
			case http.StatusForbidden:
				return "", common.NewPermissionError(url)
			case http.StatusUnauthorized:
				return "", common.NewPermissionError(url)
			case http.StatusTooManyRequests:
				errMsg := fmt.Errorf("请求过于频繁，状态码: %d", resp2.StatusCode)
				logger.GlobalLogger.Errorf("[curl] %v", errMsg)
				return "", common.NewNetworkError(url, errMsg)
			default:
				errMsg := fmt.Errorf("请求失败，状态码: %d", resp2.StatusCode)
				logger.GlobalLogger.Errorf("[curl] %v", errMsg)
				return "", common.NewNetworkError(url, errMsg)
			}
		}

		var body []byte
		var reader io.Reader = resp2.Body

		// 检查响应是否使用了gzip压缩
		if resp2.Header.Get("Content-Encoding") == "gzip" {
			logger.GlobalLogger.Debugf("[curl] 重试响应使用gzip压缩，正在解压...")
			gzipReader, err := gzip.NewReader(resp2.Body)
			if err != nil {
				errMsg := fmt.Errorf("创建gzip读取器失败: %v", err)
				logger.GlobalLogger.Errorf("[curl] %v", errMsg)
				return "", common.NewParseError(url, errMsg)
			}
			defer gzipReader.Close()
			reader = gzipReader
		}

		body, err = io.ReadAll(reader)
		if err != nil {
			errMsg := fmt.Errorf("读取响应体失败: %v", err)
			return "", common.NewParseError(url, errMsg)
		}

		content := string(body)
		logger.GlobalLogger.Debugf("[curl] 成功读取重试响应体，内容长度: %d 字符", len(content))
		return content, nil
	}

	var body []byte
	var reader io.Reader = resp.Body

	// 检查响应是否使用了gzip压缩
	if resp.Header.Get("Content-Encoding") == "gzip" {
		logger.GlobalLogger.Debugf("[curl] 响应使用gzip压缩，正在解压...")
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			logger.GlobalLogger.Errorf("[curl] 创建gzip读取器失败: %v", err)
			return "", common.NewParseError(url, fmt.Errorf("创建gzip读取器失败: %v", err))
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err = io.ReadAll(reader)
	if err != nil {
		errMsg := fmt.Errorf("读取响应体失败: %v", err)
		logger.GlobalLogger.Errorf("[curl] %v", errMsg)
		return "", common.NewParseError(url, errMsg)
	}

	content := string(body)
	logger.GlobalLogger.Debugf("[curl] 成功读取响应体，内容长度: %d 字符", len(content))

	// 保存页面内容到文件，以便调试
	// 从URL中提取一个简单的文件名
	urlParts := strings.Split(url, "/")
	lastPart := urlParts[len(urlParts)-1]
	if lastPart == "" {
		lastPart = urlParts[len(urlParts)-2]
	}
	// 移除特殊字符
	safeFileName := regexp.MustCompile(`[^\w\-.]`).ReplaceAllString(lastPart, "_")
	if safeFileName == "" {
		safeFileName = "page_content"
	}
	filePath := fmt.Sprintf("/tmp/%s.html", safeFileName)

	err = os.WriteFile(filePath, body, 0644)
	if err != nil {
		logger.GlobalLogger.Warnf("[curl] 无法保存页面内容到文件 %s: %v", filePath, err)
	} else {
		logger.GlobalLogger.Infof("[curl] 页面内容已保存到文件: %s", filePath)
	}

	return content, nil
}
