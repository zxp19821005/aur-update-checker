package checkers

import (
	"aur-update-checker/internal/logger"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	checkerInterfaces "aur-update-checker/internal/interfaces/checkers"
)

// PyPIPackage PyPI包信息
type PyPIPackage struct {
	Info struct {
		Name             string `json:"name"`
		Version          string `json:"version"`
		HomePage         string `json:"home_page"`
		PackageURL       string `json:"package_url"`
		ReleaseURL       string `json:"release_url"`
	} `json:"info"`
	URLs []struct {
		PackageType string `json:"packagetype"`
		URL         string `json:"url"`
	} `json:"urls"`
	Releases map[string][]struct {
		PackageType string `json:"packagetype"`
		URL         string `json:"url"`
	} `json:"releases"`
}

// PyPIChecker PyPI检查器
type PyPIChecker struct {
	*checkerInterfaces.BaseChecker
	client *http.Client
}

// NewPyPIChecker 创建PyPI检查器
func NewPyPIChecker() *PyPIChecker {
	return &PyPIChecker{
		BaseChecker: checkerInterfaces.NewBaseChecker("pypi"),
		client:      &http.Client{},
	}
}

// Check 实现检查器接口，从PyPI获取包版本
func (c *PyPIChecker) Check(ctx context.Context, url, versionExtractKey string) (string, error) {
	// 默认不检查测试版本
	return c.CheckWithOption(ctx, url, versionExtractKey, 0)
}

// CheckWithOption 实现检查器接口，根据选项从PyPI获取包版本
func (c *PyPIChecker) CheckWithOption(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error) {
	// 调用CheckWithVersionRef方法，传入空版本引用
	return c.CheckWithVersionRef(ctx, url, versionExtractKey, "", checkTestVersion)
}

// CheckWithVersionRef 实现检查器接口，根据版本引用从PyPI获取包版本
func (c *PyPIChecker) CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	// 从URL或versionExtractKey中提取包名
	packageName, err := c.extractPackageName(url, versionExtractKey)
	if err != nil {
		logger.GlobalLogger.Errorf("[pypi] 提取PyPI包名失败: %v", err)
		return "", fmt.Errorf("提取PyPI包名失败: %v", err)
	}

	// 获取PyPI包信息
	packageInfo, err := c.fetchPackageInfo(ctx, packageName)
	if err != nil {
		logger.GlobalLogger.Errorf("[pypi] 获取PyPI包信息失败: %v", err)
		return "", fmt.Errorf("获取PyPI包信息失败: %v", err)
	}

	// 提取版本
	version, err := c.extractVersionWithVersionRef(packageInfo, versionExtractKey, versionRef, checkTestVersion)
	if err != nil {
		logger.GlobalLogger.Errorf("[pypi] 提取版本失败: %v", err)
		return "", fmt.Errorf("提取版本失败: %v", err)
	}

	// 规范化版本号，移除平台特定信息
	normalizedVersion := c.BaseChecker.NormalizeVersionWithOption(version, checkTestVersion)
	return normalizedVersion, nil
}

// extractPackageName 从URL或versionExtractKey中提取包名
func (c *PyPIChecker) extractPackageName(url, versionExtractKey string) (string, error) {
	// 优先尝试从versionExtractKey中提取包名
	if versionExtractKey != "" {
		// 如果versionExtractKey看起来像包名，直接使用
		if c.isValidPackageName(versionExtractKey) {
			return versionExtractKey, nil
		}
	}

	// 尝试从URL中提取包名
	// 匹配pypi.org/project/<package-name>/格式
	re := regexp.MustCompile(`pypi\.org/project/([^/\s]+)/`)
	matches := re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		return matches[1], nil
	}

	// 匹配pypi.org/project/<package-name>格式
	re = regexp.MustCompile(`pypi\.org/project/([^/\s]+)`)
	matches = re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		return matches[1], nil
	}

	// 匹配pypi.python.org/pypi/<package-name>格式
	re = regexp.MustCompile(`pypi\.python\.org/pypi/([^/\s]+)`)
	matches = re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		return matches[1], nil
	}

	logger.GlobalLogger.Errorf("[pypi] 无法从URL或versionExtractKey中提取有效的PyPI包名")
	return "", fmt.Errorf("无法从URL或versionExtractKey中提取有效的PyPI包名")
}

// isValidPackageName 检查是否是有效的PyPI包名
func (c *PyPIChecker) isValidPackageName(name string) bool {
	// 简单检查，实际PyPI包名规则更复杂
	if name == "" || name == "project" || name == "pypi" || name == "search" {
		return false
	}
	return true
}

// fetchPackageInfo 获取PyPI包信息
func (c *PyPIChecker) fetchPackageInfo(ctx context.Context, packageName string) (*PyPIPackage, error) {
	apiURL := fmt.Sprintf("https://pypi.org/pypi/%s/json", packageName)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		logger.GlobalLogger.Errorf("[pypi] 创建请求失败: %v", err)
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GlobalLogger.Errorf("[pypi] 读取响应体失败: %v", err)
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	var packageInfo PyPIPackage
	if err := json.Unmarshal(body, &packageInfo); err != nil {
		logger.GlobalLogger.Errorf("[pypi] 解析响应失败: %v", err)
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &packageInfo, nil
}

// extractVersionWithVersionRef 根据版本引用从包信息中提取版本
func (c *PyPIChecker) extractVersionWithVersionRef(packageInfo *PyPIPackage, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	// 如果提供了版本引用，尝试使用它
	if versionRef != "" {
		// 检查versionRef是否是有效的版本号
		if _, ok := packageInfo.Releases[versionRef]; ok {
			// 根据checkTestVersion参数决定是否检查测试版本
			if checkTestVersion > 0 {
				return versionRef, nil
			}
			// 如果不检查测试版本，需要过滤掉测试版本
			return c.BaseChecker.NormalizeVersion(versionRef), nil
		}
		
		// 如果versionRef不是有效的版本号，记录警告并使用最新版本
		logger.GlobalLogger.Warnf("[pypi] 版本引用 %s 不是有效的版本号，将使用最新版本", versionRef)
	}

	// 如果versionExtractKey为空，使用最新版本
	if versionExtractKey == "" {
		// 根据checkTestVersion参数决定是否检查测试版本
		if checkTestVersion > 0 {
			return packageInfo.Info.Version, nil
		}
		// 如果不检查测试版本，需要过滤掉测试版本
		return c.BaseChecker.NormalizeVersion(packageInfo.Info.Version), nil
	}

	// 检查versionExtractKey是否是版本号
	if _, ok := packageInfo.Releases[versionExtractKey]; ok {
		// 根据checkTestVersion参数决定是否检查测试版本
		if checkTestVersion > 0 {
			return versionExtractKey, nil
		}
		// 如果不检查测试版本，需要过滤掉测试版本
		return c.BaseChecker.NormalizeVersion(versionExtractKey), nil
	}

	// 如果都不是，尝试使用versionExtractKey作为提取规则
	version, err := c.BaseChecker.ExtractVersionFromContent(packageInfo.Info.Version, versionExtractKey)
	if err != nil {
		return "", err
	}
	// 根据checkTestVersion参数决定是否检查测试版本
	if checkTestVersion > 0 {
		return version, nil
	}
	// 如果不检查测试版本，需要过滤掉测试版本
	return c.BaseChecker.NormalizeVersion(version), nil
}
