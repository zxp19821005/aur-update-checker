package checkers

import (
	"fmt"
	"sync"

	"aur-update-checker/internal/checkers/common"
	"aur-update-checker/internal/logger"
)

// UpstreamCheckerRegistry 上游检查器注册器，用于管理所有可用的检查器
type UpstreamCheckerRegistry struct {
	checkers map[string]func() common.UpstreamChecker
	mutex    sync.RWMutex
}

var (
	// globalRegistry 全局检查器注册器实例
	globalRegistry *UpstreamCheckerRegistry
	// once 确保注册器只初始化一次
	once sync.Once
)

// GetRegistry 获取全局检查器注册器实例
func GetRegistry() *UpstreamCheckerRegistry {
	once.Do(func() {
		globalRegistry = &UpstreamCheckerRegistry{
			checkers: make(map[string]func() common.UpstreamChecker),
		}
	})
	return globalRegistry
}

// Register 注册检查器
func (r *UpstreamCheckerRegistry) Register(name string, constructor func() common.UpstreamChecker) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.checkers[name]; ok {
		// 检查器已存在，将被覆盖
		if ok {
			logger.GlobalLogger.Debugf("检查器 %s 已存在，将被覆盖", name)
		}
	}

	r.checkers[name] = constructor
}

// Get 获取检查器构造函数
func (r *UpstreamCheckerRegistry) Get(name string) (func() common.UpstreamChecker, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	constructor, ok := r.checkers[name]
	return constructor, ok
}

// GetAll 获取所有检查器名称
func (r *UpstreamCheckerRegistry) GetAll() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.checkers))
	for name := range r.checkers {
		names = append(names, name)
	}
	return names
}

// Create 创建指定名称的检查器实例
func (r *UpstreamCheckerRegistry) Create(name string) (common.UpstreamChecker, error) {
	constructor, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("未找到名为 '%s' 的检查器", name)
	}
	return constructor(), nil
}

// RegisterChecker 注册检查器的便捷函数
func RegisterChecker(name string, constructor func() common.UpstreamChecker) {
	GetRegistry().Register(name, constructor)
}

// init 初始化函数，注册所有内置检查器
func init() {
	// 确保日志系统已初始化
	if logger.GlobalLogger == nil {
		logger.InitLogger()
	}

	logger.GlobalLogger.Info("开始初始化上游检查器注册器")
	
	// 设置注册器提供者
	common.SetRegistryProvider(&registryProvider{})

	// 注册所有内置检查器
	RegisterChecker("redirect", func() common.UpstreamChecker { return NewRedirectChecker() })
	logger.GlobalLogger.Debug("已注册检查器: redirect")

	RegisterChecker("github", func() common.UpstreamChecker { return NewGitHubChecker() })
	logger.GlobalLogger.Debug("已注册检查器: github")

	RegisterChecker("gitee", func() common.UpstreamChecker { return NewGiteeChecker() })
	logger.GlobalLogger.Debug("已注册检查器: gitee")

	RegisterChecker("gitlab", func() common.UpstreamChecker { return NewGitLabChecker() })
	logger.GlobalLogger.Debug("已注册检查器: gitlab")

	RegisterChecker("json", func() common.UpstreamChecker { return NewJsonChecker() })
	logger.GlobalLogger.Debug("已注册检查器: json")

	RegisterChecker("npm", func() common.UpstreamChecker { return NewNpmChecker() })
	logger.GlobalLogger.Debug("已注册检查器: npm")

	RegisterChecker("pypi", func() common.UpstreamChecker { return NewPyPIChecker() })
	logger.GlobalLogger.Debug("已注册检查器: pypi")

	RegisterChecker("curl", func() common.UpstreamChecker { return NewCurlChecker() })
	logger.GlobalLogger.Debug("已注册检查器: curl")

	RegisterChecker("http", func() common.UpstreamChecker { return NewHttpChecker() })
	logger.GlobalLogger.Debug("已注册检查器: http")

	RegisterChecker("playwright", func() common.UpstreamChecker { return NewPlaywrightChecker() })
	logger.GlobalLogger.Debug("已注册检查器: playwright")

	logger.GlobalLogger.Info("上游检查器注册器初始化完成")
}

// registryProvider 注册器提供者，实现 common.RegistryInterface 接口
type registryProvider struct{}

// Register 注册检查器
func (p *registryProvider) Register(name string, constructor func() common.UpstreamChecker) {
	RegisterChecker(name, constructor)
}

// Get 获取检查器构造函数
func (p *registryProvider) Get(name string) (func() common.UpstreamChecker, bool) {
	constructor, ok := globalRegistry.checkers[name]
	if !ok {
		return nil, false
	}
	// 将 checkers.UpstreamChecker 转换为 common.UpstreamChecker
	return func() common.UpstreamChecker {
		return constructor()
	}, true
}

// GetAll 获取所有检查器名称
func (p *registryProvider) GetAll() []string {
	return globalRegistry.GetAll()
}

// Create 创建指定名称的检查器实例
func (p *registryProvider) Create(name string) (common.UpstreamChecker, error) {
	return globalRegistry.Create(name)
}