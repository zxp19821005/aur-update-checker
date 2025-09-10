package common

import "time"

// AsyncCheckResult 异步检查结果
type AsyncCheckResult struct {
	ID         string
	URL        string
	Version    string
	Error      error
	Status     string
	CreateTime time.Time
	UpdateTime time.Time
	CreatedAt  time.Time
	Duration   time.Duration
}

// AsyncCheckRequest 异步检查请求
type AsyncCheckRequest struct {
	ID                string
	URL               string
	VersionExtractKey string
	CheckTestVersion  int
	Status            string // "pending", "completed", "failed"
	Result            string
	Error             error
	CreatedAt         time.Time
	CompletedAt       *time.Time
	Callback          func(result AsyncCheckResult)
}

// AsyncCheckerStats 异步检查器统计信息
type AsyncCheckerStats struct {
	TotalRequests     int64
	CompletedRequests int64
	FailedRequests   int64
	AverageTime      time.Duration
}

// CheckResult 检查结果
type CheckResult struct {
	URL      string
	Version  string
	Error    error
	Duration time.Duration
}
