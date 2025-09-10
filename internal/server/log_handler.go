package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"aur-update-checker/internal/logger"
)

// getLogs 获取日志
func (s *APIServer) getLogs(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	if level == "" {
		level = "all"
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize == 0 {
		pageSize = 100
	}

	// 获取应用数据目录
	appDir, err := logger.EnsureAppDataDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 日志文件路径
	logPath := filepath.Join(appDir, "logs")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		// 日志目录不存在，返回空结果
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"logs": []map[string]interface{}{},
			"total": 0,
		})
		return
	}

	// 读取日志目录中的所有文件
	files, err := os.ReadDir(logPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var logs []map[string]interface{}

	// 遍历所有日志文件
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".log" {
			// 读取日志文件内容
			filePath := filepath.Join(logPath, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
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
				var logLevel, message, timeStr string
				// 使用正则表达式匹配日志格式: [LEVEL] time message
				re := regexp.MustCompile(`^\[(DEBUG|INFO|WARN|ERROR|FATAL|PANIC)\] (\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) (.+)$`)
				matches := re.FindStringSubmatch(line)
				if len(matches) >= 4 {
					logLevel = strings.ToLower(matches[1])
					timeStr = matches[2]
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

				// 如果指定了日志级别，则进行过滤
				if level != "all" && logLevel != level {
					continue
				}

				// 添加到日志列表
				logs = append(logs, map[string]interface{}{
					"time":    timeStr,
					"level":   logLevel,
					"message": message,
				})
			}
		}
	}

	// 按时间倒序排序（最新的在前）
	// 由于日志是按文件和行顺序读取的，需要反转
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

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
	var pageLogs []map[string]interface{}
	if start < total {
		pageLogs = logs[start:end]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs": pageLogs,
		"total": total,
	})
}

// getLatestLogs 获取最新日志
func (s *APIServer) getLatestLogs(w http.ResponseWriter, r *http.Request) {
	sinceTime := r.URL.Query().Get("sinceTime")
	level := r.URL.Query().Get("level")
	if level == "" {
		level = "all"
	}

	// 获取应用数据目录
	appDir, err := logger.EnsureAppDataDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 日志文件路径
	logPath := filepath.Join(appDir, "logs")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		// 日志目录不存在，返回空数组
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var logs []map[string]interface{}

	// 遍历所有日志文件
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".log" {
			// 读取日志文件内容
			filePath := filepath.Join(logPath, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
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
				var logLevel, message, timeStr string
				// 使用正则表达式匹配日志格式: [LEVEL] time message
				re := regexp.MustCompile(`^\[(DEBUG|INFO|WARN|ERROR|FATAL|PANIC)\] (\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) (.+)$`)
				matches := re.FindStringSubmatch(line)
				if len(matches) >= 4 {
					logLevel = strings.ToLower(matches[1])
					timeStr = matches[2]
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

				// 如果指定了日志级别，则进行过滤
				if level != "all" && logLevel != level {
					continue
				}

				// 解析日志时间，只返回比sinceTime更新的日志
				var logTime time.Time
				if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
					logTime = t
				} else if t, err := time.Parse("2006-01-02T15:04:05+08:00", timeStr); err == nil {
					logTime = t
				} else {
					// 如果解析失败，使用当前时间
					logTime = time.Now()
				}

				// 只返回比sinceTime更新的日志
				if logTime.After(sinceTimeParsed) || logTime.Equal(sinceTimeParsed) {
					logs = append(logs, map[string]interface{}{
						"time":    timeStr,
						"level":   logLevel,
						"message": message,
					})
				}
			}
		}
	}

	// 按时间倒序排序（最新的在前）
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// clearLogs 清除日志
func (s *APIServer) clearLogs(w http.ResponseWriter, r *http.Request) {
	// 获取应用数据目录
	appDir, err := logger.EnsureAppDataDir()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 日志文件路径
	logPath := filepath.Join(appDir, "logs")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		// 日志目录不存在，无需清空
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}

	// 读取日志目录中的所有文件
	files, err := os.ReadDir(logPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 删除所有日志文件
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".log" {
			filePath := filepath.Join(logPath, file.Name())
			if err := os.Remove(filePath); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	// 创建一个新的空日志文件，确保日志系统可以继续工作
	currentDate := time.Now().Format("2006-01-02")
	newLogFile := filepath.Join(logPath, fmt.Sprintf("aur-update-checker-%s.log", currentDate))
	file, err := os.OpenFile(newLogFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	file.Close()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
