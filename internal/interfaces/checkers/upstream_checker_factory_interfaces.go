package checkers

import (
	"aur-update-checker/internal/checkers/common"
	"aur-update-checker/internal/config"
	"aur-update-checker/internal/logger"
	"context"
	"fmt"
	"sync"
	"time"
)

// FactoryProvider 检查器工厂提供者接口
type FactoryProvider interface {
	// RegisterChecker 注册检查器
	RegisterChecker(name string, checker common.UpstreamChecker)

	// GetChecker 获取检查器
	GetChecker(name string) (common.UpstreamChecker, error)

	// GetAllCheckers 获取所有检查器
	GetAllCheckers() map[string]common.UpstreamChecker

	// Check 使用指定检查器检查上游版本
	Check(ctx context.Context, checkerName, url, versionExtractKey string) (string, error)

	// CheckWithOption 使用指定检查器根据选项检查上游版本
	CheckWithOption(ctx context.Context, checkerName, url, versionExtractKey string, checkTestVersion int) (string, error)

	// CheckWithVersionRef 使用指定检查器根据选项和版本引用检查上游版本
	CheckWithVersionRef(ctx context.Context, checkerName, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error)

	// GetConcurrentChecker 获取并发检查器
	GetConcurrentChecker() common.ConcurrentCheckerInterface

	// GetEnhancedConcurrentChecker 获取增强版并发检查器
	GetEnhancedConcurrentChecker() interface{}

	// AutoCheck 自动选择检查器并检查上游版本
	AutoCheck(ctx context.Context, url, versionExtractKey string) (string, string, error)

	// CheckMultiple 并发检查多个URL
	CheckMultiple(ctx context.Context, urls []string, versionExtractKey string, checkTestVersion int) interface{}

	// CheckWithConfig 使用配置执行检查
	CheckWithConfig(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, string, error)

	// ReloadConfig 重新加载配置
	ReloadConfig() error

	// ClearCache 清除检查缓存
	ClearCache()

	// SubmitAsyncCheck 提交异步检查请求
	SubmitAsyncCheck(url, versionExtractKey string, checkTestVersion int, callback func(result common.AsyncCheckResult)) (string, error)

	// GetAsyncCheckStatus 获取异步检查状态
	GetAsyncCheckStatus(id string) (string, error)

	// GetAsyncCheckResult 获取异步检查结果
	GetAsyncCheckResult(id string) (*common.AsyncCheckResult, error)

	// RemoveAsyncCheck 移除异步检查请求
	RemoveAsyncCheck(id string) error

	// ClearAsyncChecks 清除所有异步检查请求
	ClearAsyncChecks()

	// GetAsyncChecker 获取异步检查器
	GetAsyncChecker() common.AsyncCheckerInterface

	// StopAsyncChecker 停止异步检查器
	StopAsyncChecker()
}

// ConfigSelector 配置选择器接口
type ConfigSelector interface {
	// SelectCheckerWithVersionKey 根据URL和版本提取键选择检查器
	SelectCheckerWithVersionKey(url, versionExtractKey string) (common.UpstreamChecker, error)

	// SelectCheckerWithOptions 根据URL、版本提取键和选项选择检查器
	SelectCheckerWithOptions(url, versionExtractKey string, checkTestVersion int) (common.UpstreamChecker, error)

	// GetCheckerSettings 获取检查器设置
	GetCheckerSettings(checkerName string) (config.CheckerSettings, bool)

	// ReloadConfig 重新加载配置
	ReloadConfig() error

	// CheckWithConfig 使用配置执行检查
	CheckWithConfig(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error)
}

// LoggerProvider 日志提供者接口
type LoggerProvider interface {
	// Info 记录信息日志
	Info(msg string)
	// Debug 记录调试日志
	Debug(msg string)
	// Warn 记录警告日志
	Warn(msg string)
	// Error 记录错误日志
	Error(msg string)
	// Infof 记录格式化信息日志
	Infof(format string, args ...interface{})
	// Debugf 记录格式化调试日志
	Debugf(format string, args ...interface{})
	// Warnf 记录格式化警告日志
	Warnf(format string, args ...interface{})
	// Errorf 记录格式化错误日志
	Errorf(format string, args ...interface{})
}

// LoggerAdapter 适配器，将logger.Logger转换为LoggerProvider
type LoggerAdapter struct {
	logger *logger.Logger
}

// 实现LoggerProvider接口
func (a *LoggerAdapter) Info(msg string) {
	a.logger.Info(msg)
}

func (a *LoggerAdapter) Debug(msg string) {
	a.logger.Debug(msg)
}

func (a *LoggerAdapter) Warn(msg string) {
	a.logger.Warn(msg)
}

func (a *LoggerAdapter) Error(msg string) {
	a.logger.Error(msg)
}

func (a *LoggerAdapter) Infof(format string, args ...interface{}) {
	a.logger.Infof(format, args...)
}

func (a *LoggerAdapter) Debugf(format string, args ...interface{}) {
	a.logger.Debugf(format, args...)
}

func (a *LoggerAdapter) Warnf(format string, args ...interface{}) {
	a.logger.Warnf(format, args...)
}

func (a *LoggerAdapter) Errorf(format string, args ...interface{}) {
	a.logger.Errorf(format, args...)
}

// CheckerFactory 检查器工厂
type CheckerFactory struct {
	// 实现ICheckerFactory接口
	checkers          map[string]common.UpstreamChecker
	mutex             sync.RWMutex
	concurrentChecker common.ConcurrentCheckerInterface
	asyncChecker      common.AsyncCheckerInterface
	configSelector    ConfigSelector
	logProvider       LoggerProvider
}

// NewCheckerFactory 创建检查器工厂
func NewCheckerFactory() *CheckerFactory {
	// 创建适配器，将logger.Logger转换为LoggerProvider
	var logProvider LoggerProvider = &LoggerAdapter{logger.GlobalLogger}
	return NewCheckerFactoryWithLogger(logProvider)
}

// NewCheckerFactoryWithLogger 使用指定的日志提供者创建检查器工厂
func NewCheckerFactoryWithLogger(logProvider LoggerProvider) *CheckerFactory {
	logger.GlobalLogger.Info("创建检查器工厂")

	// 加载配置
	cfg := config.GetConfig()

	factory := &CheckerFactory{
		checkers:    make(map[string]common.UpstreamChecker),
		logProvider: logProvider,
	}

	// 初始化配置驱动的检查器选择器
	factory.configSelector = NewConfigCheckerSelector()

	// 初始化并发检查器，使用配置中的缓存TTL
	cacheTTL := time.Duration(cfg.Global.CacheTTL) * time.Minute
	// 使用接口方式创建并发检查器，避免直接依赖 types 包
	factory.concurrentChecker = factory.createConcurrentChecker(cacheTTL)

	// 初始化异步检查器，使用配置中的工作线程数
	logger.GlobalLogger.Infof("初始化异步检查器，工作线程数: %d", cfg.Global.AsyncWorkerCount)
	factory.asyncChecker = common.NewAsyncChecker(factory, cfg.Global.AsyncWorkerCount)
	factory.asyncChecker.Start()
	logger.GlobalLogger.Info("异步检查器初始化完成")

	// 从注册器中获取所有检查器并创建实例
	logger.GlobalLogger.Info("从注册器中获取检查器实例")
	registryAdapter := common.GetRegistry()
	checkerNames := registryAdapter.GetAll()
	logger.GlobalLogger.Infof("从注册器中获取到 %d 个检查器: %v", len(checkerNames), checkerNames)

	logger.GlobalLogger.Debug("开始实例化检查器")
	for _, name := range checkerNames {
		logger.GlobalLogger.Debugf("正在创建检查器实例: %s", name)
		checker, err := registryAdapter.Create(name)
		if err != nil {
			logger.GlobalLogger.Errorf("创建检查器 '%s' 失败: %v", name, err)
			continue
		}
		factory.RegisterChecker(name, checker)
		logger.GlobalLogger.Infof("已实例化检查器: %s", name)
	}

	// 获取工厂中所有检查器名称
	factoryCheckerNames := factory.GetAllCheckerNames()
	logger.GlobalLogger.Infof("检查器工厂创建完成，包含 %d 个检查器: %v", len(factoryCheckerNames), factoryCheckerNames)
	return factory
}

// StopAsyncChecker 停止异步检查器
func (f *CheckerFactory) StopAsyncChecker() {
	if f.asyncChecker != nil {
		f.asyncChecker.Stop()
		logger.GlobalLogger.Info("已停止异步检查器")
	}
}

// RegisterChecker 注册检查器
func (f *CheckerFactory) RegisterChecker(name string, checker common.UpstreamChecker) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.checkers[name] = checker
}

// GetAllCheckerNames 获取所有检查器名称
func (f *CheckerFactory) GetAllCheckerNames() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	names := make([]string, 0, len(f.checkers))
	for name := range f.checkers {
		names = append(names, name)
	}
	return names
}

// GetChecker 获取检查器
func (f *CheckerFactory) GetChecker(name string) (common.UpstreamChecker, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	checker, ok := f.checkers[name]
	if !ok {
		return nil, fmt.Errorf("未找到名为 '%s' 的检查器", name)
	}

	return checker, nil
}

// GetAllCheckers 获取所有检查器
func (f *CheckerFactory) GetAllCheckers() map[string]common.UpstreamChecker {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	// 返回检查器的副本
	checkers := make(map[string]common.UpstreamChecker)
	for name, checker := range f.checkers {
		checkers[name] = checker
	}

	return checkers
}

// SetConfigSelector 设置配置选择器
func (f *CheckerFactory) SetConfigSelector(selector ConfigSelector) {
	f.configSelector = selector
}

// GetConfigSelector 获取配置选择器
func (f *CheckerFactory) GetConfigSelector() ConfigSelector {
	return f.configSelector
}

// Check 使用指定检查器检查上游版本
func (f *CheckerFactory) Check(ctx context.Context, checkerName, url, versionExtractKey string) (string, error) {
	// 默认不检查测试版本
	return f.CheckWithOption(ctx, checkerName, url, versionExtractKey, 0)
}

// CheckWithOption 使用指定检查器根据选项检查上游版本
func (f *CheckerFactory) CheckWithOption(ctx context.Context, checkerName, url, versionExtractKey string, checkTestVersion int) (string, error) {
	return f.CheckWithVersionRef(ctx, checkerName, url, versionExtractKey, "", checkTestVersion)
}

// CheckWithVersionRef 使用指定检查器根据选项和版本引用检查上游版本
func (f *CheckerFactory) CheckWithVersionRef(ctx context.Context, checkerName, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	logger.GlobalLogger.Infof("使用检查器 '%s' 检查上游版本 - URL: %s, 检查测试版本: %d", checkerName, url, checkTestVersion)

	checker, err := f.GetChecker(checkerName)
	if err != nil {
		logger.GlobalLogger.Errorf("获取检查器 '%s' 失败: %v", checkerName, err)
		return "", fmt.Errorf("获取检查器失败: %v", err)
	}
	logger.GlobalLogger.Debugf("成功获取检查器: %s", checkerName)

	logger.GlobalLogger.Debugf("开始使用检查器 '%s' 检查版本", checkerName)

	// 尝试使用CheckWithVersionRef方法，如果检查器实现了该方法
	if checkerWithVersionRef, ok := checker.(interface {
		CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error)
	}); ok {
		version, err := checkerWithVersionRef.CheckWithVersionRef(ctx, url, versionExtractKey, versionRef, checkTestVersion)
		if err != nil {
			logger.GlobalLogger.Errorf("使用检查器 '%s' 检查版本失败: %v", checkerName, err)
			return "", err
		}
		logger.GlobalLogger.Infof("检查器 '%s' 成功获取版本: %s", checkerName, version)
		return version, nil
	}

	// 如果检查器没有实现CheckWithVersionRef方法，则尝试使用CheckWithOption方法
	if checkerWithOption, ok := checker.(interface {
		CheckWithOption(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error)
	}); ok {
		logger.GlobalLogger.Warnf("检查器 '%s' 未实现CheckWithVersionRef方法，将使用CheckWithOption方法", checkerName)
		version, err := checkerWithOption.CheckWithOption(ctx, url, versionExtractKey, checkTestVersion)
		if err != nil {
			logger.GlobalLogger.Errorf("使用检查器 '%s' 检查版本失败: %v", checkerName, err)
			return "", err
		}
		return version, nil
	} else {
		// 如果检查器没有实现CheckWithOption方法，则使用普通的Check方法
		logger.GlobalLogger.Warnf("检查器 '%s' 未实现CheckWithOption方法，将使用普通Check方法", checkerName)
		version, err := checker.Check(ctx, url, versionExtractKey)
		if err != nil {
			logger.GlobalLogger.Errorf("使用检查器 '%s' 检查版本失败: %v", checkerName, err)
			return "", err
		}
		logger.GlobalLogger.Infof("检查器 '%s' 成功获取版本: %s", checkerName, version)
		return version, nil
	}
}

// createConcurrentChecker 创建并发检查器
func (f *CheckerFactory) createConcurrentChecker(cacheTTL time.Duration) common.ConcurrentCheckerInterface {
	// 这是一个空实现，具体实现会在 types 包中提供
	_ = cacheTTL // 使用参数以避免未使用参数的警告
	return nil
}

// GetConcurrentChecker 获取并发检查器
func (f *CheckerFactory) GetConcurrentChecker() common.ConcurrentCheckerInterface {
	return f.concurrentChecker
}

// GetEnhancedConcurrentChecker 获取增强版并发检查器
func (f *CheckerFactory) GetEnhancedConcurrentChecker() common.EnhancedConcurrentCheckerInterface {
	// 基础检查器工厂不提供增强版并发检查器，返回nil
	return nil
}

// EnhancedCheckerFactory 增强版检查器工厂
type EnhancedCheckerFactory struct {
	*CheckerFactory
	enhancedConcurrentChecker common.EnhancedConcurrentCheckerInterface
}

// NewEnhancedCheckerFactory 创建增强版检查器工厂
func NewEnhancedCheckerFactory(baseFactory *CheckerFactory, cacheTTL time.Duration, minWorkers, maxWorkers int, adjustFactor float64) *EnhancedCheckerFactory {
	factory := &EnhancedCheckerFactory{
		CheckerFactory: baseFactory,
	}

	// 创建增强版并发检查器
	// 注意：这里不能直接创建 common.EnhancedConcurrentChecker 实例，因为它在 checkers/async 包中
	// 我们需要使用工厂函数或者依赖注入的方式来创建它
	// 这里先设置为 nil，实际使用时应该通过某种方式初始化它
	factory.enhancedConcurrentChecker = nil

	// 注意：由于 enhancedConcurrentChecker 为 nil，这里不能启动工作线程
	// 实际使用时应该在初始化 enhancedConcurrentChecker 后再启动工作线程

	return factory
}

// GetEnhancedConcurrentChecker 获取增强版并发检查器
func (f *EnhancedCheckerFactory) GetEnhancedConcurrentChecker() common.EnhancedConcurrentCheckerInterface {
	return f.enhancedConcurrentChecker
}

// Start 启动增强版检查器
func (f *EnhancedCheckerFactory) Start() {
	// 增强版检查器在创建时已经启动了工作线程，这里可以添加其他初始化逻辑
}

// Stop 停止增强版检查器
func (f *EnhancedCheckerFactory) Stop() {
	// 停止增强版并发检查器的工作线程
	if f.enhancedConcurrentChecker != nil {
		f.enhancedConcurrentChecker.Stop()
	}
}

// AutoCheck 自动选择检查器并检查上游版本
func (f *CheckerFactory) AutoCheck(ctx context.Context, url, versionExtractKey string) (string, string, error) {
	logger.GlobalLogger.Infof("自动选择检查器并检查上游版本 - URL: %s", url)

	// 使用配置选择器选择检查器
	checker, err := f.configSelector.SelectCheckerWithVersionKey(url, versionExtractKey)
	if err != nil {
		logger.GlobalLogger.Errorf("配置选择器选择检查器失败，使用默认检查器 'github': %v", err)
		// 使用默认检查器
		checkerName := "github"
		logger.GlobalLogger.Infof("使用默认检查器: %s", checkerName)

		// 使用选定的检查器检查版本
		version, err := f.Check(ctx, checkerName, url, versionExtractKey)
		if err != nil {
			logger.GlobalLogger.Errorf("使用检查器 '%s' 检查失败: %v", checkerName, err)
			return "", checkerName, fmt.Errorf("使用检查器 '%s' 检查失败: %v", checkerName, err)
		}

		logger.GlobalLogger.Infof("自动检查成功，检查器: %s, 版本: %s", checkerName, version)
		return version, checkerName, nil
	}

	checkerName := checker.Name()
	logger.GlobalLogger.Infof("配置选择器选择的检查器: %s", checkerName)

	// 获取检查器设置
	settings, ok := f.configSelector.GetCheckerSettings(checkerName)
	if ok {
		// 应用设置到检查器
		ApplyConfigToChecker(checker, settings)
	}

	// 使用选定的检查器检查版本
	version, err := checker.Check(ctx, url, versionExtractKey)
	if err != nil {
		logger.GlobalLogger.Errorf("使用检查器 '%s' 检查失败: %v", checkerName, err)
		return "", checkerName, fmt.Errorf("使用检查器 '%s' 检查失败: %v", checkerName, err)
	}

	logger.GlobalLogger.Infof("自动检查成功，检查器: %s, 版本: %s", checkerName, version)
	return version, checkerName, nil
}

// CheckMultiple 并发检查多个URL
func (f *CheckerFactory) CheckMultiple(ctx context.Context, urls []string, versionExtractKey string, checkTestVersion int) interface{} {
	logger.GlobalLogger.Infof("并发检查 %d 个URL", len(urls))
	// 使用接口方式调用并发检查器
	if checker, ok := f.concurrentChecker.(interface {
		CheckMultiple(ctx context.Context, urls []string, versionExtractKey string, checkTestVersion int) interface{}
	}); ok {
		return checker.CheckMultiple(ctx, urls, versionExtractKey, checkTestVersion)
	}
	return nil
}

// CheckWithConfig 使用配置执行检查
func (f *CheckerFactory) CheckWithConfig(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, string, error) {
	logger.GlobalLogger.Infof("使用配置执行检查 - URL: %s, 检查测试版本: %d", url, checkTestVersion)

	// 使用配置选择器执行检查
	version, err := f.configSelector.CheckWithConfig(ctx, url, versionExtractKey, checkTestVersion)
	if err != nil {
		logger.GlobalLogger.Errorf("使用配置执行检查失败: %v", err)
		return "", "", err
	}

	// 获取使用的检查器名称
	checker, err := f.configSelector.SelectCheckerWithOptions(url, versionExtractKey, checkTestVersion)
	if err != nil {
		logger.GlobalLogger.Warnf("无法获取使用的检查器: %v", err)
		return version, "unknown", nil
	}

	checkerName := checker.Name()
	logger.GlobalLogger.Infof("配置检查成功，检查器: %s, 版本: %s", checkerName, version)
	return version, checkerName, nil
}

// ReloadConfig 重新加载配置
func (f *CheckerFactory) ReloadConfig() error {
	err := f.configSelector.ReloadConfig()
	if err != nil {
		return err
	}

	// 重新加载全局配置
	cfg := config.GetConfig()

	// 重新初始化并发检查器，使用新的缓存TTL
	cacheTTL := time.Duration(cfg.Global.CacheTTL) * time.Minute
	f.concurrentChecker = f.createConcurrentChecker(cacheTTL)

	logger.GlobalLogger.Info("检查器工厂配置已重新加载")
	return nil
}

// ClearCache 清除检查缓存
func (f *CheckerFactory) ClearCache() {
	// 使用接口方式调用并发检查器的 ClearCache 方法
	if checker, ok := f.concurrentChecker.(interface {
		ClearCache()
	}); ok {
		checker.ClearCache()
	}
}

// SubmitAsyncCheck 提交异步检查请求
func (f *CheckerFactory) SubmitAsyncCheck(url, versionExtractKey string, checkTestVersion int, callback func(result common.AsyncCheckResult)) (string, error) {
	return f.asyncChecker.Submit(url, versionExtractKey, checkTestVersion, callback)
}

// GetAsyncCheckStatus 获取异步检查状态
func (f *CheckerFactory) GetAsyncCheckStatus(id string) (string, error) {
	return f.asyncChecker.GetStatus(id)
}

// GetAsyncCheckResult 获取异步检查结果
func (f *CheckerFactory) GetAsyncCheckResult(id string) (*common.AsyncCheckResult, error) {
	return f.asyncChecker.GetResult(id)
}

// RemoveAsyncCheck 移除异步检查请求
func (f *CheckerFactory) RemoveAsyncCheck(id string) error {
	return f.asyncChecker.Remove(id)
}

// ClearAsyncChecks 清除所有异步检查请求
func (f *CheckerFactory) ClearAsyncChecks() {
	f.asyncChecker.Clear()
}

// GetAsyncChecker 获取异步检查器
func (f *CheckerFactory) GetAsyncChecker() common.AsyncCheckerInterface {
	return f.asyncChecker
}
