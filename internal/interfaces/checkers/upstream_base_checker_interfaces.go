package checkers

import (
	"aur-update-checker/internal/logger"
	"fmt"
	"regexp"

	versionProcessor "aur-update-checker/internal/checkers/version"
)

// VersionExtractFunc 版本提取函数类型
type VersionExtractFunc func(content, key string) (string, error)

// BaseChecker 基础检查器，提供通用功能
type BaseChecker struct {
	name              string
	versionParser     interface{}
}

// NewBaseChecker 创建基础检查器
func NewBaseChecker(name string) *BaseChecker {
	return &BaseChecker{
		name:          name,
		versionParser: nil,
	}
}

// Name 实现接口方法
func (c *BaseChecker) Name() string {
	return c.name
}

// Supports 默认实现，所有URL都支持
// 具体检查器可以重写此方法以实现特定的URL支持逻辑
func (c *BaseChecker) Supports(url string) bool {
	return true
}

// Priority 默认实现，返回中等优先级
// 具体检查器可以重写此方法以设置特定的优先级
func (c *BaseChecker) Priority() int {
	return 50 // 0-100范围，50为中等优先级
}

// ExtractVersionFromContent 从内容中提取版本
func (c *BaseChecker) ExtractVersionFromContent(content, key string) (string, error) {
	// 如果没有提供key，尝试从内容中提取版本号
	if key == "" {
		// 尝试匹配常见的版本号格式
		versionPatterns := []string{
			`(\d+\.\d+\.\d+\.\d+)`,
			`(\d+\.\d+\.\d+)`,
			`(\d+\.\d+)`,
			`[a-zA-Z-]+-(\d+(?:\.\d+)+)`,
			`[a-zA-Z]+(\d+(?:\.\d+)+)`,
			`(\d+(?:\.\d+)+)`,
		}
		
		for _, pattern := range versionPatterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(content)
			if len(matches) > 1 {
				logger.GlobalLogger.Debugf("[%s] 从内容中提取到版本号: %s", c.name, matches[1])
				return matches[1], nil
			}
		}
		return "", fmt.Errorf("无法从内容中提取版本号")
	}
	
	// 使用提供的key作为正则表达式提取版本号
	re, err := regexp.Compile(key)
	if err != nil {
		return "", fmt.Errorf("编译正则表达式失败: %v", err)
	}
	
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		logger.GlobalLogger.Debugf("[%s] 使用正则表达式 %s 从内容中提取到版本号: %s", c.name, key, matches[1])
		return matches[1], nil
	}
	
	return "", fmt.Errorf("无法使用正则表达式 %s 从内容中提取版本号", key)
}

// NormalizeVersion 规范化版本号，移除平台特定信息
func (c *BaseChecker) NormalizeVersion(version string) string {
	return c.NormalizeVersionWithOption(version, 0) // 默认不检查测试版本
}

// NormalizeVersionWithOption 根据选项规范化版本号，移除平台特定信息
func (c *BaseChecker) NormalizeVersionWithOption(version string, checkTestVersion int) string {
	logger.GlobalLogger.Debugf("[%s] 原始版本: %s, 检查测试版本选项: %d", c.name, version, checkTestVersion)
	
	// 使用 UpstreamVersionParser 来清理版本号
	parser := versionProcessor.NewUpstreamVersionParser()
	normalized := parser.StandardizeVersion(version)
	
	logger.GlobalLogger.Debugf("[%s] 规范化后版本: %s", c.name, normalized)
	return normalized
}

// CleanVersionWithOption 清理版本号
func (c *BaseChecker) CleanVersionWithOption(version string, checkTestVersion int) string {
	// 简化实现，直接返回原始版本
	return version
}

// StandardizeVersion 标准化版本号
func (c *BaseChecker) StandardizeVersion(version string) string {
	// 尝试从文本中提取版本号
	versionPatterns := []string{
		// 匹配 x.y.z 格式，如 1.2.3
		`(\d+\.\d+\.\d+)`,
		// 匹配 x.y 格式，如 1.2
		`(\d+\.\d+)`,
		// 匹配 x.y.z.w 格式，如 1.2.3.4
		`(\d+\.\d+\.\d+\.\d+)`,
		// 匹配带前缀的版本，如 v1.2.3
		`[vV]?(\d+(?:\.\d+)+)`,
		// 匹配文件名中的版本，如 file-1.2.3.tar.gz
		`[\w-]+-(\d+(?:\.\d+)+)`,
		// 匹配任何数字和点组成的版本号
		`(\d+(?:\.\d+)+)`,
	}

	for _, pattern := range versionPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(version)
		if len(matches) > 1 {
			logger.GlobalLogger.Debugf("[%s] 从文本中提取到版本号: %s", c.name, matches[1])
			return matches[1]
		}
	}

	// 如果没有找到匹配的版本号，返回空字符串
	logger.GlobalLogger.Debugf("[%s] 无法从文本中提取版本号: %s", c.name, version)
	return ""
}
