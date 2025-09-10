package services

import (
	"strings"
)

// AurVersionParser AUR版本解析器
type AurVersionParser struct{}

// NewAurVersionParser 创建AUR版本解析器实例
func NewAurVersionParser() *AurVersionParser {
	return &AurVersionParser{}
}

// ExtractPkgver 从完整的版本字符串中提取 pkgver 部分
// AUR 版本格式通常是 epoch:pkgver-pkgrel，我们需要提取其中的 pkgver 部分
func (p *AurVersionParser) ExtractPkgver(fullVersion string) string {
	// 首先处理 epoch 部分（如果存在）
	pkgver := fullVersion
	if colonIndex := strings.Index(fullVersion, ":"); colonIndex != -1 {
		pkgver = fullVersion[colonIndex+1:]
	}

	// 然后处理 pkgrel 部分（如果存在）
	if dashIndex := strings.LastIndex(pkgver, "-"); dashIndex != -1 {
		pkgver = pkgver[:dashIndex]
	}

	// AUR规定，软件版本中不允许包含-，所以有些软件包使用_代替-
	// 这里我们将_替换为-，以便正确比较版本
	pkgver = strings.ReplaceAll(pkgver, "_", "-")

	return pkgver
}

// ParseAndSaveVersion 解析完整版本并返回应该保存的版本字符串
// 这里我们将提取的 pkgver 作为应该保存的版本
func (p *AurVersionParser) ParseAndSaveVersion(fullVersion string) string {
	return p.ExtractPkgver(fullVersion)
}
