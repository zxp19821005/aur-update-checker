package checkers

import (
	"context"
	"regexp"
	"sort"
	"sync"

	"aur-update-checker/internal/config"
	"aur-update-checker/internal/logger"
)

// ConfigCheckerSelector 配置驱动的检查器选择器
type ConfigCheckerSelector struct {
	registry *CheckerRegistryAdapter
	config   *config.Config
	mutex    sync.RWMutex
	urlRules []config.URLRule
}

// NewConfigCheckerSelector 创建配置驱动的检查器选择器
func NewConfigCheckerSelector() *ConfigCheckerSelector {
	selector := &ConfigCheckerSelector{
		registry: GetPluginRegistry(),
		config:   config.GetConfig(),
	}

	// 初始化URL规则
	selector.updateURLRules()

	return selector
}

// updateURLRules 更新URL规则
func (s *ConfigCheckerSelector) updateURLRules() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 从配置中获取URL规则
	s.urlRules = s.config.Checkers.URLRules

	// 按优先级排序，优先级高的排在前面
	sort.Slice(s.urlRules, func(i, j int) bool {
		return s.urlRules[i].Priority > s.urlRules[j].Priority
	})

	logger.GlobalLogger.Debugf("已加载 %d 条URL规则", len(s.urlRules))
}



// SelectCheckerWithVersionKey 根据URL和版本提取键选择检查器
func (s *ConfigCheckerSelector) SelectCheckerWithVersionKey(url, versionExtractKey string) (UpstreamChecker, error) {
	// 重新加载URL规则，以防配置已更新
	s.updateURLRules()

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 遍历所有URL规则，找到第一个匹配的
	for _, rule := range s.urlRules {
		matched, err := regexp.MatchString(rule.Pattern, url)
		if err != nil {
			logger.GlobalLogger.Errorf("URL规则 '%s' 正则表达式错误: %v", rule.Name, err)
			continue
		}

		if matched {
			// 如果规则中指定了版本提取键，则使用它
			if rule.VersionExtractKey != "" {
				versionExtractKey = rule.VersionExtractKey
			}

			logger.GlobalLogger.Debugf("URL '%s' 匹配规则 '%s', 使用检查器 '%s', 版本提取键 '%s'", 
				url, rule.Name, rule.Checker, versionExtractKey)

			checker, err := s.registry.Create(rule.Checker)
			if err != nil {
				return nil, err
			}

			// 如果检查器支持设置版本提取键，则设置它
			if keySetter, ok := checker.(VersionKeySetter); ok {
				keySetter.SetVersionExtractKey(versionExtractKey)
			}

			return checker, nil
		}
	}

	// 如果没有匹配的规则，使用默认的检查器
	logger.GlobalLogger.Debugf("URL '%s' 未匹配任何规则，使用默认检查器 'github'", url)
	checker, err := s.registry.Create("github")
	if err != nil {
		return nil, err
	}

	// 如果检查器支持设置版本提取键，则设置它
	if keySetter, ok := checker.(VersionKeySetter); ok {
		keySetter.SetVersionExtractKey(versionExtractKey)
	}

	return checker, nil
}

// SelectCheckerWithOptions 根据URL、版本提取键和选项选择检查器
func (s *ConfigCheckerSelector) SelectCheckerWithOptions(url, versionExtractKey string, checkTestVersion int) (UpstreamChecker, error) {
	// 重新加载URL规则，以防配置已更新
	s.updateURLRules()

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 遍历所有URL规则，找到第一个匹配的
	for _, rule := range s.urlRules {
		matched, err := regexp.MatchString(rule.Pattern, url)
		if err != nil {
			logger.GlobalLogger.Errorf("URL规则 '%s' 正则表达式错误: %v", rule.Name, err)
			continue
		}

		if matched {
			// 如果规则中指定了版本提取键，则使用它
			if rule.VersionExtractKey != "" {
				versionExtractKey = rule.VersionExtractKey
			}

			// 如果规则中指定了是否检查测试版本，则使用它
			if rule.CheckTestVersion {
				checkTestVersion = 1
			}

			logger.GlobalLogger.Debugf("URL '%s' 匹配规则 '%s', 使用检查器 '%s', 版本提取键 '%s', 检查测试版本: %d", 
				url, rule.Name, rule.Checker, versionExtractKey, checkTestVersion)

			checker, err := s.registry.Create(rule.Checker)
			if err != nil {
				return nil, err
			}

			// 如果检查器支持设置版本提取键，则设置它
			if keySetter, ok := checker.(VersionKeySetter); ok {
				keySetter.SetVersionExtractKey(versionExtractKey)
			}

			return checker, nil
		}
	}

	// 如果没有匹配的规则，使用默认的检查器
	logger.GlobalLogger.Debugf("URL '%s' 未匹配任何规则，使用默认检查器 'github'", url)
	checker, err := s.registry.Create("github")
	if err != nil {
		return nil, err
	}

	// 如果检查器支持设置版本提取键，则设置它
	if keySetter, ok := checker.(VersionKeySetter); ok {
		keySetter.SetVersionExtractKey(versionExtractKey)
	}

	return checker, nil
}

// GetCheckerSettings 获取检查器设置
func (s *ConfigCheckerSelector) GetCheckerSettings(checkerName string) (config.CheckerSettings, bool) {
	settings, ok := s.config.Checkers.Settings[checkerName]
	return settings, ok
}

// ReloadConfig 重新加载配置
func (s *ConfigCheckerSelector) ReloadConfig() error {
	err := config.ReloadConfig()
	if err != nil {
		return err
	}

	s.config = config.GetConfig()
	s.updateURLRules()

	logger.GlobalLogger.Info("检查器选择器配置已重新加载")
	return nil
}

// VersionKeySetter 版本提取键设置接口
type VersionKeySetter interface {
	SetVersionExtractKey(key string)
}

// ConfigurableChecker 可配置检查器接口
type ConfigurableChecker interface {
	UpstreamChecker
	ApplySettings(settings config.CheckerSettings)
}

// ApplyConfigToChecker 将配置应用到检查器
func ApplyConfigToChecker(checker UpstreamChecker, settings config.CheckerSettings) {
	if configurable, ok := checker.(ConfigurableChecker); ok {
		configurable.ApplySettings(settings)
		logger.GlobalLogger.Debugf("已应用配置到检查器 '%s'", configurable.Name())
	}
}

// CheckWithConfig 使用配置执行检查
func (s *ConfigCheckerSelector) CheckWithConfig(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error) {
	// 选择检查器
	checker, err := s.SelectCheckerWithOptions(url, versionExtractKey, checkTestVersion)
	if err != nil {
		return "", err
	}

	// 获取检查器设置
	settings, ok := s.GetCheckerSettings(checker.Name())
	if ok {
		// 应用设置到检查器
		ApplyConfigToChecker(checker, settings)
	}

	// 执行检查
	// 尝试使用 CheckWithVersionRef 方法，如果检查器实现了该方法
	if checkerWithVersionRef, ok := checker.(interface {
		CheckWithVersionRef(ctx context.Context, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error)
	}); ok {
		return checkerWithVersionRef.CheckWithVersionRef(ctx, url, versionExtractKey, "", checkTestVersion)
	}
	
	// 如果检查器没有实现 CheckWithVersionRef 方法，则使用 CheckWithOption 方法
	return checker.CheckWithOption(ctx, url, versionExtractKey, checkTestVersion)
}
