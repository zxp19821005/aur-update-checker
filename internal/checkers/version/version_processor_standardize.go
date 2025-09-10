package checkers

import (
	"aur-update-checker/internal/logger"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
)

// StandardizeVersion 标准化版本号，将不同格式的版本号转换为统一格式
func (p *VersionProcessor) StandardizeVersion(version string) string {
	logger.GlobalLogger.Debugf("开始标准化版本号: %s", version)

	// 检查是否包含字母前缀的版本号，如 Alpha0.10.1, Beta1.2.3 等
	alphaPrefix := regexp.MustCompile(`^([A-Za-z]+)(\d+(?:\.\d+)*)$`)
	if alphaPrefix.MatchString(version) {
		matches := alphaPrefix.FindStringSubmatch(version)
		if len(matches) >= 3 {
			prefix := matches[1]
			versionPart := matches[2]
			logger.GlobalLogger.Debugf("检测到字母前缀版本号: %s, 前缀: %s, 版本部分: %s", version, prefix, versionPart)

			// 排除网络协议词汇，如 IPv4, IPv6 等
			if strings.HasPrefix(strings.ToLower(prefix), "ipv") {
				logger.GlobalLogger.Debugf("排除网络协议词汇: %s", version)
				return ""
			}

			// 尝试解析版本部分
			if _, err := semver.NewVersion(versionPart); err == nil {
				// 如果版本部分可以解析，返回原始版本号
				logger.GlobalLogger.Debugf("字母前缀版本号的版本部分解析成功，保留原始版本号: %s", version)

				// 对于单个字母前缀（如 V），移除前缀
				if len(prefix) == 1 {
					return versionPart
				}

				// 对于多字母前缀（如 Alpha, Beta），保留前缀
				return version
			}
		}
	}

	// 检查是否是特殊格式的版本号，如 5.8-5.3.14
	specialVersionFormat := regexp.MustCompile(`^\d+\.\d+-\d+\.\d+\.\d+$`)
	if specialVersionFormat.MatchString(version) {
		logger.GlobalLogger.Debugf("检测到特殊格式的版本号，直接返回原始版本号: %s", version)
		return version
	}

	// 检查是否是五部分版本号格式，如 1.10.12.394.001
	fivePartVersionFormat := regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+\.\d+$`)
	if fivePartVersionFormat.MatchString(version) {
		logger.GlobalLogger.Debugf("检测到五部分版本号格式，直接返回原始版本号: %s", version)
		return version
	}

	// 尝试使用 semver 库解析版本号
	if v, err := semver.NewVersion(version); err == nil {
		logger.GlobalLogger.Debugf("使用 semver 库成功解析版本号: %s", version)

		// 检查原始版本号是否为简化格式（如 9.4）
		simplifiedFormat := regexp.MustCompile(`^\d+\.\d+$`)
		if simplifiedFormat.MatchString(version) {
			// 如果是简化格式，直接使用原始版本号，不进行补全
			logger.GlobalLogger.Debugf("检测到简化版本号格式，保留原始格式: %s", version)
			cleaned := p.cleanVersionWithOption(version, 0)
			if cleaned != "" {
				logger.GlobalLogger.Debugf("清理后的版本号: %s -> %s", version, cleaned)
				return cleaned
			}
			return version
		}

		// 清理版本号，移除平台特定信息
		cleaned := p.cleanVersionWithOption(v.String(), 0)
		if cleaned != "" {
			logger.GlobalLogger.Debugf("清理后的版本号: %s -> %s", v.String(), cleaned)
			return cleaned
		}
		// 如果清理失败，返回原始版本号
		return v.String()
	}

	// 如果 semver 解析失败，尝试清理版本号后再次解析
	cleaned := p.cleanVersionWithOption(version, 0)
	if v, err := semver.NewVersion(cleaned); err == nil {
		logger.GlobalLogger.Debugf("清理版本号后使用 semver 库成功解析: %s -> %s", version, cleaned)
		return v.String()
	}

	// 如果仍然失败，尝试添加 v 前缀
	if !strings.HasPrefix(version, "v") {
		if v, err := semver.NewVersion("v" + cleaned); err == nil {
			logger.GlobalLogger.Debugf("添加 v 前缀后使用 semver 库成功解析: %s -> v%s", version, cleaned)
			return v.String()
		}
	}

	// 如果 semver 库无法解析，回退到原有逻辑
	logger.GlobalLogger.Debugf("semver 库无法解析版本号，回退到原有逻辑: %s", version)
	return p.fallbackStandardizeVersion(version)
}

// fallbackStandardizeVersion 当 semver 库无法解析时的回退标准化方法
func (p *VersionProcessor) fallbackStandardizeVersion(version string) string {
	// 移除前缀
	result := version
	for _, re := range p.prefixRegexes {
		result = re.ReplaceAllString(result, "")
	}

	// 检查是否为语义化版本号
	if p.semverRegex.MatchString(result) {
		logger.GlobalLogger.Debugf("版本号 %s 符合语义化版本号格式", result)
		// 已经是标准格式，直接返回
		return result
	}

	// 首先尝试匹配最具体的版本号格式
	// 定义版本模式，按照优先级排序
	versionPatterns := []string{
		`(\d+\.\d+\.\d+\.\d+[A-Z]+(?:\.[A-Z0-9]+)?)`,  // 如 9.0.3988.101ZH.S1 或 9.0.3988.101ZH
		`(\d+\.\d+\.\d+\.\d+\.\d+)`,                   // 如 1.10.12.394.001
		`(\d+\.\d+\.\d+\.\d+-\d+)`,                     // 如 1.0.45966.7-1
		`(\d+\.\d+\.\d+\.\d+)`,                        // 如 9.0.3988.101
		`(\d+\.\d+\.\d+-\d+)`,                         // 如 1.2.3-1
		`(\d+\.\d+\.\d+)`,                             // 如 9.0.3988
		`(\d+\.\d+-\d+)`,                              // 如 1.2-1
		`(\d+\.\d+)`,                                  // 如 9.0
		`[a-zA-Z-]+-(\d+(?:\.\d+)+)`,                  // 从文件名中提取版本号，如 helio-3.16
		`[a-zA-Z]+(\d+(?:\.\d+)+)`,                    // 从文件名中提取版本号，如 file3.16
		`(\d+(?:\.\d+)+)`,                             // 任何数字加点序列
	}

	// 尝试匹配版本模式
	for _, pattern := range versionPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(result)
		if len(matches) > 1 {
			logger.GlobalLogger.Debugf("从 %s 中使用模式 %s 匹配到版本号: %s", result, pattern, matches[1])
			// 如果匹配到最完整的版本号格式（包含字母和后缀），需要进一步验证
			if strings.Contains(matches[1], "ZH") || strings.Contains(matches[1], "S1") {
				// 检查是否是有效的版本号格式，如 9.0.5004.101ZH.S1
				validVersionFormat := regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+ZH\.S1$`)
				if validVersionFormat.MatchString(matches[1]) {
					cleaned := p.cleanVersion(matches[1])
					logger.GlobalLogger.Debugf("版本号标准化结果: %s -> %s", version, cleaned)
					return cleaned
				}
				// 如果不是有效的版本号格式，继续处理
			}
			result = matches[1]
			break
		}
	}

	// 如果没有匹配到版本模式，尝试提取主版本号
	if p.versionMainRegex.MatchString(result) {
		mainVersion := p.versionMainRegex.FindString(result)
		logger.GlobalLogger.Debugf("从 %s 中提取主版本号: %s", result, mainVersion)

		// 验证主版本号是否有效
		// 如果主版本号只是一个单一的数字，且没有点号分隔，那么它可能不是一个有效的版本号
		if !strings.Contains(mainVersion, ".") {
			logger.GlobalLogger.Debugf("主版本号 %s 不包含点号分隔符，可能不是有效的版本号", mainVersion)
			return ""
		}

		// 检查原始字符串中是否包含连字符和额外的版本信息
		// 如果原始字符串中包含连字符，并且连字符后面有数字，则保留这部分
		if strings.Contains(result, "-") {
			// 使用正则表达式匹配连字符和后面的数字
			hyphenPattern := regexp.MustCompile(`^(\d+(?:\.\d+)*)(-\d+)`)
			hyphenMatches := hyphenPattern.FindStringSubmatch(result)
			if len(hyphenMatches) > 2 {
				// 组合主版本号和连字符部分
				result = mainVersion + hyphenMatches[2]
				logger.GlobalLogger.Debugf("保留连字符部分，最终版本号: %s", result)
				cleaned := p.cleanVersion(result)
				logger.GlobalLogger.Debugf("版本号标准化结果: %s -> %s", version, cleaned)
				return cleaned
			}
		}

		result = mainVersion
	} else {
		// 如果无法提取主版本号，我们需要更严格的验证
		// 首先检查是否是URL，如果是URL，则不应该从中提取版本号
		if strings.Contains(result, "://") {
			logger.GlobalLogger.Debugf("%s 是一个URL，不从中提取版本号", result)
			return ""
		}

		// 尝试提取第一个数字
		firstNum := p.numberRegex.FindString(result)
		if firstNum != "" {
			logger.GlobalLogger.Debugf("从 %s 中提取第一个数字: %s", result, firstNum)

			// 验证提取的数字是否可能是有效的版本号
			// 需要满足以下条件：
			// 1. 原始字符串中必须包含点号
			// 2. 提取的数字附近必须有点号或其他版本号特征
			if !strings.Contains(result, ".") {
				logger.GlobalLogger.Debugf("提取的数字 %s 不包含点号分隔符，可能不是有效的版本号", firstNum)
				return ""
			}

			// 检查提取的数字附近是否有点号或其他版本号特征
			// 查找提取的数字在原始字符串中的位置
			numIndex := strings.Index(result, firstNum)
			if numIndex == -1 {
				logger.GlobalLogger.Debugf("无法在原始字符串中找到提取的数字 %s", firstNum)
				return ""
			}

			// 检查数字前后是否有版本号特征
			// 获取数字前后的上下文
			contextStart := numIndex - 5
			if contextStart < 0 {
				contextStart = 0
			}
			contextEnd := numIndex + len(firstNum) + 5
			if contextEnd > len(result) {
				contextEnd = len(result)
			}
			context := result[contextStart:contextEnd]

			// 检查上下文中是否包含版本号特征
			versionFeatures := []string{".", "-", "_", "v", "V", "version", "Version"}
			hasVersionFeature := false
			for _, feature := range versionFeatures {
				if strings.Contains(context, feature) {
					hasVersionFeature = true
					break
				}
			}

			if !hasVersionFeature {
				logger.GlobalLogger.Debugf("提取的数字 %s 附近没有版本号特征，可能不是有效的版本号", firstNum)
				return ""
			}

			result = firstNum
		}
	}

	// 清理版本号，移除平台信息和测试版本标识符
	cleaned := p.cleanVersion(result)
	logger.GlobalLogger.Debugf("版本号标准化结果: %s -> %s", version, cleaned)
	return cleaned
}
