package checkers

import (
	"aur-update-checker/internal/logger"
	"reflect"
)

// UpdatePackageWithParsedVersion 使用解析后的版本更新包信息
// 该方法接受一个包含 UpstreamVersion 和 UpstreamVersionRef 字段的结构体，并返回更新后的版本和是否应该更新的标志
func (p *VersionProcessor) UpdatePackageWithParsedVersion(pkgDetail interface{}, upstreamVersion string) (string, bool) {
	logger.GlobalLogger.Debugf("开始使用解析后的版本更新包信息，上游版本: %s", upstreamVersion)

	// 使用反射获取 UpstreamVersionRef 字段的值
	// 注意：这里简化了处理，实际应用中可能需要更健壮的类型检查
	pkgValue := reflect.ValueOf(pkgDetail)
	if pkgValue.Kind() == reflect.Ptr {
		pkgValue = pkgValue.Elem()
	}

	aurVersionRef := ""
	if upstreamVersionRefField := pkgValue.FieldByName("UpstreamVersionRef"); upstreamVersionRefField.IsValid() {
		aurVersionRef = upstreamVersionRefField.String()
	}
	logger.GlobalLogger.Debugf("从包详情中提取的AUR版本引用: %s", aurVersionRef)

	parsedVersion, shouldUpdate := p.ParseAndCompare(upstreamVersion, aurVersionRef)
	logger.GlobalLogger.Infof("包版本更新结果 - 解析版本: %s, 是否应该更新: %v", parsedVersion, shouldUpdate)

	return parsedVersion, shouldUpdate
}
