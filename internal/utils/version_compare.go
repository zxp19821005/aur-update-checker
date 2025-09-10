package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CompareVersionStrings 比较两个版本号
// 返回值: -1表示version1小于version2, 0表示相等, 1表示version1大于version2
func CompareVersionStrings(version1, version2 string) int {
	// 规范化版本号
	v1 := NormalizeVersionString(version1)
	v2 := NormalizeVersionString(version2)

	// 提取版本组件
	v1Nums, v1Strs := ExtractVersionComponentsFromString(v1)
	v2Nums, v2Strs := ExtractVersionComponentsFromString(v2)

	// 比较数字部分
	maxLen := max(len(v1Nums), len(v2Nums))
	for i := 0; i < maxLen; i++ {
		var n1, n2 int

		if i < len(v1Nums) {
			n1 = v1Nums[i]
		}

		if i < len(v2Nums) {
			n2 = v2Nums[i]
		}

		if n1 > n2 {
			return 1
		} else if n1 < n2 {
			return -1
		}
	}

	// 比较字符串部分
	maxStrLen := max(len(v1Strs), len(v2Strs))
	for i := 0; i < maxStrLen; i++ {
		var s1, s2 string

		if i < len(v1Strs) {
			s1 = v1Strs[i]
		}

		if i < len(v2Strs) {
			s2 = v2Strs[i]
		}

		if s1 > s2 {
			return 1
		} else if s1 < s2 {
			return -1
		}
	}

	return 0
}

// NormalizeVersionString 规范化版本号
func NormalizeVersionString(version string) string {
	// 移除前后空格
	version = strings.TrimSpace(version)

	// 处理常见的版本前缀
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")
	version = strings.TrimPrefix(version, "version")
	version = strings.TrimPrefix(version, "Version")

	return version
}

// IsVersionStable 检查版本是否为稳定版本
func IsVersionStable(version string) bool {
	// 转换为小写
	lowerVersion := strings.ToLower(version)

	// 检查是否包含不稳定版本标识
	unstablePatterns := []string{
		"alpha", "beta", "rc", "preview", "dev", "test", "nightly", "snapshot",
		"a", "b", "pre",
	}

	for _, pattern := range unstablePatterns {
		if strings.Contains(lowerVersion, pattern) {
			return false
		}
	}

	return true
}

// ExtractVersionComponentsFromString 从版本字符串中提取版本组件
func ExtractVersionComponentsFromString(version string) ([]int, []string) {
	var numbers []int
	var strParts []string  // 重命名变量，避免与 strings 包冲突

	// 使用正则表达式分割版本号
	re := regexp.MustCompile(`(\d+)|([a-zA-Z]+)`)
	matches := re.FindAllStringSubmatch(version, -1)

	for _, match := range matches {
		if match[1] != "" {
			// 数字部分
			num, err := strconv.Atoi(match[1])
			if err == nil {
				numbers = append(numbers, num)
			}
		} else if match[2] != "" {
			// 字符串部分
			strParts = append(strParts, strings.ToLower(match[2]))
		}
	}

	return numbers, strParts
}

// GenerateVersionRef 生成上游版本提取参考值
func GenerateVersionRef(version string) string {
	// 规范化版本号
	version = NormalizeVersionString(version)

	// 提取版本组件
	numbers, _ := ExtractVersionComponentsFromString(version)

	// 根据版本组件数量生成参考值
	if len(numbers) >= 3 {
		return "a.b.c.d"
	} else if len(numbers) >= 2 {
		return "a.b.c"
	} else if len(numbers) >= 1 {
		return "a.b"
	}

	// 如果没有数字部分，则使用原始版本
	return version
}

// ExtractUpstreamVersions 从上游URL提取版本信息
// 注意：此函数已弃用，请使用 services.UpstreamService 中的 getUpstreamVersions 方法
func ExtractUpstreamVersions(upstreamUrl, versionExtractKey, versionRef string) ([]UpstreamVersion, error) {
	// 此函数已弃用，不再实现
	return nil, fmt.Errorf("此函数已弃用，请使用 services.UpstreamService 中的 getUpstreamVersions 方法")
}

// UpstreamVersion 上游版本信息
type UpstreamVersion struct {
	Version     string `json:"version"`
	IsPrerelease bool  `json:"isPrerelease"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	DownloadURL string `json:"downloadUrl,omitempty"`
}



// ParseReleaseDate 解析发布日期
func ParseReleaseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now()
	}

	// 尝试不同的日期格式
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"Jan 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	// 如果无法解析，返回当前时间
	return time.Now()
}

// Now 返回当前时间
func Now() time.Time {
	return time.Now()
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
