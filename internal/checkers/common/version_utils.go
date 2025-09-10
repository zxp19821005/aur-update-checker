package common

import (
	"aur-update-checker/internal/logger"
	"regexp"
	"strings"
)

// ExtractVersionFromString 从字符串中提取版本号
// 这是一个通用的版本号提取函数，可以被多个检查器共享使用
func ExtractVersionFromString(s string) string {
	logger.GlobalLogger.Debugf("[version_utils] 尝试从字符串中提取版本号: %s", s)

	// 检查是否包含测试版本标识符
	testVersionPatterns := []string{"-alpha", "-beta", "-dev", "-rc", "-test", "-preview", "-pre"}
	for _, pattern := range testVersionPatterns {
		if strings.Contains(strings.ToLower(s), pattern) {
			logger.GlobalLogger.Debugf("[version_utils] 字符串包含测试版本标识符 %s，跳过提取", pattern)
			return ""
		}
	}

	// 1. 首先尝试匹配最复杂的版本号格式，如 9.0.3988.101ZH.S1
	re := regexp.MustCompile(`v?(\d+\.\d+\.\d+\.\d+[A-Z]+(?:\.[A-Z0-9]+)?)`)
	matches := re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到复杂版本号格式: %s", matches[1])
		return matches[1]
	}

	// 2. 匹配四部分版本号格式，如 1.2.3.4 或 1.2.3.4-1
	re = regexp.MustCompile(`v?(\d+\.\d+\.\d+\.\d+(?:-\d+)?)`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到四部分版本号格式: %s", matches[1])
		return matches[1]
	}

	// 3. 匹配文件名中的特殊格式版本号，如 spark-dwine-helper_5.8-5.3.14_all.deb 中的 5.8-5.3.14
	re = regexp.MustCompile(`spark-dwine-helper_(\d+\.\d+-\d+\.\d+\.\d+)`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到文件名中的特殊格式版本号: %s", matches[1])
		return matches[1]
	}

	// 4. 匹配文件名中的通用特殊格式版本号，如 xxx_5.8-5.3.14_all.deb 中的 5.8-5.3.14
	re = regexp.MustCompile(`_(\d+\.\d+-\d+\.\d+\.\d+)_`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到文件名中的通用特殊格式版本号: %s", matches[1])
		return matches[1]
	}

	// 4. 匹配URL路径中的版本号，如 /vikunja/0.24.6 中的 0.24.6
	re = regexp.MustCompile(`/(\d+\.\d+\.\d+)`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到URL路径中的版本号格式: %s", matches[1])
		return matches[1]
	}

	// 5. 匹配URL中的文件名版本号，如 youdao-dict_6.0.0-ubuntu-amd64.deb 中的 6.0.0
	re = regexp.MustCompile(`[a-zA-Z-]+_(\d+\.\d+\.\d+)-`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到URL中的文件名版本号格式: %s", matches[1])
		return matches[1]
	}

	// 7. 匹配HTML中的标题标签中的版本号，如 <h2 id="2647" class="markdown-doc-viewer-heading">2.6.47
	re = regexp.MustCompile(`<h\d+[^>]*>(\d+\.\d+(?:\.\d+)?)`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到HTML标题标签中的版本号格式: %s", matches[1])
		return matches[1]
	}

	// 8. 匹配HTML内容中的版本号，如 <p>V1.4.2<br> 中的 V1.4.2
	re = regexp.MustCompile(`<p>([Vv]?\d+\.\d+(?:\.\d+)?)`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到HTML内容中的版本号格式: %s", matches[1])
		return matches[1]
	}

	// 9. 匹配HTML注释中的版本号，如 v<!-- -->7.2.1<!-- --> 中的 7.2.1
	re = regexp.MustCompile(`v<!-- -->(\d+\.\d+(?:\.\d+)?)`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到HTML注释中的版本号格式: %s", matches[1])
		return matches[1]
	}

	// 10. 匹配更新日志中的版本号，如更新日志</h3><div data-v-086de9f7="" class="klBox updateLog">4.0.2
	re = regexp.MustCompile(`更新日志[^>]*>([^<]*\d+\.\d+(?:\.\d+)?)`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到更新日志中的版本号格式: %s", matches[1])
		return matches[1]
	}

	// 11. 匹配文件名中的版本号，如 spark-dwine-helper_5.8-5.3.14_all.deb 中的 5.8-5.3.14
	re = regexp.MustCompile(`_(\d+\.\d+-\d+\.\d+\.\d+)_`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到文件名中的版本号格式: %s", matches[1])
		return matches[1]
	}

	// 12. 匹配连字符后的版本号，如 flomo卡笔记-5.25.91-最新版本.exe 中的 5.25.91
	re = regexp.MustCompile(`-(\d+\.\d+\.\d+)`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到连字符后的版本号格式: %s", matches[1])
		return matches[1]
	}

	// 13. 匹配包含字母前缀的版本号，如 Alpha0.10.1, Beta1.2.3 等，但要排除HTML标签和网络协议词汇
	re = regexp.MustCompile(`([A-Za-z]{3,}\d+(?:\.\d+)*)`)  // 至少3个字母，避免匹配h1, h2等HTML标签
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		version := matches[1]
		// 排除网络协议词汇，如 IPv4, IPv6 等
		if strings.HasPrefix(strings.ToLower(version), "ipv") {
			logger.GlobalLogger.Debugf("[version_utils] 排除网络协议词汇: %s", version)
		} else {
			logger.GlobalLogger.Debugf("[version_utils] 找到包含字母前缀的版本号格式: %s", version)
			return version
		}
	}

	// 14. 匹配引号内的版本号，如 "9.4", "1.2.3" 等
	re = regexp.MustCompile(`"(\d+(?:\.\d+)*)"`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到引号内的版本号格式: %s", matches[1])
		return matches[1]
	}

	// 15. 匹配标准版本号格式，如 v1.0.0, 1.2.3, 2.1.0-beta, 1.2.3-1等
	re = regexp.MustCompile(`v?(\d+\.\d+\.\d+(?:[-\.]\d+)*(?:[-\w]*)?)`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到标准版本号格式: %s", matches[1])
		return matches[1]
	}

	// 16. 匹配简化的版本号格式，如 1.2, 3.0等
	re = regexp.MustCompile(`v?(\d+\.\d+)`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到简化版本号格式: %s", matches[1])
		return matches[1]
	}

	// 17. 匹配单个数字版本，如 版本1, 版本2等
	re = regexp.MustCompile(`(?:版本|version|ver|v)[\s:：]*(\d+)`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到数字版本格式: %s", matches[1])
		return matches[1]
	}

	// 18. 匹配中文网站常见的版本表示，如 "版本：1.2.3"
	re = regexp.MustCompile(`(?:版本|ver|v)[\s:：]*(\d+\.\d+(?:\.\d+)?)`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到中文版本格式: %s", matches[1])
		return matches[1]
	}

	// 19. 匹配日期格式的版本，如 20230815
	re = regexp.MustCompile(`(20\d{2}(?:\d{2}){0,2})`)
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到日期格式版本: %s", matches[1])
		return matches[1]
	}

	// 20. 尝试从字符串中提取任何可能的数字序列作为版本，但要排除CSS样式中的数字
	re = regexp.MustCompile(`(?:^|\s|>|"|')(\d+(?:\.\d+)+)(?:\s|<|"|')`)  // 确保数字序列前后有分隔符
	matches = re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		logger.GlobalLogger.Debugf("[version_utils] 找到数字序列作为版本: %s", matches[1])
		return matches[1]
	}

	logger.GlobalLogger.Debugf("[version_utils] 未找到任何版本号格式")
	return ""
}
