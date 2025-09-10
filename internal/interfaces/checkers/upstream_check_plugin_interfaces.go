package checkers

import (
	"context"
	"errors"
	"fmt"
	"plugin"
	"sync"

	"aur-update-checker/internal/logger"
	"aur-update-checker/internal/checkers/common"
)

// UpstreamChecker 上游检查器接口已在 upstream_base_checker.go 中定义

// PluginChecker 插件检查器接口
// 插件需要实现此接口才能被系统识别和使用
type PluginChecker interface {
	// UpstreamChecker 嵌入上游检查器接口
	UpstreamChecker

	// PluginInfo 返回插件信息
	PluginInfo() PluginInfo
}

// PluginInfo 插件信息
type PluginInfo struct {
	// Name 插件名称
	Name string
	// Version 插件版本
	Version string
	// Author 插件作者
	Author string
	// Description 插件描述
	Description string
}

// PluginLoader 插件加载器接口
type PluginLoader interface {
	// Load 加载插件
	Load(path string) (PluginChecker, error)

	// Unload 卸载插件
	Unload(name string) error
}

// CheckerRegistry 检查器注册器
type CheckerRegistry struct {
	mutex    sync.RWMutex
	checkers map[string]func() UpstreamChecker
}

// Register 注册检查器
func (r *CheckerRegistry) Register(name string, factory func() UpstreamChecker) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.checkers[name] = factory
}

// Create 创建指定名称的检查器
func (r *CheckerRegistry) Create(name string) (UpstreamChecker, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	factory, exists := r.checkers[name]
	if !exists {
		return nil, fmt.Errorf("检查器 '%s' 未注册", name)
	}

	return factory(), nil
}

// GetAll 获取所有检查器名称
func (r *CheckerRegistry) GetAll() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.checkers))
	for name := range r.checkers {
		names = append(names, name)
	}
	return names
}

// AutoSelect 根据URL自动选择合适的检查器
func (r *CheckerRegistry) AutoSelect(url string) (UpstreamChecker, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// 简单的自动选择逻辑，可以根据URL模式选择检查器
	// 这里只是一个示例实现，实际使用时可能需要更复杂的逻辑

	// 默认返回一个通用的检查器
	for name, factory := range r.checkers {
		logger.GlobalLogger.Debugf("为URL '%s' 自动选择检查器 '%s'", url, name)
		return factory(), nil
	}

	return nil, fmt.Errorf("没有可用的检查器")
}

// CheckerRegistryAdapter 检查器注册器适配器
type CheckerRegistryAdapter struct {
	registry common.RegistryInterface
	// 为了支持插件加载器的功能，我们需要在适配器中维护一个额外的检查器映射
	pluginCheckers map[string]func() UpstreamChecker
	mutex          sync.RWMutex
}

// GetAll 获取所有检查器名称
func (a *CheckerRegistryAdapter) GetAll() []string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	// 获取内置检查器名称
	builtinNames := a.registry.GetAll()
	
	// 创建结果切片，预分配足够空间
	names := make([]string, 0, len(builtinNames)+len(a.pluginCheckers))
	
	// 添加内置检查器名称
	names = append(names, builtinNames...)
	
	// 添加插件检查器名称
	for name := range a.pluginCheckers {
		names = append(names, name)
	}
	
	return names
}

// Create 创建检查器实例
func (a *CheckerRegistryAdapter) Create(name string) (UpstreamChecker, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	
	// 首先尝试从插件检查器中创建
	if constructor, ok := a.pluginCheckers[name]; ok {
		return constructor(), nil
	}
	
	// 然后尝试从内置检查器中创建
	constructor, ok := a.registry.Get(name)
	if !ok {
		return nil, fmt.Errorf("未找到名为 '%s' 的检查器", name)
	}
	
	// 创建检查器实例
	checker := constructor()
	return &UpstreamCheckerAdapter{checker}, nil
}

// Register 注册检查器
func (a *CheckerRegistryAdapter) Register(name string, constructor func() UpstreamChecker) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	if a.pluginCheckers == nil {
		a.pluginCheckers = make(map[string]func() UpstreamChecker)
	}
	
	a.pluginCheckers[name] = constructor
}

// AutoSelect 根据URL自动选择最合适的检查器
// 它会遍历所有已注册的检查器，包括内置检查器和插件检查器，找出支持该URL且优先级最高的检查器
func (a *CheckerRegistryAdapter) AutoSelect(url string) (UpstreamChecker, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	
	var selectedChecker UpstreamChecker
	highestPriority := -1
	
	// 首先检查内置检查器
	builtinNames := a.registry.GetAll()
	for _, name := range builtinNames {
		constructor, ok := a.registry.Get(name)
		if !ok {
			continue
		}
		
		// 创建检查器实例
		checker := constructor()
		adapter := &UpstreamCheckerAdapter{checker}
		
		if adapter.Supports(url) && adapter.Priority() > highestPriority {
			highestPriority = adapter.Priority()
			selectedChecker = adapter
		}
	}
	
	// 然后检查插件检查器
	for _, constructor := range a.pluginCheckers {
		checker := constructor()
		if checker.Supports(url) && checker.Priority() > highestPriority {
			highestPriority = checker.Priority()
			selectedChecker = checker
		}
	}
	
	if selectedChecker == nil {
		return nil, fmt.Errorf("未找到支持URL '%s' 的检查器", url)
	}
	
	return selectedChecker, nil
}

// UpstreamCheckerAdapter 上游检查器适配器，用于适配checkers.UpstreamChecker
type UpstreamCheckerAdapter struct {
	checker UpstreamChecker
}

// 实现checkers.UpstreamChecker接口
func (a *UpstreamCheckerAdapter) Check(ctx context.Context, url, versionExtractKey string) (string, error) {
	return a.checker.Check(ctx, url, versionExtractKey)
}

func (a *UpstreamCheckerAdapter) CheckWithOption(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error) {
	return a.checker.CheckWithOption(ctx, url, versionExtractKey, checkTestVersion)
}

func (a *UpstreamCheckerAdapter) CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	// 尝试将底层检查器转换为带有CheckWithVersionRef方法的接口
	if checkerWithVersionRef, ok := a.checker.(interface {
		CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error)
	}); ok {
		return checkerWithVersionRef.CheckWithVersionRef(ctx, url, versionExtractKey, versionRef, checkTestVersion)
	}
	
	// 如果底层检查器没有实现CheckWithVersionRef方法，则回退到CheckWithOption
	return a.checker.CheckWithOption(ctx, url, versionExtractKey, checkTestVersion)
}

func (a *UpstreamCheckerAdapter) Name() string {
	return a.checker.Name()
}

func (a *UpstreamCheckerAdapter) Supports(url string) bool {
	return a.checker.Supports(url)
}

func (a *UpstreamCheckerAdapter) Priority() int {
	return a.checker.Priority()
}

// DefaultPluginLoader 默认插件加载器实现
type DefaultPluginLoader struct {
	registry *CheckerRegistryAdapter
}

// GetPluginRegistry 获取插件检查器注册器
func GetPluginRegistry() *CheckerRegistryAdapter {
	logger.GlobalLogger.Info("创建检查器注册器适配器")
	
	// 创建一个适配器
	adapter := &CheckerRegistryAdapter{
		registry:       common.GetRegistry(),
		pluginCheckers: make(map[string]func() UpstreamChecker),
	}
	
	// 获取所有检查器名称
	checkerNames := adapter.GetAll()
	logger.GlobalLogger.Infof("检查器注册器适配器创建完成，包含 %d 个检查器: %v", len(checkerNames), checkerNames)
	
	return adapter
}

// NewDefaultPluginLoader 创建默认插件加载器
func NewDefaultPluginLoader() *DefaultPluginLoader {
	return &DefaultPluginLoader{
		registry: GetPluginRegistry(),
	}
}

// Load 实现插件加载
func (l *DefaultPluginLoader) Load(path string) (PluginChecker, error) {
	// 加载插件
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	// 获取插件符号
	sym, err := p.Lookup("NewPluginChecker")
	if err != nil {
		return nil, err
	}

	// 类型断言检查
	newFunc, ok := sym.(func() PluginChecker)
	if !ok {
		return nil, err
	}

	// 创建插件实例
	checker := newFunc()
	info := checker.PluginInfo()

	// 注册到注册器
	l.registry.Register(info.Name, func() UpstreamChecker {
		return checker
	})

	return checker, nil
}

// Unload 实现插件卸载
func (l *DefaultPluginLoader) Unload(name string) error {
	// 在Go中，插件一旦加载就无法真正卸载
	// 这里我们只是从注册器中移除它
	l.registry.mutex.Lock()
	defer l.registry.mutex.Unlock()

	if _, ok := l.registry.pluginCheckers[name]; ok {
		delete(l.registry.pluginCheckers, name)
		return nil
	}

	return errors.New("plugin not found")
}

// PluginManager 插件管理器
type PluginManager struct {
	loaders map[string]PluginLoader
	plugins map[string]PluginChecker
}

// NewPluginManager 创建插件管理器
func NewPluginManager() *PluginManager {
	return &PluginManager{
		loaders: make(map[string]PluginLoader),
		plugins: make(map[string]PluginChecker),
	}
}

// RegisterLoader 注册插件加载器
func (m *PluginManager) RegisterLoader(name string, loader PluginLoader) {
	m.loaders[name] = loader
}

// LoadPlugin 加载插件
func (m *PluginManager) LoadPlugin(loaderName, path string) (PluginChecker, error) {
	loader, ok := m.loaders[loaderName]
	if !ok {
		return nil, errors.New("loader not found")
	}

	checker, err := loader.Load(path)
	if err != nil {
		return nil, err
	}

	info := checker.PluginInfo()
	m.plugins[info.Name] = checker

	return checker, nil
}

// UnloadPlugin 卸载插件
func (m *PluginManager) UnloadPlugin(name string) error {
	_, ok := m.plugins[name]
	if !ok {
		return errors.New("plugin not found")
	}

	// 使用默认加载器卸载
	loader := NewDefaultPluginLoader()
	err := loader.Unload(name)
	if err != nil {
		return err
	}

	delete(m.plugins, name)
	return nil
}

// GetPlugin 获取插件
func (m *PluginManager) GetPlugin(name string) (PluginChecker, bool) {
	checker, ok := m.plugins[name]
	return checker, ok
}

// ListPlugins 列出所有已加载插件
func (m *PluginManager) ListPlugins() []PluginInfo {
	infos := make([]PluginInfo, 0, len(m.plugins))
	for _, checker := range m.plugins {
		infos = append(infos, checker.PluginInfo())
	}
	return infos
}

// 全局插件管理器实例
var (
	globalPluginManager *PluginManager
	pluginOnce          sync.Once
)

// GetPluginManager 获取全局插件管理器
func GetPluginManager() *PluginManager {
	pluginOnce.Do(func() {
		globalPluginManager = NewPluginManager()
		// 注册默认加载器
		globalPluginManager.RegisterLoader("default", NewDefaultPluginLoader())
	})
	return globalPluginManager
}
