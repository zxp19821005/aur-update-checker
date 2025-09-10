package container

import (
	checkers "aur-update-checker/internal/interfaces/checkers"
	"aur-update-checker/internal/checkers/common"
	"aur-update-checker/internal/config"
	"aur-update-checker/internal/database"
	"aur-update-checker/internal/logger"
	"context"
	"database/sql"
	"sync"
)

// Container 依赖注入容器
type Container struct {
	db                 *sql.DB
	log                *logger.Logger
	config             *config.Config
	checkerFactory     checkers.FactoryProvider
	configSelector     checkers.ConfigSelector
	once               sync.Once
}

// CheckerFactoryAdapter 检查器工厂适配器
type CheckerFactoryAdapter struct {
	factory *checkers.CheckerFactory
}

// 实现checkers.ICheckerFactory接口
func (a *CheckerFactoryAdapter) RegisterChecker(name string, checker checkers.UpstreamChecker) {
	// 转换为checkers.UpstreamChecker
	c := &UpstreamCheckerAdapter{checker}
	a.factory.RegisterChecker(name, c)
}

func (a *CheckerFactoryAdapter) GetChecker(name string) (checkers.UpstreamChecker, error) {
	checker, err := a.factory.GetChecker(name)
	if err != nil {
		return nil, err
	}
	return &UpstreamCheckerAdapter{checker}, nil
}

func (a *CheckerFactoryAdapter) GetAllCheckers() map[string]checkers.UpstreamChecker {
	checkersMap := a.factory.GetAllCheckers()
	result := make(map[string]checkers.UpstreamChecker)
	for name, checker := range checkersMap {
		result[name] = &UpstreamCheckerAdapter{checker: checker}
	}
	return result
}

func (a *CheckerFactoryAdapter) Check(ctx context.Context, checkerName, url, versionExtractKey string) (string, error) {
	return a.factory.Check(ctx, checkerName, url, versionExtractKey)
}

func (a *CheckerFactoryAdapter) CheckWithOption(ctx context.Context, checkerName, url, versionExtractKey string, checkTestVersion int) (string, error) {
	return a.factory.CheckWithOption(ctx, checkerName, url, versionExtractKey, checkTestVersion)
}

func (a *CheckerFactoryAdapter) CheckWithVersionRef(ctx context.Context, checkerName, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error) {
	return a.factory.CheckWithVersionRef(ctx, checkerName, url, versionExtractKey, versionRef, checkTestVersion)
}

func (a *CheckerFactoryAdapter) GetConcurrentChecker() common.ConcurrentCheckerInterface {
	// 返回并发检查器
	return a.factory.GetConcurrentChecker()
}

func (a *CheckerFactoryAdapter) GetEnhancedConcurrentChecker() interface{} {
	// 转换为interface{}
	return a.factory.GetEnhancedConcurrentChecker()
}

func (a *CheckerFactoryAdapter) AutoCheck(ctx context.Context, url, versionExtractKey string) (string, string, error) {
	return a.factory.AutoCheck(ctx, url, versionExtractKey)
}

func (a *CheckerFactoryAdapter) CheckMultiple(ctx context.Context, urls []string, versionExtractKey string, checkTestVersion int) interface{} {
	return a.factory.CheckMultiple(ctx, urls, versionExtractKey, checkTestVersion)
}

func (a *CheckerFactoryAdapter) CheckWithConfig(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, string, error) {
	return a.factory.CheckWithConfig(ctx, url, versionExtractKey, checkTestVersion)
}

func (a *CheckerFactoryAdapter) ReloadConfig() error {
	return a.factory.ReloadConfig()
}

func (a *CheckerFactoryAdapter) ClearCache() {
	a.factory.ClearCache()
}

func (a *CheckerFactoryAdapter) SubmitAsyncCheck(url, versionExtractKey string, checkTestVersion int, callback func(result common.AsyncCheckResult)) (string, error) {
	// 转换回调函数
	wrappedCallback := func(result common.AsyncCheckResult) {
		interfaceResult := common.AsyncCheckResult{
			ID:         result.ID,
			URL:        result.URL,
			Version:    result.Version,
			Error:      result.Error,
			Status:     result.Status,
			CreateTime: result.CreateTime,
			UpdateTime: result.UpdateTime,
		}
		callback(interfaceResult)
	}
	return a.factory.SubmitAsyncCheck(url, versionExtractKey, checkTestVersion, wrappedCallback)
}

func (a *CheckerFactoryAdapter) GetAsyncCheckStatus(id string) (string, error) {
	return a.factory.GetAsyncCheckStatus(id)
}

func (a *CheckerFactoryAdapter) GetAsyncCheckResult(id string) (*common.AsyncCheckResult, error) {
	result, err := a.factory.GetAsyncCheckResult(id)
	if err != nil {
		return nil, err
	}
	// 转换为common.AsyncCheckResult
	return &common.AsyncCheckResult{
		ID:         result.ID,
		URL:        result.URL,
		Version:    result.Version,
		Error:      result.Error,
		Status:     result.Status,
		CreateTime: result.CreateTime,
		UpdateTime: result.UpdateTime,
	}, nil
}

func (a *CheckerFactoryAdapter) RemoveAsyncCheck(id string) error {
	return a.factory.RemoveAsyncCheck(id)
}

func (a *CheckerFactoryAdapter) ClearAsyncChecks() {
	a.factory.ClearAsyncChecks()
}

func (a *CheckerFactoryAdapter) GetAsyncChecker() common.AsyncCheckerInterface {
	// 转换为common.AsyncCheckerInterface
	return a.factory.GetAsyncChecker()
}

func (a *CheckerFactoryAdapter) StopAsyncChecker() {
	a.factory.StopAsyncChecker()
}

// UpstreamCheckerAdapter 上游检查器适配器
type UpstreamCheckerAdapter struct {
	checker checkers.UpstreamChecker
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

// ConfigSelectorAdapter 配置选择器适配器
type ConfigSelectorAdapter struct {
	selector *checkers.ConfigCheckerSelector
}



func (a *ConfigSelectorAdapter) SelectCheckerWithVersionKey(url, versionExtractKey string) (checkers.UpstreamChecker, error) {
	checker, err := a.selector.SelectCheckerWithVersionKey(url, versionExtractKey)
	if err != nil {
		return nil, err
	}
	return &UpstreamCheckerAdapter{checker}, nil
}

func (a *ConfigSelectorAdapter) SelectCheckerWithOptions(url, versionExtractKey string, checkTestVersion int) (checkers.UpstreamChecker, error) {
	checker, err := a.selector.SelectCheckerWithOptions(url, versionExtractKey, checkTestVersion)
	if err != nil {
		return nil, err
	}
	return &UpstreamCheckerAdapter{checker}, nil
}

func (a *ConfigSelectorAdapter) GetCheckerSettings(checkerName string) (config.CheckerSettings, bool) {
	return a.selector.GetCheckerSettings(checkerName)
}

func (a *ConfigSelectorAdapter) ReloadConfig() error {
	return a.selector.ReloadConfig()
}

func (a *ConfigSelectorAdapter) CheckWithConfig(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error) {
	return a.selector.CheckWithConfig(ctx, url, versionExtractKey, checkTestVersion)
}

var (
	// globalContainer 全局容器实例
	globalContainer *Container
	// containerOnce 确保容器只初始化一次
	containerOnce sync.Once
)

// GetContainer 获取全局容器实例
func GetContainer() *Container {
	containerOnce.Do(func() {
		globalContainer = &Container{}
	})
	return globalContainer
}

// Initialize 初始化容器
func (c *Container) Initialize() error {
	var err error
	c.once.Do(func() {
		// 初始化日志系统
		c.log = logger.InitLogger()

		// 加载配置
		c.config, err = config.LoadConfig("")
		if err != nil {
			c.log.Fatalf("加载配置失败: %v", err)
			return
		}

		// 初始化数据库连接
		c.db, err = database.InitDatabase()
		if err != nil {
			c.log.Fatalf("数据库初始化失败: %v", err)
			return
		}

		// 执行数据库迁移
		err = database.RunMigrations(database.GetDB())
		if err != nil {
			c.log.Fatalf("数据库迁移失败: %v", err)
			return
		}

		// 初始化检查器工厂
		factory := checkers.NewCheckerFactory()
		c.checkerFactory = &CheckerFactoryAdapter{factory}
		// 确保检查器工厂实现了FactoryProvider接口
		var _ checkers.FactoryProvider = c.checkerFactory

		// 初始化配置选择器
		selector := checkers.NewConfigCheckerSelector()
		c.configSelector = &ConfigSelectorAdapter{selector}
	})
	return err
}

// GetDB 获取数据库连接
func (c *Container) GetDB() *sql.DB {
	return c.db
}

// GetLogger 获取日志记录器
func (c *Container) GetLogger() *logger.Logger {
	return c.log
}

// GetConfig 获取配置
func (c *Container) GetConfig() *config.Config {
	return c.config
}

// GetCheckerFactory 获取检查器工厂
func (c *Container) GetCheckerFactory() checkers.FactoryProvider {
	return c.checkerFactory
}

// GetConfigSelector 获取配置选择器
func (c *Container) GetConfigSelector() checkers.ConfigSelector {
	return c.configSelector
}
