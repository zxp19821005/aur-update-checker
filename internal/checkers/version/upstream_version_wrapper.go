package checkers

import (
	"sync"
)

// 全局版本处理器实例
var (
	globalVersionProcessor *VersionProcessor
	versionProcessorOnce  sync.Once
)

// getGlobalVersionProcessor 获取全局版本处理器实例
func getGlobalVersionProcessor() *VersionProcessor {
	versionProcessorOnce.Do(func() {
		globalVersionProcessor = NewVersionProcessor()
	})
	return globalVersionProcessor
}

// UpstreamVersionParser 上游版本解析器，用于处理和比较上游版本与AUR版本引用
// 此结构体现在作为 VersionProcessor 的包装器，保持向后兼容性
type UpstreamVersionParser struct {
	processor *VersionProcessor
}

// NewUpstreamVersionParser 创建上游版本解析器
func NewUpstreamVersionParser() *UpstreamVersionParser {
	return &UpstreamVersionParser{
		processor: getGlobalVersionProcessor(),
	}
}

// ParseAndCompare 解析并比较上游版本与AUR版本引用
func (p *UpstreamVersionParser) ParseAndCompare(upstreamVersion, aurVersionRef string) (string, bool) {
	return p.processor.ParseAndCompare(upstreamVersion, aurVersionRef)
}

// cleanVersionWithOption 根据选项清理版本字符串，移除平台特定信息
// 注意：此方法目前未被使用，但保留以备将来可能需要使用
// nolint:unused
func (p *UpstreamVersionParser) cleanVersionWithOption(version string, checkTestVersion int) string {
	return p.processor.cleanVersionWithOption(version, checkTestVersion)
}

// UpdatePackageWithParsedVersion 使用解析后的版本更新包信息
func (p *UpstreamVersionParser) UpdatePackageWithParsedVersion(pkgDetail interface{}, upstreamVersion string) (string, bool) {
	return p.processor.UpdatePackageWithParsedVersion(pkgDetail, upstreamVersion)
}

// StandardizeVersion 标准化版本号，将不同格式的版本号转换为统一格式
func (p *UpstreamVersionParser) StandardizeVersion(version string) string {
	return p.processor.StandardizeVersion(version)
}

// CompareVersions 比较两个版本号，返回1表示v1大于v2，0表示相等，-1表示v1小于v2
func (p *UpstreamVersionParser) CompareVersions(v1, v2 string) int {
	return p.processor.CompareVersions(v1, v2)
}

// VersionComparator 版本比较器，用于比较各种格式的版本号
// 此结构体现在作为 VersionProcessor 的包装器，保持向后兼容性
type VersionComparator struct {
	processor *VersionProcessor
}

// NewVersionComparator 创建版本比较器
func NewVersionComparator() *VersionComparator {
	return &VersionComparator{
		processor: getGlobalVersionProcessor(),
	}
}

// CompareVersions 比较两个版本号，返回1表示v1大于v2，0表示相等，-1表示v1小于v2
func (c *VersionComparator) CompareVersions(v1, v2 string) int {
	return c.processor.CompareVersions(v1, v2)
}

// StandardizeVersion 标准化版本号，移除平台信息和测试版本标识符
func (c *VersionComparator) StandardizeVersion(version string) string {
	return c.processor.StandardizeVersion(version)
}



// IsStableVersion 检查版本是否为稳定版本
func (c *VersionComparator) IsStableVersion(version string) bool {
	return c.processor.IsStableVersion(version)
}

// ExtractVersionComponents 从版本字符串中提取版本组件
func (c *VersionComparator) ExtractVersionComponents(version string) ([]int, []string) {
	return c.processor.ExtractVersionComponents(version)
}

// GenerateVersionRef 生成上游版本提取参考值
func (c *VersionComparator) GenerateVersionRef(version string) string {
	return c.processor.GenerateVersionRef(version)
}

