package common

import (
	"context"
)

// AsyncCheckerInterface 异步检查器接口
type AsyncCheckerInterface interface {
	// Submit 提交异步检查请求
	Submit(url, versionExtractKey string, checkTestVersion int, callback func(result AsyncCheckResult)) (string, error)

	// GetResult 获取异步检查结果
	GetResult(id string) (*AsyncCheckResult, error)

	// Start 启动异步检查器
	Start()

	// Stop 停止异步检查器
	Stop()

	// Clear 清除所有异步检查请求
	Clear()

	// GetStats 获取异步检查器统计信息
	GetStats() AsyncCheckerStats

	// SetMaxPending 设置最大待处理请求数
	SetMaxPending(max int)

	// SetWorkerCount 设置工作线程数
	SetWorkerCount(count int)

	// GetStatus 获取检查状态
	GetStatus(id string) (string, error)

	// Remove 移除检查请求
	Remove(id string) error
}

// AsyncCheckRequest 和 AsyncCheckerStats 类型定义已移至 async_types.go 文件中

// EnhancedConcurrentCheckerInterface 增强版并发检查器接口
type EnhancedConcurrentCheckerInterface interface {
	// Start 启动增强版并发检查器
	Start()
	
	// Stop 停止增强版并发检查器
	Stop()
	
	// CheckMultipleWithDynamicConcurrency 动态并发检查多个URL
	CheckMultipleWithDynamicConcurrency(ctx context.Context, requests []AsyncCheckRequest) []AsyncCheckResult
}
