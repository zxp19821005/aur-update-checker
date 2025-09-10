package common

import (
	"context"
)

// FactoryProvider 检查器工厂提供者接口
type FactoryProvider interface {
	// RegisterChecker 注册检查器
	RegisterChecker(name string, checker UpstreamChecker)

	// GetChecker 获取检查器
	GetChecker(name string) (UpstreamChecker, error)

	// GetAllCheckers 获取所有检查器
	GetAllCheckers() map[string]UpstreamChecker

	// Check 使用指定检查器检查上游版本
	Check(ctx context.Context, checkerName, url, versionExtractKey string) (string, error)

	// CheckWithOption 使用指定检查器根据选项检查上游版本
	CheckWithOption(ctx context.Context, checkerName, url, versionExtractKey string, checkTestVersion int) (string, error)

	// GetConcurrentChecker 获取并发检查器
	GetConcurrentChecker() ConcurrentCheckerInterface
}

// ConcurrentCheckerInterface 并发检查器接口
type ConcurrentCheckerInterface interface {
	CheckSingle(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error)
}
