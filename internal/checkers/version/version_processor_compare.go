package checkers

import (
	"aur-update-checker/internal/logger"
	"aur-update-checker/internal/utils"
	"fmt"

	"github.com/Masterminds/semver"
)

// CompareVersions 比较两个版本号，返回1表示v1大于v2，0表示相等，-1表示v1小于v2
func (p *VersionProcessor) CompareVersions(v1, v2 string) int {
	logger.GlobalLogger.Debugf("比较版本号: %s 和 %s", v1, v2)

	// 标准化版本号
	stdV1 := p.StandardizeVersion(v1)
	stdV2 := p.StandardizeVersion(v2)
	logger.GlobalLogger.Debugf("标准化后的版本号: %s 和 %s", stdV1, stdV2)

	// 如果标准化后的版本号相同，则认为版本相等
	if stdV1 == stdV2 {
		logger.GlobalLogger.Debugf("版本号 %s 和 %s 相等", v1, v2)
		return 0
	}

	// 尝试使用 semver 库解析版本号
	v1Semver, err1 := p.parseToSemver(stdV1)
	v2Semver, err2 := p.parseToSemver(stdV2)

	if err1 == nil && err2 == nil {
		// 两个版本都是有效的语义化版本，直接比较
		logger.GlobalLogger.Debugf("使用 semver 库比较版本号: %s 和 %s", stdV1, stdV2)
		if v1Semver.GreaterThan(v2Semver) {
			return 1
		} else if v1Semver.LessThan(v2Semver) {
			return -1
		}
		return 0
	}

	// 如果不是有效的语义化版本，使用 utils 包中的版本比较函数
	logger.GlobalLogger.Debugf("版本号不是有效的语义化版本，使用 utils 包中的版本比较函数: %s 和 %s", stdV1, stdV2)
	return utils.CompareVersionStrings(stdV1, stdV2)
}

// ParseAndCompare 解析并比较上游版本与AUR版本引用
func (p *VersionProcessor) ParseAndCompare(upstreamVersion, aurVersionRef string) (string, bool) {
	logger.GlobalLogger.Debugf("开始解析和比较版本 - 上游版本: %s, AUR版本引用: %s", upstreamVersion, aurVersionRef)

	// 如果AUR版本引用为空，直接返回上游版本
	if aurVersionRef == "" {
		stdVersion := p.StandardizeVersion(upstreamVersion)
		logger.GlobalLogger.Debugf("AUR版本引用为空，直接返回标准化后的上游版本: %s", stdVersion)
		return stdVersion, true
	}

	// 标准化两个版本字符串
	stdUpstream := p.StandardizeVersion(upstreamVersion)
	stdAurRef := p.StandardizeVersion(aurVersionRef)
	logger.GlobalLogger.Debugf("标准化后的版本 - 上游: %s, AUR引用: %s", stdUpstream, stdAurRef)

	// 比较版本号
	comparison := p.CompareVersions(stdUpstream, stdAurRef)
	logger.GlobalLogger.Debugf("版本号比较结果: %d", comparison)

	// 如果上游版本大于或等于AUR版本引用，建议更新
	if comparison >= 0 {
		logger.GlobalLogger.Infof("上游版本 %s 大于或等于AUR版本引用 %s，建议更新", stdUpstream, stdAurRef)
		return stdUpstream, true
	}

	// 如果上游版本小于AUR版本引用，不建议更新
	logger.GlobalLogger.Infof("上游版本 %s 小于AUR版本引用 %s，不建议更新", stdUpstream, stdAurRef)
	return stdUpstream, false
}

// parseToSemver 尝试将版本字符串解析为语义化版本
func (p *VersionProcessor) parseToSemver(version string) (*semver.Version, error) {
	// 检查版本字符串是否符合语义化版本格式
	if !p.semverRegex.MatchString(version) {
		return nil, fmt.Errorf("版本 '%s' 不符合语义化版本格式", version)
	}

	// 使用 semver 库解析版本
	return semver.NewVersion(version)
}

// IsStableVersion 检查版本是否为稳定版本
func (p *VersionProcessor) IsStableVersion(version string) bool {
	return utils.IsVersionStable(version)
}

// ExtractVersionComponents 从版本字符串中提取版本组件
func (p *VersionProcessor) ExtractVersionComponents(version string) ([]int, []string) {
	return utils.ExtractVersionComponentsFromString(version)
}

// GenerateVersionRef 生成上游版本提取参考值
func (p *VersionProcessor) GenerateVersionRef(version string) string {
	return utils.GenerateVersionRef(version)
}
