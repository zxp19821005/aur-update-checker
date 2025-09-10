package checkers

import (
	"aur-update-checker/internal/checkers/common"
	"context"
)

// Cache 缓存接口
type Cache interface {
	Get(key string) (string, bool)
	Set(key, value string)
}

// ConcurrentCheckerInterface 并发检查器接口
type ConcurrentCheckerInterface interface {
	CheckSingle(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error)
}

// UpstreamChecker 上游检查器接口
// 使用 common 包中定义的 UpstreamChecker 接口
type UpstreamChecker = common.UpstreamChecker

// FactoryProvider 定义获取检查器的接口
// 使用 upstream_checker_factory_interfaces.go 中定义的 FactoryProvider 接口
