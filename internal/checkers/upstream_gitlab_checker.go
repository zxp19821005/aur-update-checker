package checkers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// GitLabRelease GitLab发布信息
type GitLabRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	WebUrl  string `json:"web_url"`
}

// GitLabChecker GitLab检查器
type GitLabChecker struct {
	*BaseGitPlatformChecker
}

// NewGitLabChecker 创建GitLab检查器
func NewGitLabChecker() *GitLabChecker {
	checker := &GitLabChecker{}
	checker.BaseGitPlatformChecker = NewBaseGitPlatformChecker(checker)
	return checker
}

// GetPlatformName 实现GitPlatformChecker接口，返回平台名称
func (c *GitLabChecker) GetPlatformName() string {
	return "gitlab"
}

// ParsePlatformURL 实现GitPlatformChecker接口，解析GitLab URL获取owner和repo
// 注意：GitLab需要host信息，所以我们只返回owner和repo，host信息在其他方法中处理
func (c *GitLabChecker) ParsePlatformURL(url string) (string, string, error) {
	// 匹配GitLab URL格式，支持自建GitLab实例
	re := regexp.MustCompile(`([^/]+)/([^/]+)/([^/]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 4 {
		return "", "", fmt.Errorf("无效的GitLab URL格式")
	}

	owner := matches[2]
	repo := matches[3]
	// 移除.git后缀
	repo = strings.TrimSuffix(repo, ".git")

	return owner, repo, nil
}

// GetLatestReleaseAPIURL 实现GitPlatformChecker接口，返回GitLab最新发布版本的API URL
// 注意：GitLab需要host信息，所以我们从完整URL中提取
func (c *GitLabChecker) GetLatestReleaseAPIURL(owner, repo string) string {
	// 由于GitLab API需要host信息，这里返回一个空字符串
	// 实际的URL在CheckWithVersionRef方法中构建
	return ""
}

// GetLatestTagsAPIURL 实现GitPlatformChecker接口，返回GitLab最新标签的API URL
// 注意：GitLab需要host信息，所以我们从完整URL中提取
func (c *GitLabChecker) GetLatestTagsAPIURL(owner, repo string) string {
	// 由于GitLab API需要host信息，这里返回一个空字符串
	// 实际的URL在CheckWithVersionRef方法中构建
	return ""
}

// SetRequestHeaders 实现GitPlatformChecker接口，设置GitLab API需要的请求头
func (c *GitLabChecker) SetRequestHeaders(req *http.Request) {
	// GitLab API不需要特殊的请求头
}

// Supports 实现GitPlatformChecker接口，检查此检查器是否支持给定的URL
func (c *GitLabChecker) Supports(url string) bool {
	// 使用正则表达式检查URL是否匹配GitLab格式
	// 支持官方gitlab.com和自建的GitLab实例
	re := regexp.MustCompile(`([^/]+)/([^/]+)/([^/]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 4 {
		return false
	}

	// 确保URL中包含gitlab关键字，避免误判
	return strings.Contains(strings.ToLower(url), "gitlab")
}

// GetPriority 实现GitPlatformChecker接口，返回检查器的优先级
func (c *GitLabChecker) GetPriority() int {
	// GitLab也是一个常用的代码托管平台，优先级略低于GitHub
	return 70 // 0-100范围，70为中等偏上优先级
}

// CheckWithVersionRef 重写基类方法，实现GitLab特定的版本引用检查逻辑
func (c *GitLabChecker) CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	// 解析GitLab URL获取host, owner和repo
	host, owner, repo, err := c.parseGitLabURL(url)
	if err != nil {
		return "", fmt.Errorf("解析GitLab URL失败: %v", err)
	}

	// 通过API获取latest release
	version, err := c.getLatestReleaseWithVersionRef(ctx, host, owner, repo, versionExtractKey, versionRef, checkTestVersion)
	if err != nil {
		return "", fmt.Errorf("获取GitLab最新发布失败: %v", err)
	}

	return version, nil
}

// parseGitLabURL 解析GitLab URL获取host, owner和repo
func (c *GitLabChecker) parseGitLabURL(url string) (string, string, string, error) {
	// 匹配GitLab URL格式，支持自建GitLab实例
	re := regexp.MustCompile(`([^/]+)/([^/]+)/([^/]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 4 {
		return "", "", "", fmt.Errorf("无效的GitLab URL格式")
	}

	// 提取host部分
	hostStart := strings.Index(url, "://") + 3
	if hostStart <= 2 {
		return "", "", "", fmt.Errorf("无效的URL格式")
	}
	hostEnd := strings.Index(url[hostStart:], "/")
	if hostEnd == -1 {
		return "", "", "", fmt.Errorf("无效的URL格式")
	}
	host := url[:hostStart+hostEnd]

	owner := matches[2]
	repo := matches[3]
	// 移除.git后缀
	repo = strings.TrimSuffix(repo, ".git")

	return host, owner, repo, nil
}

// getLatestReleaseWithVersionRef 根据版本引用获取发布版本
func (c *GitLabChecker) getLatestReleaseWithVersionRef(ctx context.Context, host, owner, repo, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	// GitLab API URL格式
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s%%2F%s/releases", host, owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置GitLab API需要的请求头
	c.SetRequestHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API请求失败，状态码: %d", resp.StatusCode)
	}

	var releases []GitLabRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(releases) == 0 {
		return "", fmt.Errorf("未找到任何发布")
	}

	var targetRelease *GitLabRelease

	// 如果提供了版本引用，尝试查找匹配的发布
	if versionRef != "" {
		for _, release := range releases {
			if release.TagName == versionRef || release.Name == versionRef {
				targetRelease = &release
				break
			}
		}

		// 如果没有找到匹配的发布，使用最新发布
		if targetRelease == nil {
			targetRelease = &releases[0]
		}
	} else {
		// 如果没有提供版本引用，使用最新发布
		targetRelease = &releases[0]
	}

	latestTag := targetRelease.TagName

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
