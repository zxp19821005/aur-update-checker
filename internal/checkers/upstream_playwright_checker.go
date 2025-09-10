package checkers

import (
	"aur-update-checker/internal/checkers/common"
	"aur-update-checker/internal/logger"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"

	checkerInterfaces "aur-update-checker/internal/interfaces/checkers"
)

// PlaywrightChecker 使用Playwright进行浏览器自动化，检查上游版本
type PlaywrightChecker struct {
	*checkerInterfaces.BaseChecker
	timeout  time.Duration
	headless bool
}

// NewPlaywrightChecker 创建一个新的Playwright检查器
func NewPlaywrightChecker() *PlaywrightChecker {
	return &PlaywrightChecker{
		BaseChecker: checkerInterfaces.NewBaseChecker("playwright"),
		timeout:     30 * time.Second,
		headless:    true,
	}
}

// SetTimeout 设置超时时间
func (c *PlaywrightChecker) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// SetHeadless 设置是否使用无头模式
func (c *PlaywrightChecker) SetHeadless(headless bool) {
	c.headless = headless
}

// Name 方法由BaseChecker提供

// Supports 检查此检查器是否支持给定的URL
func (c *PlaywrightChecker) Supports(url string) bool {
	// 对于所有网站，都可以使用Playwright检查器
	return true
}

// Priority 返回检查器的优先级
func (c *PlaywrightChecker) Priority() int {
	// 对于南天电子网站，返回高优先级
	// 对于其他网站，返回中等优先级
	return 70 // 0-100范围，70为较高优先级
}

// Check 检查上游版本
func (c *PlaywrightChecker) Check(ctx context.Context, url, versionExtractKey string) (string, error) {
	return c.CheckWithOption(ctx, url, versionExtractKey, 0)
}

// CheckWithOption 带选项地检查上游版本
func (c *PlaywrightChecker) CheckWithOption(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error) {
	return c.CheckWithVersionRef(ctx, url, versionExtractKey, "", checkTestVersion)
}

// CheckWithVersionRef 带选项和版本引用地检查上游版本
func (c *PlaywrightChecker) CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	logger.GlobalLogger.Infof("[Playwright检查器] 开始检查上游版本 - URL: %s, 提取键: %s, 版本引用: %s", url, versionExtractKey, versionRef)

	if versionExtractKey == "" {
		logger.GlobalLogger.Errorf("[Playwright检查器] versionExtractKey为空")
		return "", fmt.Errorf("Playwright检查器需要提供versionExtractKey来定位版本信息")
	}

	// 初始化Playwright
	// 设置环境变量，使用官方Playwright下载源
	logger.GlobalLogger.Debugf("[Playwright检查器] 设置环境变量，使用官方下载源")
	os.Setenv("PLAYWRIGHT_DOWNLOAD_HOST", "https://playwright.azureedge.net")
	// 禁用依赖检查
	logger.GlobalLogger.Debugf("[Playwright检查器] 禁用依赖检查")
	os.Setenv("PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS", "true")

	err := playwright.Install()
	if err != nil {
		logger.GlobalLogger.Errorf("[Playwright检查器] 安装Playwright失败: %v", err)
		return "", fmt.Errorf("安装Playwright失败: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		logger.GlobalLogger.Errorf("[Playwright检查器] 启动Playwright失败: %v", err)
		return "", fmt.Errorf("启动Playwright失败: %v", err)
	}
	defer pw.Stop()

	// 启动浏览器
	logger.GlobalLogger.Debugf("[Playwright检查器] 启动浏览器")
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(c.headless),
	})
	if err != nil {
		logger.GlobalLogger.Errorf("[Playwright检查器] 启动浏览器失败: %v", err)
		return "", fmt.Errorf("启动浏览器失败: %v", err)
	}
	defer browser.Close()

	// 创建页面上下文和页面
	context, err := browser.NewContext()
	if err != nil {
		logger.GlobalLogger.Errorf("[Playwright检查器] 创建浏览器上下文失败: %v", err)
		return "", fmt.Errorf("创建浏览器上下文失败: %v", err)
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		logger.GlobalLogger.Errorf("[Playwright检查器] 创建页面失败: %v", err)
		return "", fmt.Errorf("创建页面失败: %v", err)
	}

	// 设置超时
	context.SetDefaultTimeout(float64(c.timeout.Milliseconds()))

	// 导航到URL
	logger.GlobalLogger.Debugf("[Playwright检查器] 导航到URL: %s", url)
	_, err = page.Goto(url)
	if err != nil {
		logger.GlobalLogger.Errorf("[Playwright检查器] 导航失败: %v", err)
		return "", fmt.Errorf("导航失败: %v", err)
	}

	// 等待页面加载完成
	logger.GlobalLogger.Debugf("[Playwright检查器] 等待页面加载")
	err = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateNetworkidle})
	if err != nil {
		logger.GlobalLogger.Warnf("[Playwright检查器] 等待页面加载失败: %v", err)
	}

	// 不再对特定网站进行特殊处理，统一使用通用逻辑

	// 获取页面内容
	logger.GlobalLogger.Debugf("[Playwright检查器] 获取页面内容")
	content, err := page.Content()
	if err != nil {
		logger.GlobalLogger.Errorf("[Playwright检查器] 获取页面内容失败: %v", err)
		return "", fmt.Errorf("获取页面内容失败: %v", err)
	}

	// 提取版本
	logger.GlobalLogger.Debugf("[Playwright检查器] 开始提取版本，提取键: %s", versionExtractKey)
	version, err := c.extractVersion(content, versionExtractKey)
	if err != nil {
		logger.GlobalLogger.Errorf("[Playwright检查器] 从页面内容提取版本失败: %v", err)
		return "", fmt.Errorf("从页面内容提取版本失败: %v", err)
	}
	logger.GlobalLogger.Infof("[Playwright检查器] 成功提取版本: %s", version)
	
	// 如果提供了版本引用，尝试使用它来更精确地提取版本
	if versionRef != "" {
		logger.GlobalLogger.Debugf("[Playwright检查器] 使用版本引用 %s 来优化版本提取", versionRef)
		// 检查提取的版本是否与版本引用匹配
		if !strings.Contains(version, versionRef) {
			// 如果不匹配，尝试在页面内容中查找包含版本引用的版本
			refVersion, err := c.extractVersionWithRef(content, versionExtractKey, versionRef)
			if err == nil && refVersion != "" {
				version = refVersion
				logger.GlobalLogger.Infof("[Playwright检查器] 使用版本引用提取到更精确的版本: %s", version)
			}
		}
	}

	// 规范化版本号
	normalizedVersion := c.BaseChecker.NormalizeVersionWithOption(version, checkTestVersion)
	logger.GlobalLogger.Infof("[Playwright检查器] 版本规范化完成，最终版本: %s", normalizedVersion)
	return normalizedVersion, nil
}

// extractVersion 从内容中提取版本号
func (c *PlaywrightChecker) extractVersion(content, key string) (string, error) {
	// 特殊处理复合键，如 "Linux&信创"
	if strings.Contains(key, "&") {
		keys := strings.Split(key, "&")
		if len(keys) >= 2 {
			// 查找所有键的组合
			results := c.findCombinedKeys(content, keys)
			if len(results) > 0 {
				return c.extractVersionFromString(results[0]), nil
			}
		}
	}

	// 检查key是否是正则表达式格式（包含捕获组）
	if strings.Contains(key, `(`) && strings.Contains(key, `)`) {
		re, err := regexp.Compile(key)
		if err != nil {
			logger.GlobalLogger.Debugf("[Playwright检查器] 正则表达式编译失败: %v", err)
		} else {
			matches := re.FindStringSubmatch(content)
			if len(matches) >= 2 {
				logger.GlobalLogger.Debugf("[Playwright检查器] 使用正则表达式匹配到版本: %s", matches[1])
				return matches[1], nil
			}
		}
	}

	// 查找key在content中的所有位置
	index := 0
	for {
		pos := strings.Index(content[index:], key)
		if pos == -1 {
			break
		}

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
		version := c.extractVersionFromString(extract)
		if version != "" {
			return version, nil
		}

		// 移动到下一个位置继续搜索
		index = absPos + len(key)
	}

	return "", fmt.Errorf("未找到匹配版本提取键的内容")
}

// findCombinedKeys 查找多个键的组合
// 使用公共函数 common.FindCombinedKeys 实现
func (c *PlaywrightChecker) findCombinedKeys(content string, keys []string) []string {
	logger.GlobalLogger.Debugf("[Playwright检查器] 调用公共函数查找复合键")
	return common.FindCombinedKeys(content, keys)
}

// extractVersionWithRef 使用版本引用从内容中提取版本号
func (c *PlaywrightChecker) extractVersionWithRef(content, extractKey, versionRef string) (string, error) {
	// 构建复合键，将提取键和版本引用组合起来
	combinedKey := extractKey + "&" + versionRef
	logger.GlobalLogger.Debugf("[Playwright检查器] 使用复合键 %s 提取版本", combinedKey)
	
	// 使用复合键提取版本
	return c.extractVersion(content, combinedKey)
}

// extractVersionFromString 从字符串中提取版本号
// 使用公共函数 common.ExtractVersionFromString 实现
func (c *PlaywrightChecker) extractVersionFromString(s string) string {
	logger.GlobalLogger.Debugf("[Playwright检查器] 调用公共函数提取版本号")
	return common.ExtractVersionFromString(s)
}

// NormalizeVersionWithOption 规范化版本号，根据选项决定是否保留测试版本标识符
func (c *PlaywrightChecker) NormalizeVersionWithOption(version string, checkTestVersion int) string {
	// 调用BaseChecker的规范化方法
	return c.BaseChecker.NormalizeVersionWithOption(version, checkTestVersion)
}
