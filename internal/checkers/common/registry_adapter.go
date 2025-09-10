package common

import (
	"sync"
)

// RegistryInterface 注册器接口，定义注册器的基本操作
type RegistryInterface interface {
	// Register 注册检查器
	Register(name string, constructor func() UpstreamChecker)

	// Get 获取检查器构造函数
	Get(name string) (func() UpstreamChecker, bool)

	// GetAll 获取所有检查器名称
	GetAll() []string

	// Create 创建指定名称的检查器实例
	Create(name string) (UpstreamChecker, error)
}

// registryAdapter 注册器适配器，将外部注册器适配到 common 包
type registryAdapter struct {
	registry RegistryInterface
}

var (
	// globalAdapter 全局注册器适配器实例
	globalAdapter *registryAdapter
	// once 确保适配器只初始化一次
	once sync.Once
	// registryProvider 注册器提供者，用于设置外部注册器
	registryProvider RegistryInterface
)

// SetRegistryProvider 设置注册器提供者
func SetRegistryProvider(provider RegistryInterface) {
	registryProvider = provider
}

// GetRegistry 获取全局检查器注册器适配器实例
func GetRegistry() *registryAdapter {
	once.Do(func() {
		globalAdapter = &registryAdapter{
			registry: registryProvider,
		}
	})
	return globalAdapter
}

// Register 注册检查器
func (r *registryAdapter) Register(name string, constructor func() UpstreamChecker) {
	if r.registry != nil {
		r.registry.Register(name, constructor)
	}
}

// Get 获取检查器构造函数
func (r *registryAdapter) Get(name string) (func() UpstreamChecker, bool) {
	if r.registry == nil {
		return nil, false
	}
	return r.registry.Get(name)
}

// GetAll 获取所有检查器名称
func (r *registryAdapter) GetAll() []string {
	if r.registry == nil {
		return []string{}
	}
	return r.registry.GetAll()
}

// Create 创建指定名称的检查器实例
func (r *registryAdapter) Create(name string) (UpstreamChecker, error) {
	if r.registry == nil {
		return nil, nil
	}
	return r.registry.Create(name)
}

// RegisterChecker 注册检查器的便捷函数
func RegisterChecker(name string, constructor func() UpstreamChecker) {
	GetRegistry().Register(name, constructor)
}
