package services

import (
	checkers "aur-update-checker/internal/interfaces/checkers"
	"aur-update-checker/internal/logger"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// LogService 日志服务，处理日志相关的业务逻辑
type LogService struct {
	logProvider checkers.LoggerProvider
}

// NewLogService 创建日志服务
func NewLogService(logProvider checkers.LoggerProvider) *LogService {
	return &LogService{
		logProvider: logProvider,
	}
}

// LogEntry 日志条目结构
type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

// GetLogs 获取日志（支持分页）
func (s *LogService) GetLogs(level string, page, pageSize int) (map[string]interface{}, error) {
	// 获取应用数据目录
	appDir, err := logger.EnsureAppDataDir()
	if err != nil {
		s.logProvider.Errorf("获取应用数据目录失败: %v", err)
		return nil, err
	}

	// 日志文件路径
	logPath := filepath.Join(appDir, "logs")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		// 日志目录不存在，返回空结果
		return map[string]interface{}{
			"logs": []LogEntry{},
			"total": 0,
		}, nil
	}

	// 读取日志目录中的所有文件
	files, err := os.ReadDir(logPath)
	if err != nil {
		s.logProvider.Errorf("读取日志目录失败: %v", err)
		// 如果目录不存在，返回空结果而不是错误
		if os.IsNotExist(err) {
			return map[string]interface{}{
				"logs": []LogEntry{},
				"total": 0,
			}, nil
		}
		return nil, err
	}

	// 解析所有日志文件
	logs, err := s.parseLogFiles(files, logPath, level)
	if err != nil {
		return nil, err
	}

	// 按时间倒序排序（最新的在前）
	s.sortLogsByTime(logs)

	// 计算分页
	total := len(logs)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	// 获取当前页的日志
	var pageLogs []LogEntry
	if start < total {
		pageLogs = logs[start:end]
	}

	return map[string]interface{}{
		"logs": pageLogs,
		"total": total,
	}, nil
}

// GetLatestLogs 获取最新日志（用于增量更新）
func (s *LogService) GetLatestLogs(sinceTime string, level string) ([]LogEntry, error) {
	// 获取应用数据目录
	appDir, err := logger.EnsureAppDataDir()
	if err != nil {
		s.logProvider.Errorf("获取应用数据目录失败: %v", err)
		return nil, err
	}

	// 日志文件路径
	logPath := filepath.Join(appDir, "logs")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		// 日志目录不存在，返回空数组
		return []LogEntry{}, nil
	}

	// 解析sinceTime
	var sinceTimeParsed time.Time
	if sinceTime != "" {
		// 尝试解析时间字符串
		if t, err := time.Parse("2006-01-02 15:04:05", sinceTime); err == nil {
			sinceTimeParsed = t
		} else if t, err := time.Parse("2006-01-02T15:04:05+08:00", sinceTime); err == nil {
			sinceTimeParsed = t
		} else {
			// 如果解析失败，使用当前时间
			sinceTimeParsed = time.Now()
		}
	} else {
		// 如果没有提供时间，使用当前时间
		sinceTimeParsed = time.Now()
	}

	// 读取日志目录中的所有文件
	files, err := os.ReadDir(logPath)
	if err != nil {
		s.logProvider.Errorf("读取日志目录失败: %v", err)
		// 如果目录不存在，返回空数组而不是错误
		if os.IsNotExist(err) {
			return []LogEntry{}, nil
		}
		return nil, err
	}

	// 解析所有日志文件
	logs, err := s.parseLogFiles(files, logPath, level)
	if err != nil {
		return nil, err
	}

	// 按时间倒序排序（最新的在前）
	s.sortLogsByTime(logs)

	// 过滤出比sinceTime更新的日志
	var filteredLogs []LogEntry
	for _, log := range logs {
		// 解析日志时间
		var logTime time.Time
		if t, err := time.Parse("2006-01-02 15:04:05", log.Time); err == nil {
			logTime = t
		} else if t, err := time.Parse("2006-01-02T15:04:05+08:00", log.Time); err == nil {
			logTime = t
		} else {
			// 如果解析失败，使用当前时间
			logTime = time.Now()
		}

		// 只返回比sinceTime更新的日志
		if logTime.After(sinceTimeParsed) || logTime.Equal(sinceTimeParsed) {
			filteredLogs = append(filteredLogs, log)
		}
	}

	return filteredLogs, nil
}

// ClearLogs 清空日志
func (s *LogService) ClearLogs() error {
	// 获取应用数据目录
	appDir, err := logger.EnsureAppDataDir()
	if err != nil {
		s.logProvider.Errorf("获取应用数据目录失败: %v", err)
		return err
	}

	// 日志文件路径
	logPath := filepath.Join(appDir, "logs")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		// 日志目录不存在，无需清空
		return nil
	}

	// 读取日志目录中的所有文件
	files, err := os.ReadDir(logPath)
	if err != nil {
		s.logProvider.Errorf("读取日志目录失败: %v", err)
		return err
	}

	// 删除所有日志文件
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".log" {
			filePath := filepath.Join(logPath, file.Name())
			if err := os.Remove(filePath); err != nil {
				s.logProvider.Errorf("删除日志文件失败: %v", err)
				return err
			}
		}
	}

	// 创建一个新的空日志文件，确保日志系统可以继续工作
	currentDate := time.Now().Format("2006-01-02")
	newLogFile := filepath.Join(logPath, "aur-update-checker-"+currentDate+".log")
	file, err := os.OpenFile(newLogFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		s.logProvider.Errorf("创建新日志文件失败: %v", err)
		return err
	}
	file.Close()

	return nil
}

// parseLogFiles 解析日志文件
func (s *LogService) parseLogFiles(files []os.DirEntry, logPath, level string) ([]LogEntry, error) {
	var logs []LogEntry

	// 遍历所有日志文件
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".log" {
			// 读取日志文件内容
			filePath := filepath.Join(logPath, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				s.logProvider.Errorf("读取日志文件失败: %v", err)
				continue
			}

			// 按行分割日志内容
			lines := strings.Split(string(content), "")

			// 处理每一行日志
			for _, line := range lines {
				if line == "" {
					continue
				}

				// 解析日志级别和内容
				logEntry, err := s.parseLogLine(line, file.Name())
				if err != nil {
					s.logProvider.Errorf("解析日志行失败: %v", err)
					continue
				}

				// 如果指定了日志级别，则进行过滤
				if level != "all" && logEntry.Level != level {
					continue
				}

				logs = append(logs, logEntry)
			}
		}
	}

	return logs, nil
}

// parseLogLine 解析单行日志
func (s *LogService) parseLogLine(line, fileName string) (LogEntry, error) {
	// 解析日志级别和内容
	var logLevel, message, timeStr string

	// 使用正则表达式匹配日志格式: time="2023-08-21 15:04:05" level=info msg="message"
	// 或者匹配文本格式: time=2023-08-21T15:04:05+08:00 level=info msg="message"
	re := regexp.MustCompile(`time="([^"]+)"\s+level=(\w+)\s+msg="([^"]*)"`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 4 {
		// 提取时间
		timeStr = matches[1]

		// 提取级别
		logLevel = strings.ToLower(matches[2])

		// 提取消息
		message = matches[3]
	} else {
		// 如果正则匹配失败，尝试简单匹配
		if strings.Contains(line, "[DEBUG]") {
			logLevel = "debug"
			message = strings.TrimSpace(strings.TrimPrefix(line, "[DEBUG] "))
		} else if strings.Contains(line, "[INFO]") {
			logLevel = "info"
			message = strings.TrimSpace(strings.TrimPrefix(line, "[INFO] "))
		} else if strings.Contains(line, "[WARN]") {
			logLevel = "warn"
			message = strings.TrimSpace(strings.TrimPrefix(line, "[WARN] "))
		} else if strings.Contains(line, "[ERROR]") {
			logLevel = "error"
			message = strings.TrimSpace(strings.TrimPrefix(line, "[ERROR] "))
		} else if strings.Contains(line, "[FATAL]") {
			logLevel = "fatal"
			message = strings.TrimSpace(strings.TrimPrefix(line, "[FATAL] "))
		} else {
			// 无法识别级别的日志，默认为info
			logLevel = "info"
			message = line
		}
	}

	// 如果从日志行中提取到了时间，则使用该时间
	var dateStr string
	if timeStr != "" {
		// 尝试解析时间字符串，处理不同格式
		if _, err := time.Parse("2006-01-02T15:04:05+08:00", timeStr); err == nil {
			dateStr = timeStr
		} else if _, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
			dateStr = timeStr
		} else {
			// 如果解析失败，使用当前时间
			dateStr = time.Now().Format("2006-01-02 15:04:05")
		}
	} else {
		// 尝试从文件名中提取日期
		dateStr = strings.TrimPrefix(fileName, "aur-update-checker-")
		dateStr = strings.TrimSuffix(dateStr, ".log")

		// 如果无法从文件名中提取日期，使用当前日期
		if dateStr == fileName {
			dateStr = time.Now().Format("2006-01-02")
		}
	}

	return LogEntry{
		Time:    dateStr,
		Level:   logLevel,
		Message: message,
	}, nil
}

// sortLogsByTime 按时间倒序排序日志
func (s *LogService) sortLogsByTime(logs []LogEntry) {
	// 按时间倒序排序（最新的在前）
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}
}
