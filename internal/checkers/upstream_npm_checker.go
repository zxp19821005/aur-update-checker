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

// NpmPackage NPM包信息
type NpmPackage struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Version     string            `json:"version"`
	DistTags    map[string]string `json:"dist-tags"`
	Versions    map[string]interface{} `json:"versions"`
}

// NpmChecker NPM检查器
type NpmChecker struct {
	*checkerInterfaces.BaseChecker
	client *http.Client
}

// NewNpmChecker 创建NPM检查器
func NewNpmChecker() *NpmChecker {
	return &NpmChecker{
		BaseChecker: checkerInterfaces.NewBaseChecker("npm"),
		client:      &http.Client{},
	}
}

// Check 实现检查器接口，从NPM获取包版本
func (c *NpmChecker) Check(ctx context.Context, url, versionExtractKey string) (string, error) {
	// 默认不检查测试版本
	return c.CheckWithOption(ctx, url, versionExtractKey, 0)
}

// CheckWithOption 实现检查器接口，根据选项从NPM获取包版本
func (c *NpmChecker) CheckWithOption(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error) {
	// 从URL或versionExtractKey中提取包名
	packageName, err := c.extractPackageName(url, versionExtractKey)
	if err != nil {
		errMsg := fmt.Errorf("提取NPM包名失败: %v", err)
		logger.GlobalLogger.Errorf("[npm] %v", errMsg)
		return "", errMsg
	}

	// 获取NPM包信息
	packageInfo, err := c.fetchPackageInfo(ctx, packageName)
	if err != nil {
		logger.GlobalLogger.Errorf("[npm] 获取NPM包信息失败: %v", err)
		return "", fmt.Errorf("获取NPM包信息失败: %v", err)
	}

	// 提取版本
	version, err := c.extractVersionWithOption(packageInfo, versionExtractKey, checkTestVersion)
	if err != nil {
		logger.GlobalLogger.Errorf("[npm] 提取版本失败: %v", err)
		return "", fmt.Errorf("提取版本失败: %v", err)
	}

	// 规范化版本号，移除平台特定信息
	normalizedVersion := c.BaseChecker.NormalizeVersionWithOption(version, checkTestVersion)
	return normalizedVersion, nil
}

// extractPackageName 从URL或versionExtractKey中提取包名
func (c *NpmChecker) extractPackageName(url, versionExtractKey string) (string, error) {
	// 优先尝试从versionExtractKey中提取包名
	if versionExtractKey != "" {
		// 如果versionExtractKey看起来像包名，直接使用
		if c.isValidPackageName(versionExtractKey) {
			return versionExtractKey, nil
		}
	}

	// 尝试从URL中提取包名
	// 匹配npmjs.com/package/<package-name>格式
	re := regexp.MustCompile(`npmjs\.com/package/([^/\s]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		return matches[1], nil
	}

	// 匹配npmjs.com/<package-name>格式
	re = regexp.MustCompile(`npmjs\.com/([^/\s]+)`)
	matches = re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		// 确保不是特殊路径，如 ~, /search 等
		pkgName := matches[1]
		if c.isValidPackageName(pkgName) {
			return pkgName, nil
		}
	}

	logger.GlobalLogger.Errorf("[npm] 无法从URL或versionExtractKey中提取有效的NPM包名")
	return "", fmt.Errorf("无法从URL或versionExtractKey中提取有效的NPM包名")
}

// isValidPackageName 检查是否是有效的NPM包名
func (c *NpmChecker) isValidPackageName(name string) bool {
	// 简单检查，实际NPM包名规则更复杂
	if name == "" || name == "package" || name == "search" || name == "~" {
		return false
	}
	return true
}

// fetchPackageInfo 获取NPM包信息
func (c *NpmChecker) fetchPackageInfo(ctx context.Context, packageName string) (*NpmPackage, error) {
	// 使用国内NPM镜像源解决访问问题
	apiURL := fmt.Sprintf("https://registry.npmmirror.com/%s", packageName)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		logger.GlobalLogger.Errorf("[npm] 创建请求失败: %v", err)
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
		logger.GlobalLogger.Errorf("[npm] 读取响应体失败: %v", err)
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	var packageInfo NpmPackage
	if err := json.Unmarshal(body, &packageInfo); err != nil {
		logger.GlobalLogger.Errorf("[npm] 解析响应失败: %v", err)
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &packageInfo, nil
}

// extractVersionWithOption 根据选项从包信息中提取版本
func (c *NpmChecker) extractVersionWithOption(packageInfo *NpmPackage, versionExtractKey string, checkTestVersion int) (string, error) {
	// 如果versionExtractKey为空，使用latest标签
	if versionExtractKey == "" {
		if latest, ok := packageInfo.DistTags["latest"]; ok {
			// 根据checkTestVersion参数决定是否检查测试版本
			if checkTestVersion > 0 {
				return latest, nil
			}
			// 如果不检查测试版本，需要过滤掉测试版本
			return c.BaseChecker.NormalizeVersion(latest), nil
		}
		// 根据checkTestVersion参数决定是否检查测试版本
		if checkTestVersion > 0 {
			return packageInfo.Version, nil
		}
		// 如果不检查测试版本，需要过滤掉测试版本
		return c.BaseChecker.NormalizeVersion(packageInfo.Version), nil
	}

	// 检查versionExtractKey是否是dist-tags中的标签
	if tag, ok := packageInfo.DistTags[versionExtractKey]; ok {
		// 根据checkTestVersion参数决定是否检查测试版本
		if checkTestVersion > 0 {
			return tag, nil
		}
		// 如果不检查测试版本，需要过滤掉测试版本
		return c.BaseChecker.NormalizeVersion(tag), nil
	}

	// 检查versionExtractKey是否是版本号
	if _, ok := packageInfo.Versions[versionExtractKey]; ok {
		// 根据checkTestVersion参数决定是否检查测试版本
		if checkTestVersion > 0 {
			return versionExtractKey, nil
		}
		// 如果不检查测试版本，需要过滤掉测试版本
		return c.BaseChecker.NormalizeVersion(versionExtractKey), nil
	}

	// 如果都不是，尝试使用versionExtractKey作为提取规则
	version, err := c.BaseChecker.ExtractVersionFromContent(packageInfo.Version, versionExtractKey)
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

// CheckWithVersionRef 带选项和版本引用地检查上游版本
func (c *NpmChecker) CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	// 如果versionRef为空，使用默认的CheckWithOption方法
	if versionRef == "" {
		return c.CheckWithOption(ctx, url, versionExtractKey, checkTestVersion)
	}

	// 从URL或versionExtractKey中提取包名
	packageName, err := c.extractPackageName(url, versionExtractKey)
	if err != nil {
		errMsg := fmt.Errorf("提取NPM包名失败: %v", err)
		logger.GlobalLogger.Errorf("[npm] %v", errMsg)
		return "", errMsg
	}

	// 获取NPM包信息
	packageInfo, err := c.fetchPackageInfo(ctx, packageName)
	if err != nil {
		logger.GlobalLogger.Errorf("[npm] 获取NPM包信息失败: %v", err)
		return "", fmt.Errorf("获取NPM包信息失败: %v", err)
	}

	// 检查versionRef是否是有效的版本号
	if _, ok := packageInfo.Versions[versionRef]; ok {
		// 如果versionRef是有效的版本号，使用它
		normalized := c.BaseChecker.NormalizeVersionWithOption(versionRef, checkTestVersion)
		// 标准化版本号，移除前缀如 'v'
		standardized := c.BaseChecker.StandardizeVersion(normalized)
		return standardized, nil
	}

	// 如果versionRef不是有效的版本号，使用默认的CheckWithOption方法
	logger.GlobalLogger.Warnf("[npm] versionRef '%s' 不是有效的版本号，将使用默认方法获取版本", versionRef)
	return c.CheckWithOption(ctx, url, versionExtractKey, checkTestVersion)
}
