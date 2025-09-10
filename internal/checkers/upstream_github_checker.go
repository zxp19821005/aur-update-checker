package checkers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// GitHubChecker GitHub检查器
type GitHubChecker struct {
	*BaseGitPlatformChecker
}

// NewGitHubChecker 创建GitHub检查器
func NewGitHubChecker() *GitHubChecker {
	checker := &GitHubChecker{}
	checker.BaseGitPlatformChecker = NewBaseGitPlatformChecker(checker)
	return checker
}

// GetPlatformName 实现GitPlatformChecker接口，返回平台名称
func (c *GitHubChecker) GetPlatformName() string {
	return "github"
}

// ParsePlatformURL 实现GitPlatformChecker接口，解析GitHub URL获取owner和repo
func (c *GitHubChecker) ParsePlatformURL(url string) (string, string, error) {
	// 匹配GitHub URL格式
	re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("无效的GitHub URL格式")
	}

	owner := matches[1]
	repo := matches[2]
	// 移除.git后缀
	repo = strings.TrimSuffix(repo, ".git")

	return owner, repo, nil
}

// GetLatestReleaseAPIURL 实现GitPlatformChecker接口，返回GitHub最新发布版本的API URL
func (c *GitHubChecker) GetLatestReleaseAPIURL(owner, repo string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
}

// GetLatestTagsAPIURL 实现GitPlatformChecker接口，返回GitHub最新标签的API URL
func (c *GitHubChecker) GetLatestTagsAPIURL(owner, repo string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", owner, repo)
}

// SetRequestHeaders 实现GitPlatformChecker接口，设置GitHub API需要的请求头
func (c *GitHubChecker) SetRequestHeaders(req *http.Request) {
	// 设置GitHub API需要的User-Agent
	req.Header.Set("User-Agent", "aur-update-checker")
}

// Supports 实现GitPlatformChecker接口，检查此检查器是否支持给定的URL
func (c *GitHubChecker) Supports(url string) bool {
	// 使用正则表达式检查URL是否匹配GitHub格式
	re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)`)
	return re.MatchString(url)
}

// GetPriority 实现GitPlatformChecker接口，返回检查器的优先级
func (c *GitHubChecker) GetPriority() int {
	// GitHub是一个常用的代码托管平台，给予较高优先级
	return 80 // 0-100范围，80为较高优先级
}