package common

import (
	"context"
	"sync"
)

// UpstreamChecker 上游检查器接口
type UpstreamChecker interface {
	// Name 获取检查器名称
	Name() string

	// Supports 检查是否支持给定的URL
	Supports(url string) bool

	// Priority 获取检查器优先级
	Priority() int

	// Check 检查上游版本
	Check(ctx context.Context, url, versionExtractKey string) (string, error)

	// CheckWithOption 带选项地检查上游版本
	CheckWithOption(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error)

	// CheckWithVersionRef 带选项和版本引用地检查上游版本
	CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error)
}

// UpstreamCheckerRegistry 上游检查器注册器，用于管理所有可用的检查器
type UpstreamCheckerRegistry struct {
	checkers map[string]func() UpstreamChecker // 存储检查器名称和构造函数的映射，将在Register和Get方法中使用
	// nolint:unused // 这个字段将在Register和Get方法的完整实现中使用
	mutex    sync.RWMutex                     // 用于保护并发访问的互斥锁，将在Register和Get方法中使用
	// nolint:unused // 这个字段将在Register和Get方法的完整实现中使用
}

// GetRegistry 获取全局检查器注册器实例
// 此函数在 registry_adapter.go 中实现
// func GetRegistry() *UpstreamCheckerRegistry

// Register 注册检查器
func (r *UpstreamCheckerRegistry) Register(name string, constructor func() UpstreamChecker) {
	// 这个方法将在 checkers 包中实现
}

// Get 获取检查器构造函数
func (r *UpstreamCheckerRegistry) Get(name string) (func() UpstreamChecker, bool) {
	// 这个方法将在 checkers 包中实现
	return nil, false
}

// GetAll 获取所有检查器名称
func (r *UpstreamCheckerRegistry) GetAll() []string {
	// 这个方法将在 checkers 包中实现
	return nil
}

// Create 创建指定名称的检查器实例
func (r *UpstreamCheckerRegistry) Create(name string) (UpstreamChecker, error) {
	// 这个方法将在 checkers 包中实现
	return nil, nil
}

// RegisterChecker 注册检查器的便捷函数
// 此函数在 registry_adapter.go 中实现
// func RegisterChecker(name string, constructor func() UpstreamChecker)
