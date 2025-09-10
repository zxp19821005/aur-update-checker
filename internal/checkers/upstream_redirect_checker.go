package checkers

import (
	"aur-update-checker/internal/logger"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	checkerInterfaces "aur-update-checker/internal/interfaces/checkers"
	versionProcessor "aur-update-checker/internal/checkers/version"
)

// RedirectChecker 重定向检查器
type RedirectChecker struct {
	*checkerInterfaces.BaseChecker
	client *http.Client
}

// NewRedirectChecker 创建重定向检查器
func NewRedirectChecker() *RedirectChecker {
	return &RedirectChecker{
		BaseChecker: checkerInterfaces.NewBaseChecker("redirect"),
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // 不自动跟随重定向
			},
		},
	}
}

// Check 实现检查器接口，通过重定向URL获取版本号
func (c *RedirectChecker) Check(ctx context.Context, url, versionExtractKey string) (string, error) {
	// 默认不检查测试版本
	return c.CheckWithOption(ctx, url, versionExtractKey, 0)
}

// CheckWithOption 实现检查器接口，根据选项通过重定向URL获取版本号
func (c *RedirectChecker) CheckWithOption(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error) {
	logger.GlobalLogger.Debugf("[%s] 开始检查重定向URL: %s, 提取规则: %s", c.BaseChecker.Name(), url, versionExtractKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logger.GlobalLogger.Errorf("[%s] 创建请求失败: %v", c.BaseChecker.Name(), err)
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置常见的User-Agent和Accept头，以避免406错误
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := c.client.Do(req)
	if err != nil {
		logger.GlobalLogger.Errorf("[%s] 请求失败: %v", c.BaseChecker.Name(), err)
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	logger.GlobalLogger.Debugf("[%s] 收到响应，状态码: %d", c.BaseChecker.Name(), resp.StatusCode)

	// 检查是否是重定向
	if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")
		logger.GlobalLogger.Debugf("[%s] 检测到重定向，目标URL: %s", c.BaseChecker.Name(), location)

		if location == "" {
			logger.GlobalLogger.Errorf("[%s] 重定向响应中没有Location头", c.BaseChecker.Name())
			return "", fmt.Errorf("重定向响应中没有Location头")
		}

		// 从重定向URL中提取版本号
		version, err := c.extractVersionFromURLWithOption(location, versionExtractKey, checkTestVersion)
		if err != nil {
			logger.GlobalLogger.Errorf("[%s] 从重定向URL提取版本号失败: %v", c.BaseChecker.Name(), err)
			return "", err
		}

		// 规范化版本号，移除平台特定信息
		normalizedVersion := c.BaseChecker.NormalizeVersionWithOption(version, checkTestVersion)
		logger.GlobalLogger.Infof("[%s] 成功提取版本号: %s -> %s", c.BaseChecker.Name(), version, normalizedVersion)
		return normalizedVersion, nil
	}

	// 尝试处理其他状态码，比如406 Not Acceptable
	if resp.StatusCode == http.StatusNotAcceptable {
		logger.GlobalLogger.Debugf("[%s] 检测到406 Not Acceptable，尝试使用不同的Accept头重试", c.BaseChecker.Name())

		// 尝试不同的Accept头
		req2, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			logger.GlobalLogger.Errorf("[%s] 创建重试请求失败: %v", c.BaseChecker.Name(), err)
			return "", fmt.Errorf("创建重试请求失败: %v", err)
		}

		// 设置更宽松的Accept头
		req2.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		req2.Header.Set("Accept", "*/*")

		resp2, err := c.client.Do(req2)
		if err != nil {
			return "", fmt.Errorf("重试请求失败: %v", err)
		}
		defer resp2.Body.Close()

		logger.GlobalLogger.Debugf("[%s] 重试请求收到响应，状态码: %d", c.BaseChecker.Name(), resp2.StatusCode)

		// 检查是否是重定向
		if resp2.StatusCode == http.StatusMovedPermanently || resp2.StatusCode == http.StatusFound {
			location := resp2.Header.Get("Location")
			logger.GlobalLogger.Debugf("[%s] 重试请求检测到重定向，目标URL: %s", c.BaseChecker.Name(), location)

			if location == "" {
				return "", fmt.Errorf("重定向响应中没有Location头")
			}

			// 从重定向URL中提取版本号
			version, err := c.extractVersionFromURLWithOption(location, versionExtractKey, checkTestVersion)
			if err != nil {
				logger.GlobalLogger.Errorf("[%s] 从重定向URL提取版本号失败: %v", c.BaseChecker.Name(), err)
				return "", err
			}

			// 规范化版本号，移除平台特定信息
			normalizedVersion := c.BaseChecker.NormalizeVersionWithOption(version, checkTestVersion)
			logger.GlobalLogger.Infof("[%s] 成功提取版本号: %s -> %s", c.BaseChecker.Name(), version, normalizedVersion)
			return normalizedVersion, nil
		}

		return "", fmt.Errorf("URL未返回重定向响应，状态码: %d", resp2.StatusCode)
	}

	return "", fmt.Errorf("URL未返回重定向响应，状态码: %d", resp.StatusCode)
}

// extractVersionFromURLWithOption 根据选项从URL中提取版本号
func (c *RedirectChecker) extractVersionFromURLWithOption(url, key string, checkTestVersion int) (string, error) {
	logger.GlobalLogger.Debugf("[%s] 开始从URL提取版本号: %s, 提取规则: %s", c.BaseChecker.Name(), url, key)

	// 如果versionExtractKey为空，尝试从URL路径中提取版本号
	if key == "" {
		// 将URL按/分割成多个部分
		parts := strings.Split(url, "/")
		var candidateVersions []string

		logger.GlobalLogger.Debugf("[%s] URL分割成 %d 个部分", c.BaseChecker.Name(), len(parts))

		// 从每个部分中提取可能的版本号
		for i, part := range parts {
			logger.GlobalLogger.Debugf("[%s] 处理第 %d 部分: %s", c.BaseChecker.Name(), i+1, part)

			// 移除查询参数
			if strings.Contains(part, "?") {
				part = strings.Split(part, "?")[0]
				logger.GlobalLogger.Debugf("[%s] 移除查询参数后: %s", c.BaseChecker.Name(), part)
			}

			// 移除文件扩展名
			if strings.Contains(part, ".") {
				extensions := []string{".exe", ".msi", ".dmg", ".pkg", ".deb", ".rpm", ".tar.gz", ".zip", ".bin", ".php"}
				for _, ext := range extensions {
					if strings.HasSuffix(part, ext) {
						part = strings.TrimSuffix(part, ext)
						logger.GlobalLogger.Debugf("[%s] 移除扩展名 %s 后: %s", c.BaseChecker.Name(), ext, part)
						break
					}
				}
			}

			// 尝试使用正则表达式提取版本号
			versionPatterns := []string{
				`(\d+\.\d+\.\d+)`,      // X.X.X
				`v(\d+\.\d+\.\d+)`,     // vX.X.X
				`version-(\d+\.\d+\.\d+)`, // version-X.X.X
				`(\d+\.\d+)`,           // X.X
				`v(\d+\.\d+)`,          // vX.X
				`version-(\d+\.\d+)`,    // version-X.X
				`-(\d+\.\d+\.\d+)-`,    // -X.X.X-
				`-(\d+\.\d+)-`,         // -X.X-
			}

			for _, pattern := range versionPatterns {
				re := regexp.MustCompile(pattern)
				matches := re.FindStringSubmatch(part)
				if len(matches) > 1 {
					// 找到版本号，添加到候选列表
					candidateVersions = append(candidateVersions, matches[1])
					logger.GlobalLogger.Debugf("[%s] 使用模式 %s 找到版本号: %s", c.BaseChecker.Name(), pattern, matches[1])
				}
			}

			// 检查这部分本身是否是版本号
			if isVersionString(part) {
				candidateVersions = append(candidateVersions, part)
				logger.GlobalLogger.Debugf("[%s] 部分本身是版本号: %s", c.BaseChecker.Name(), part)
			}
		}

		// 如果没有找到任何候选版本号
		if len(candidateVersions) == 0 {
			logger.GlobalLogger.Errorf("[%s] 没有找到任何候选版本号", c.BaseChecker.Name())
			return "", fmt.Errorf("无法从URL中提取版本号")
		}

		logger.GlobalLogger.Debugf("[%s] 找到 %d 个候选版本号: %v", c.BaseChecker.Name(), len(candidateVersions), candidateVersions)

		// 使用UpstreamVersionParser比较版本号，选择最新的一个
		parser := versionProcessor.NewUpstreamVersionParser()
		latestVersion := candidateVersions[0]

		for _, version := range candidateVersions[1:] {
			// 比较两个版本，选择更新的一个
			_, isSimilar := parser.ParseAndCompare(version, latestVersion)
			if isSimilar {
				latestVersion = version
				logger.GlobalLogger.Debugf("[%s] 选择更新的版本号: %s", c.BaseChecker.Name(), version)
			}
		}

		logger.GlobalLogger.Infof("[%s] 最终选择的版本号: %s", c.BaseChecker.Name(), latestVersion)

		// 规范化版本号
		return c.BaseChecker.NormalizeVersionWithOption(latestVersion, checkTestVersion), nil
	}

	// 如果提供了versionExtractKey，尝试从URL中提取匹配的部分
	if strings.Contains(url, key) {
		// 尝试从key后提取版本号
		index := strings.Index(url, key)
		if index+len(key) < len(url) {
			// 获取key后面的内容
			afterKey := url[index+len(key):]

			// 如果后面跟着非字母数字字符，跳过它
			if len(afterKey) > 0 && !isAlphanumeric(rune(afterKey[0])) {
				afterKey = afterKey[1:]
			}

			// 提取版本号部分，直到遇到非版本号字符
			version := ""
			for _, char := range afterKey {
				if isVersionChar(char) {
					version += string(char)
				} else {
					break
				}
			}

			if version != "" {
				// 规范化版本号，移除平台特定信息
				return c.BaseChecker.NormalizeVersionWithOption(version, checkTestVersion), nil
			}
		}
	}

	return "", fmt.Errorf("无法从URL中提取版本号")
}

// isAlphanumeric 检查字符是否是字母或数字
func isAlphanumeric(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')
}

// isVersionChar 检查字符是否可能是版本号的一部分
func isVersionChar(char rune) bool {
	return isAlphanumeric(char) || char == '.' || char == '-' || char == '_' || char == '+'
}

// isVersionString 检查字符串是否可能是版本号
func isVersionString(s string) bool {
	// 简单的版本号格式检查：X.X.X 或 X.X
	re := regexp.MustCompile(`^v?\d+\.\d+(\.\d+)?([a-zA-Z]+\d*)?$`)
	return re.MatchString(s)
}

// CheckWithVersionRef 带选项和版本引用地检查上游版本
func (c *RedirectChecker) CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	// 简单地调用CheckWithOption方法，忽略versionRef参数
	return c.CheckWithOption(ctx, url, versionExtractKey, checkTestVersion)
}
