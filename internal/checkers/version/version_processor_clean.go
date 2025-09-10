package checkers

import (
	"aur-update-checker/internal/logger"
	"fmt"
	"regexp"
	"strings"
)

// cleanVersionWithOption 根据选项清理版本字符串，移除平台特定信息
func (p *VersionProcessor) cleanVersionWithOption(version string, checkTestVersion int) string {
	logger.GlobalLogger.Debugf("开始清理版本字符串: %s, 检查测试版本选项: %d", version, checkTestVersion)

	result := version

	// 移除版本号前的V前缀（如果有）
	result = strings.TrimPrefix(result, "V")

	// 检查是否包含字母前缀的版本号，如 Alpha0.10.1, Beta1.2.3 等
	alphaPrefix := regexp.MustCompile(`^([A-Za-z]+)(\d+(?:\.\d+)*)$`)
	if alphaPrefix.MatchString(result) {
		matches := alphaPrefix.FindStringSubmatch(result)
		if len(matches) >= 3 {
			prefix := matches[1]
			versionPart := matches[2]
			logger.GlobalLogger.Debugf("检测到字母前缀版本号: %s, 前缀: %s, 版本部分: %s", result, prefix, versionPart)

			// 检查版本部分是否是语义化版本号格式
			if p.semverRegex.MatchString(versionPart) {
				logger.GlobalLogger.Debugf("版本部分 %s 符合语义化版本号格式，保留原始版本号", versionPart)
				// 对于包含字母前缀的版本号，直接返回原始版本号，不进行任何处理
				return result
			}
		}
	}

	// 检查是否是语义化版本号格式
	if p.semverRegex.MatchString(result) {
		logger.GlobalLogger.Debugf("版本号 %s 符合语义化版本号格式，但仍需检查平台信息", result)
		// 即使是语义化版本号，也需要检查并移除平台特定信息
		// 特殊处理常见的平台后缀，如 -ubuntu-amd64, -windows, -macos 等
		platformSuffixes := []string{
			`-ubuntu-\w+`, `-ubuntu`, `-linux-\w+`, `-linux`,
			`-windows-\w+`, `-windows`, `-win\d{2}`, `-win`,
			`-macos-\w+`, `-macos`, `-osx-\w+`, `-osx`, `-darwin-\w+`, `-darwin`,
			`-android-\w+`, `-android`,
			`-x86_64`, `-x64`, `-amd64`, `-i386`, `-i686`, `-x86`, `-arm\d*`, `-aarch\d*`,
			`-32bit`, `-64bit`, `-bit`,
			`-deb`, `-rpm`, `-dmg`, `-exe`, `-msi`, `-tar\.gz`, `-zip`,
		}

		for _, suffix := range platformSuffixes {
			re := regexp.MustCompile(fmt.Sprintf("%s$", suffix))
			if re.MatchString(result) {
				result = re.ReplaceAllString(result, "")
				logger.GlobalLogger.Debugf("移除平台后缀 %s 后的版本号: %s", suffix, result)
				break
			}
		}
	} else {
		// 检查是否是五部分版本号格式，如 1.10.12.394.001
		fivePartVersion := regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+\.\d+$`)
		if fivePartVersion.MatchString(result) {
			logger.GlobalLogger.Debugf("版本号 %s 符合五部分版本号格式，保留完整版本号", result)
			// 直接返回，不做进一步处理
			return result
		}
	}

	// 使用预编译的平台正则表达式，但要确保不会破坏版本号结构
	for _, re := range p.platformRegexes {
		// 只替换版本号末尾的平台信息，保留中间部分
		result = regexp.MustCompile(fmt.Sprintf("%s$", re.String())).ReplaceAllString(result, "")
	}

	// 使用预编译的平台正则表达式，检查并移除版本号末尾的平台信息
	for _, re := range p.platformRegexes {
		// 只替换版本号末尾的平台信息
		result = re.ReplaceAllString(result, "")
	}

	// 特殊处理常见的平台后缀，如 -ubuntu-amd64, -windows, -macos 等
	platformSuffixes := []string{
		`-ubuntu-\w+`, `-ubuntu`, `-linux-\w+`, `-linux`,
		`-windows-\w+`, `-windows`, `-win\d{2}`, `-win`,
		`-macos-\w+`, `-macos`, `-osx-\w+`, `-osx`, `-darwin-\w+`, `-darwin`,
		`-android-\w+`, `-android`,
		`-x86_64`, `-x64`, `-amd64`, `-i386`, `-i686`, `-x86`, `-arm\d*`, `-aarch\d*`,
		`-32bit`, `-64bit`, `-bit`,
		`-deb`, `-rpm`, `-dmg`, `-exe`, `-msi`, `-tar\.gz`, `-zip`,
	}

	for _, suffix := range platformSuffixes {
		re := regexp.MustCompile(fmt.Sprintf("%s$", suffix))
		result = re.ReplaceAllString(result, "")
	}

	// 检查是否包含测试版本标识符
	hasTestVersion := false
	for _, re := range p.testRegexes {
		if re.MatchString(result) {
			hasTestVersion = true
			break
		}
	}

	// 根据检查测试版本选项处理
	if checkTestVersion == 0 {
		// 不检查测试版本，排除包含测试版本标识符的版本
		if hasTestVersion {
			logger.GlobalLogger.Debugf("不检查测试版本，版本 %s 包含测试版本标识符，将被忽略", result)
			return ""
		}
		logger.GlobalLogger.Debugf("不检查测试版本，版本 %s 不包含测试版本标识符，保留", result)
	} else {
		// 检查测试版本，移除测试版本标识符
		// 使用预编译的测试版本正则表达式
		for _, re := range p.testRegexes {
			// 只替换版本号末尾的测试版本标识符，保留中间部分
			result = regexp.MustCompile(fmt.Sprintf("%s$", re.String())).ReplaceAllString(result, "")
		}
		logger.GlobalLogger.Debugf("检查测试版本，已移除测试版本标识符: %s", result)
	}

	// 移除多余的下划线，但保留连字符（可能包含版本号后缀如-1）
	result = strings.TrimRight(result, "_")

	// 确保版本号格式正确，至少包含主版本号和次版本号
	// 首先检查版本号是否包含连字符后缀（如-1）
	hasHyphenSuffix := strings.Contains(result, "-")
	var hyphenPart string
	if hasHyphenSuffix {
		// 提取连字符部分
		parts := strings.SplitN(result, "-", 2)
		result = parts[0]
		if len(parts) > 1 {
			hyphenPart = "-" + parts[1]
		}
	}

	// 处理点号分隔的版本号部分
	versionParts := strings.Split(result, ".")
	if len(versionParts) >= 2 {
		// 确保每个部分都是有效的，但保留字母部分（如ZH）
		for i, part := range versionParts {
			// 对于包含字母的部分（如101ZH），保留整个部分
			if regexp.MustCompile(`^\d+[A-Za-z]+`).MatchString(part) {
				continue // 保留原样
			}
			// 对于纯数字部分，确保是有效的数字
			re := regexp.MustCompile(`^\d+`)
			if match := re.FindString(part); match != "" {
				versionParts[i] = match
			}
		}
		result = strings.Join(versionParts, ".")
	}

	// 如果有连字符后缀，重新添加
	if hasHyphenSuffix && hyphenPart != "" {
		result += hyphenPart
	}

	cleaned := strings.TrimSpace(result)
	// 去掉末尾的空连字符
	cleaned = strings.TrimSuffix(cleaned, "-")
	logger.GlobalLogger.Debugf("版本字符串清理结果: %s -> %s", version, cleaned)
	return cleaned
}

// cleanVersion 清理版本字符串，移除平台特定信息（保留测试版本标识符）
func (p *VersionProcessor) cleanVersion(version string) string {
	return p.cleanVersionWithOption(version, 0) // 默认保留测试版本标识符
}


