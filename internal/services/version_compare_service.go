package services

import (
	"aur-update-checker/internal/logger"
	"aur-update-checker/internal/utils"
)

// VersionService 版本服务
type VersionService struct {
	log *logger.Logger
}

// NewVersionService 创建版本服务实例
func NewVersionService(log *logger.Logger) *VersionService {
	return &VersionService{
		log: log,
	}
}

// CompareVersions 比较两个版本号
// 返回值: -1表示version1小于version2, 0表示相等, 1表示version1大于version2
func (s *VersionService) CompareVersions(version1, version2 string) int {
	// 使用工具函数比较版本
	return utils.CompareVersionStrings(version1, version2)
}

// NormalizeVersion 规范化版本号
func (s *VersionService) NormalizeVersion(version string) string {
	return utils.NormalizeVersionString(version)
}

// IsStableVersion 检查版本是否为稳定版本
func (s *VersionService) IsStableVersion(version string) bool {
	return utils.IsVersionStable(version)
}

// ExtractVersionComponents 从版本字符串中提取版本组件
func (s *VersionService) ExtractVersionComponents(version string) ([]int, []string) {
	return utils.ExtractVersionComponentsFromString(version)
}
