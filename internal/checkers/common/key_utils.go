package common

import (
	"aur-update-checker/internal/logger"
	"strings"
)

// FindAllKeyPositions 查找键在内容中的所有位置
// 这是一个公共函数，可以被多个检查器共享使用
func FindAllKeyPositions(content, key string) []int {
	var positions []int

	index := 0
	for {
		pos := strings.Index(content[index:], key)
		if pos == -1 {
			break
		}

		// 计算绝对位置
		absPos := index + pos
		positions = append(positions, absPos)

		// 移动到下一个位置继续搜索
		index = absPos + len(key)
	}

	return positions
}

// FindCombinedKeys 查找多个键的组合
// 这是一个公共函数，可以被多个检查器共享使用
func FindCombinedKeys(content string, keys []string) []string {
	var results []string

	logger.GlobalLogger.Debugf("[key_utils] 开始查找复合键组合，第一个键: %s", keys[0])

	// 查找第一个键的所有位置
	firstKeyPositions := FindAllKeyPositions(content, keys[0])
	logger.GlobalLogger.Debugf("[key_utils] 第一个键 '%s' 找到 %d 个位置", keys[0], len(firstKeyPositions))

	// 对于每个第一个键的位置，查找附近是否有其他键
	for i, pos := range firstKeyPositions {
		logger.GlobalLogger.Debugf("[key_utils] 检查第一个键的第 %d 个位置，位置: %d", i+1, pos)

		// 定义搜索范围（前后200个字符）
		searchStart := pos - 200
		if searchStart < 0 {
			searchStart = 0
		}
		searchEnd := pos + len(keys[0]) + 200
		if searchEnd > len(content) {
			searchEnd = len(content)
		}

		searchArea := content[searchStart:searchEnd]
		logger.GlobalLogger.Debugf("[key_utils] 搜索区域长度: %d", len(searchArea))

		// 检查搜索区域是否包含所有其他键
		allKeysFound := true
		for _, key := range keys[1:] {
			if strings.Contains(searchArea, key) {
				logger.GlobalLogger.Debugf("[key_utils] 在搜索区域找到键: %s", key)
			} else {
				logger.GlobalLogger.Debugf("[key_utils] 在搜索区域未找到键: %s", key)
				allKeysFound = false
				break
			}
		}

		if allKeysFound {
			logger.GlobalLogger.Debugf("[key_utils] 找到包含所有键的区域")
			// 提取包含所有键的区域
			results = append(results, searchArea)
		}
	}

	logger.GlobalLogger.Debugf("[key_utils] 复合键搜索完成，找到 %d 个结果", len(results))
	return results
}
