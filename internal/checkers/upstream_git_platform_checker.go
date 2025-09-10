package checkers

import (
	"aur-update-checker/internal/logger"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	checkerInterfaces "aur-update-checker/internal/interfaces/checkers"
)

// GitPlatformRelease Git平台发布信息的通用结构
type GitPlatformRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HtmlUrl string `json:"html_url"`
	WebUrl  string `json:"web_url"` // GitLab使用web_url而不是html_url
}

// GitPlatformChecker Git平台检查器接口，定义了各平台需要实现的特定方法
type GitPlatformChecker interface {
	// 获取平台名称
	GetPlatformName() string
	// 解析平台URL，返回owner和repo
	ParsePlatformURL(url string) (string, string, error)
	// 获取最新发布版本的API URL
	GetLatestReleaseAPIURL(owner, repo string) string
	// 获取最新标签的API URL
	GetLatestTagsAPIURL(owner, repo string) string
	// 设置HTTP请求的特定头信息
	SetRequestHeaders(req *http.Request)
	// 检查是否支持给定的URL
	Supports(url string) bool
	// 获取检查器优先级
	GetPriority() int
}

// BaseGitPlatformChecker Git平台检查器的基类，包含共同的方法和逻辑
type BaseGitPlatformChecker struct {
	*checkerInterfaces.BaseChecker
	client         *http.Client
	platformChecker GitPlatformChecker
}

// NewBaseGitPlatformChecker 创建Git平台检查器基类
func NewBaseGitPlatformChecker(platformChecker GitPlatformChecker) *BaseGitPlatformChecker {
	return &BaseGitPlatformChecker{
		BaseChecker:     checkerInterfaces.NewBaseChecker(platformChecker.GetPlatformName()),
		client:          &http.Client{},
		platformChecker: platformChecker,
	}
}

// Check 实现检查器接口，检查Git平台项目版本
func (c *BaseGitPlatformChecker) Check(ctx context.Context, url, versionExtractKey string) (string, error) {
	// 默认不检查测试版本
	return c.CheckWithOption(ctx, url, versionExtractKey, 0)
}

// CheckWithOption 实现检查器接口，根据选项检查Git平台项目版本
func (c *BaseGitPlatformChecker) CheckWithOption(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error) {
	// 解析URL获取owner和repo
	owner, repo, err := c.platformChecker.ParsePlatformURL(url)
	if err != nil {
		platformName := c.platformChecker.GetPlatformName()
		errMsg := fmt.Errorf("解析%s URL失败: %v", platformName, err)
		logger.GlobalLogger.Errorf("[%s] %v", platformName, errMsg)
		return "", errMsg
	}

	// 方法1: 通过API获取latest release
	version, err := c.getLatestReleaseWithOption(ctx, owner, repo, versionExtractKey, checkTestVersion)
	if err == nil && version != "" {
		return version, nil
	}

	// 方法2: 如果获取的latest release失败，则通过API获取latest tag
	version, err = c.getLatestTagWithOption(ctx, owner, repo, versionExtractKey, checkTestVersion)
	if err == nil && version != "" {
		return version, nil
	}

	// 所有检查方法均失败
	platformName := c.platformChecker.GetPlatformName()
	logger.GlobalLogger.Errorf("[%s] 所有%s检查方法均失败", platformName, platformName)
	return "", fmt.Errorf("所有%s检查方法均失败", platformName)
}

// CheckWithVersionRef 带选项和版本引用地检查上游版本
func (c *BaseGitPlatformChecker) CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	// 简单地调用CheckWithOption方法，忽略versionRef参数
	return c.CheckWithOption(ctx, url, versionExtractKey, checkTestVersion)
}

// Supports 检查此检查器是否支持给定的URL
func (c *BaseGitPlatformChecker) Supports(url string) bool {
	return c.platformChecker.Supports(url)
}

// Priority 返回检查器的优先级
func (c *BaseGitPlatformChecker) Priority() int {
	return c.platformChecker.GetPriority()
}

// getLatestReleaseWithOption 根据选项获取最新发布版本
func (c *BaseGitPlatformChecker) getLatestReleaseWithOption(ctx context.Context, owner, repo, versionExtractKey string, checkTestVersion int) (string, error) {
	platformName := c.platformChecker.GetPlatformName()
	apiURL := c.platformChecker.GetLatestReleaseAPIURL(owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		errMsg := fmt.Errorf("创建请求失败: %v", err)
		logger.GlobalLogger.Errorf("[%s] %v", platformName, errMsg)
		return "", errMsg
	}

	// 设置平台特定的请求头
	c.platformChecker.SetRequestHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API请求失败，状态码: %d", resp.StatusCode)
	}

	var release GitPlatformRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		errMsg := fmt.Errorf("解析响应失败: %v", err)
		logger.GlobalLogger.Errorf("[%s] %v", platformName, errMsg)
		return "", errMsg
	}

	// 如果versionExtractKey为空，直接返回tag_name并进行规范化
	if versionExtractKey == "" {
		return c.BaseChecker.NormalizeVersionWithOption(release.TagName, checkTestVersion), nil
	}

	// 使用versionExtractKey提取版本
	version, err := c.BaseChecker.ExtractVersionFromContent(release.TagName, versionExtractKey)
	if err != nil {
		return "", err
	}
	return c.BaseChecker.NormalizeVersionWithOption(version, checkTestVersion), nil
}

// getLatestTagWithOption 根据选项获取最新标签
func (c *BaseGitPlatformChecker) getLatestTagWithOption(ctx context.Context, owner, repo, versionExtractKey string, checkTestVersion int) (string, error) {
	platformName := c.platformChecker.GetPlatformName()
	apiURL := c.platformChecker.GetLatestTagsAPIURL(owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		errMsg := fmt.Errorf("创建请求失败: %v", err)
		logger.GlobalLogger.Errorf("[%s] %v", platformName, errMsg)
		return "", errMsg
	}

	// 设置平台特定的请求头
	c.platformChecker.SetRequestHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API请求失败，状态码: %d", resp.StatusCode)
	}

	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		errMsg := fmt.Errorf("解析响应失败: %v", err)
		logger.GlobalLogger.Errorf("[%s] %v", platformName, errMsg)
		return "", errMsg
	}

	if len(tags) == 0 {
		errMsg := fmt.Errorf("未找到任何标签")
		logger.GlobalLogger.Errorf("[%s] %v", platformName, errMsg)
		return "", errMsg
	}

	// 获取第一个标签（最新的）
	latestTag := tags[0].Name

	// 如果versionExtractKey为空，直接返回标签名并进行规范化
	if versionExtractKey == "" {
		return c.BaseChecker.NormalizeVersionWithOption(latestTag, checkTestVersion), nil
	}

	// 使用versionExtractKey提取版本
	version, err := c.BaseChecker.ExtractVersionFromContent(latestTag, versionExtractKey)
	if err != nil {
		return "", err
	}
	return c.BaseChecker.NormalizeVersionWithOption(version, checkTestVersion), nil
}