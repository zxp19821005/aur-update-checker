package checkers

import (
	"regexp"
)

// VersionProcessor 版本处理器，合并了版本解析和比较功能
// 用于处理、解析和比较各种格式的版本号
type VersionProcessor struct {
	// 预编译的正则表达式缓存
	regexCache map[string]*regexp.Regexp
	// 预定义的正则表达式
	versionMainRegex *regexp.Regexp
	numberRegex     *regexp.Regexp
	semverRegex     *regexp.Regexp
	// 平台模式正则表达式
	platformRegexes []*regexp.Regexp
	// 测试版本模式正则表达式
	testRegexes []*regexp.Regexp
	// 前缀模式正则表达式
	prefixRegexes []*regexp.Regexp
}

// NewVersionProcessor 创建版本处理器
func NewVersionProcessor() *VersionProcessor {
	processor := &VersionProcessor{
		regexCache: make(map[string]*regexp.Regexp),
	}

	// 初始化预定义的正则表达式
	processor.initializeRegexes()

	return processor
}

// initializeRegexes 初始化所有正则表达式
func (p *VersionProcessor) initializeRegexes() {
	// 初始化基本正则表达式
	p.versionMainRegex = regexp.MustCompile(`^(\d+(?:\.\d+)*)`)
	p.numberRegex = regexp.MustCompile(`\d+`)
	// 语义化版本号正则表达式，匹配 v1.2.3, 1.2.3, 1.2.3-alpha, 1.2.3+build 等格式
	p.semverRegex = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

	// 初始化平台模式正则表达式
	p.initializePlatformRegexes()

	// 初始化测试版本模式正则表达式
	p.initializeTestRegexes()

	// 初始化前缀模式正则表达式
	p.initializePrefixRegexes()
}

// initializePlatformRegexes 初始化平台模式正则表达式
func (p *VersionProcessor) initializePlatformRegexes() {
	platformPatterns := []string{
		`[_\-\.](Linux|linux|Windows|Mac|windows|mac|osx|android|Android|iOS|ios|ubuntu|debian|fedora|centos|redhat|opensuse|arch|gentoo|mint)`,
		`[_\-\.](x86|x64|x86_64|x86-64|X86|X86_64|X86-64|amd64|arm|arm64|aarch64)`,
		`[_\-\.](32bit|64bit|32|_32|64|_64)`,
		`[_\-\.](bin|exe|dmg|pkg|deb|rpm|apk|AppImage)`,
		`[\.\-_](x86|x64|x86_64|amd64|arm|arm64|aarch64)`,
		`[\.\-_](signed|unsigned)`,
	}

	for _, pattern := range platformPatterns {
		p.platformRegexes = append(p.platformRegexes, regexp.MustCompile("(?i)"+pattern))
	}
}

// initializeTestRegexes 初始化测试版本模式正则表达式
func (p *VersionProcessor) initializeTestRegexes() {
	testPatterns := []string{
		`[_\-](stable|beta|alpha|rc|nightly|preview|pre|dev|test)`,
		`[_\-\.](snapshot|SNAPSHOT)`,
		`[_\-](milestone|m)`,
	}

	for _, pattern := range testPatterns {
		p.testRegexes = append(p.testRegexes, regexp.MustCompile("(?i)"+pattern))
	}
}

// initializePrefixRegexes 初始化前缀模式正则表达式
func (p *VersionProcessor) initializePrefixRegexes() {
	prefixPatterns := []string{
		`^release[_\-]`,
		`^version[_\-]`,
		`^ver[_\-]?`,
		`^v`,
		`^=V`,
		`^r`,
	}

	for _, pattern := range prefixPatterns {
		p.prefixRegexes = append(p.prefixRegexes, regexp.MustCompile("(?i)"+pattern))
	}
}
