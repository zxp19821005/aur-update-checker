package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"aur-update-checker/internal/interfaces/checkers"
)

// GitLabCheckerExample GitLab检查器示例
type GitLabCheckerExample struct {
	base *checkers.BaseChecker
}

// Name 返回检查器名称
func (c *GitLabCheckerExample) Name() string {
	return c.base.Name()
}

// Supports 检查是否支持给定的URL
func (c *GitLabCheckerExample) Supports(url string) bool {
	return strings.Contains(url, "gitlab.com")
}

// Priority 返回检查器优先级
func (c *GitLabCheckerExample) Priority() int {
	return 80 // GitLab检查器优先级较高
}

// NewGitLabCheckerExample 创建GitLab检查器示例
func NewGitLabCheckerExample() *GitLabCheckerExample {
	base := checkers.NewBaseChecker("gitlab_example")
	return &GitLabCheckerExample{base: base}
}

// Check 实现检查方法
func (c *GitLabCheckerExample) Check(ctx context.Context, url, versionExtractKey string) (string, error) {
	return c.CheckWithOption(ctx, url, versionExtractKey, 0)
}

// CheckWithOption 实现带选项的检查方法
func (c *GitLabCheckerExample) CheckWithOption(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error) {
	// 调用CheckWithVersionRef方法，传入空版本引用
	return c.CheckWithVersionRef(ctx, url, versionExtractKey, "", checkTestVersion)
}

// CheckWithVersionRef 实现带版本引用的检查方法
func (c *GitLabCheckerExample) CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	// 这里实现GitLab特定的版本检查逻辑
	// 示例实现，实际使用时需要根据GitLab API进行修改

	// 检查URL是否为GitLab仓库URL
	if !c.Supports(url) {
		return "", fmt.Errorf("不支持的URL格式: %s", url)
	}

	// 从URL中提取项目路径
	re := regexp.MustCompile(`gitlab\.com/([^/]+/[^/]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("无法从URL中提取项目路径: %s", url)
	}

	// 在实际实现中，应该使用projectPath调用GitLab API
	// 示例代码仅记录projectPath，避免未使用变量错误
	projectPath := matches[1]
	fmt.Printf("提取到项目路径: %s\n", projectPath)

	// 如果提供了版本引用，可以使用它来获取特定版本
	if versionRef != "" {
		fmt.Printf("使用版本引用: %s\n", versionRef)
		// 在实际实现中，应该使用versionRef调用GitLab API获取特定版本
		// 示例代码仅返回一个模拟版本号
		latestVersion := versionRef
		// 规范化版本号
		normalizedVersion := c.base.NormalizeVersionWithOption(latestVersion, checkTestVersion)
		return normalizedVersion, nil
	}

	// 这里应该调用GitLab API获取最新版本
	// 示例代码仅返回一个模拟版本号
	latestVersion := "v1.0.0"

	// 规范化版本号
	normalizedVersion := c.base.NormalizeVersionWithOption(latestVersion, checkTestVersion)

	return normalizedVersion, nil
}

// PluginInfo 返回插件信息
func (c *GitLabCheckerExample) PluginInfo() checkers.PluginInfo {
	return checkers.PluginInfo{
		Name:        "gitlab_example",
		Version:     "1.0.0",
		Author:      "Plugin Developer",
		Description: "GitLab仓库版本检查器示例插件",
	}
}

// NewPluginChecker 插件入口函数
// 插件系统会查找此函数来创建检查器实例
func NewPluginChecker() checkers.PluginChecker {
	return NewGitLabCheckerExample()
}

func main() {
	// 插件的主函数，可以留空
	// 实际的插件逻辑在NewPluginChecker函数中
}
