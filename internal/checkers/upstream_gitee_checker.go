package checkers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// GiteeChecker Gitee检查器
type GiteeChecker struct {
	*BaseGitPlatformChecker
}

// NewGiteeChecker 创建Gitee检查器
func NewGiteeChecker() *GiteeChecker {
	checker := &GiteeChecker{}
	checker.BaseGitPlatformChecker = NewBaseGitPlatformChecker(checker)
	return checker
}

// GetPlatformName 实现GitPlatformChecker接口，返回平台名称
func (c *GiteeChecker) GetPlatformName() string {
	return "gitee"
}

// ParsePlatformURL 实现GitPlatformChecker接口，解析Gitee URL获取owner和repo
func (c *GiteeChecker) ParsePlatformURL(url string) (string, string, error) {
	// 匹配Gitee URL格式
	re := regexp.MustCompile(`gitee\.com/([^/]+)/([^/]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("无效的Gitee URL格式")
	}

	owner := matches[1]
	repo := matches[2]
	// 移除.git后缀
	repo = strings.TrimSuffix(repo, ".git")

	return owner, repo, nil
}

// GetLatestReleaseAPIURL 实现GitPlatformChecker接口，返回Gitee最新发布版本的API URL
func (c *GiteeChecker) GetLatestReleaseAPIURL(owner, repo string) string {
	return fmt.Sprintf("https://gitee.com/api/v5/repos/%s/%s/releases/latest", owner, repo)
}

// GetLatestTagsAPIURL 实现GitPlatformChecker接口，返回Gitee最新标签的API URL
func (c *GiteeChecker) GetLatestTagsAPIURL(owner, repo string) string {
	return fmt.Sprintf("https://gitee.com/api/v5/repos/%s/%s/tags", owner, repo)
}

// SetRequestHeaders 实现GitPlatformChecker接口，设置Gitee API需要的请求头
func (c *GiteeChecker) SetRequestHeaders(req *http.Request) {
	// Gitee API不需要特殊的请求头
}

// Supports 实现GitPlatformChecker接口，检查此检查器是否支持给定的URL
func (c *GiteeChecker) Supports(url string) bool {
	// 使用正则表达式检查URL是否匹配Gitee格式
	re := regexp.MustCompile(`gitee\.com/([^/]+)/([^/]+)`)
	return re.MatchString(url)
}

// GetPriority 实现GitPlatformChecker接口，返回检查器的优先级
func (c *GiteeChecker) GetPriority() int {
	// Gitee是一个常用的代码托管平台，给予中等优先级
	return 70 // 0-100范围，70为中等优先级
}