package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"aur-update-checker/internal/errors"
	"aur-update-checker/internal/logger"
	"aur-update-checker/internal/utils"
)

// Config 应用程序配置
type Config struct {
	// 检查器配置
	Checkers CheckerConfig `json:"checkers"`

	// 插件配置
	Plugins PluginConfig `json:"plugins"`

	// 应用全局配置
	Global GlobalConfig `json:"global"`
}

// CheckerConfig 检查器配置
type CheckerConfig struct {
	// 默认检查器
	Default string `json:"default"`

	// 检查器特定配置
	Settings map[string]CheckerSettings `json:"settings"`

	// URL匹配规则
	URLRules []URLRule `json:"urlRules"`
}

// CheckerSettings 检查器特定设置
type CheckerSettings struct {
	// 检查器优先级
	Priority int `json:"priority"`

	// 超时时间（秒）
	Timeout int `json:"timeout"`

	// 重试次数
	RetryCount int `json:"retryCount"`

	// 自定义参数
	CustomParams map[string]interface{} `json:"customParams"`
}

// URLRule URL匹配规则
type URLRule struct {
	// 规则名称
	Name string `json:"name"`

	// URL匹配模式（正则表达式）
	Pattern string `json:"pattern"`

	// 使用的检查器
	Checker string `json:"checker"`

	// 版本提取键
	VersionExtractKey string `json:"versionExtractKey"`

	// 是否检查测试版本
	CheckTestVersion bool `json:"checkTestVersion"`

	// 优先级（数值越高越优先）
	Priority int `json:"priority"`
}

// PluginConfig 插件配置
type PluginConfig struct {
	// 插件目录
	Directory string `json:"directory"`

	// 启用的插件列表
	Enabled []string `json:"enabled"`

	// 插件特定配置
	Settings map[string]PluginSettings `json:"settings"`
}

// PluginSettings 插件特定设置
type PluginSettings struct {
	// 插件参数
	Params map[string]interface{} `json:"params"`

	// 是否启用
	Enabled bool `json:"enabled"`
}

// GlobalConfig 全局配置
type GlobalConfig struct {
	// 日志级别
	LogLevel string `json:"logLevel"`

	// 数据库路径
	DatabasePath string `json:"databasePath"`

	// 检查间隔（分钟）
	CheckInterval int `json:"checkInterval"`

	// 并发检查数
	MaxConcurrentChecks int `json:"maxConcurrentChecks"`

	// 异步检查工作线程数
	AsyncWorkerCount int `json:"asyncWorkerCount"`

	// 缓存TTL（分钟）
	CacheTTL int `json:"cacheTTL"`
}

var (
	// 全局配置实例
	globalConfig *Config
	// 配置加载锁
	configOnce sync.Once
	// 默认配置文件名
	defaultConfigName = "config.json"
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	// 如果未指定配置文件路径，使用默认路径
	if configPath == "" {
		// 获取应用配置目录
		configDir, err := utils.EnsureAppConfigDir()
		if err != nil {
			return nil, errors.NewConfigurationError("获取配置目录失败", err)
		}
		configPath = filepath.Join(configDir, defaultConfigName)
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置
		logger.GlobalLogger.Infof("配置文件 %s 不存在，创建默认配置", configPath)
		config := GetDefaultConfig()
		err := SaveConfig(config, configPath)
		if err != nil {
			return nil, errors.NewConfigurationError("创建默认配置失败", err)
		}
		return config, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.NewSystemError("读取配置文件失败", err)
	}

	// 解析配置文件
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, errors.NewParseError("解析配置文件失败", err)
	}

	logger.GlobalLogger.Infof("成功加载配置文件: %s", configPath)
	return &config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, configPath string) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return errors.NewSystemError("创建配置目录失败", err)
		}
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.NewSystemError("序列化配置失败", err)
	}

	// 写入文件
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// GetConfig 获取全局配置实例
func GetConfig() *Config {
	configOnce.Do(func() {
		var err error
		globalConfig, err = LoadConfig("")
		if err != nil {
			logger.GlobalLogger.Errorf("加载配置失败，使用默认配置: %v", err)
			globalConfig = GetDefaultConfig()
		} else {
			logger.GlobalLogger.Infof("成功加载配置文件")
		}
	})
	return globalConfig
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *Config {
	return &Config{
		Checkers: CheckerConfig{
			Default: "auto",
			Settings: map[string]CheckerSettings{
				"github": {
					Priority:    90,
					Timeout:     30,
					RetryCount:  3,
					CustomParams: map[string]interface{}{
						"api_token": "",
					},
				},
				"gitlab": {
					Priority:    85,
					Timeout:     30,
					RetryCount:  3,
					CustomParams: map[string]interface{}{
						"api_token": "",
					},
				},
				"http": {
					Priority:    50,
					Timeout:     20,
					RetryCount:  2,
				},
				"redirect": {
					Priority:    60,
					Timeout:     15,
					RetryCount:  2,
				},
			},
			URLRules: []URLRule{
				{
					Name:             "GitHub",
					Pattern:          `^https://github\.com/.+`,
					Checker:          "github",
					Priority:         90,
				},
				{
					Name:             "GitLab",
					Pattern:          `^https://gitlab\.com/.+`,
					Checker:          "gitlab",
					Priority:         85,
				},
				{
					Name:             "PyPI",
					Pattern:          `^https://pypi\.org/.+`,
					Checker:          "pypi",
					Priority:         80,
				},
				{
					Name:             "NPM",
					Pattern:          `^https://www\.npmjs\.com/.+`,
					Checker:          "npm",
					Priority:         75,
				},
			},
		},
		Plugins: PluginConfig{
			Directory: "plugins",
			Enabled:   []string{},
			Settings:  map[string]PluginSettings{},
		},
		Global: GlobalConfig{
			LogLevel:            "info",
			DatabasePath:        "aur-checker.db",
			CheckInterval:       60,
			MaxConcurrentChecks: 10,
			AsyncWorkerCount:    5,
			CacheTTL:            5,
		},
	}
}

// ReloadConfig 重新加载配置
func ReloadConfig() error {
	configPath := ""
	if globalConfig != nil {
		// 如果已经有配置实例，尝试从原路径重新加载
		// 这里简化处理，实际应用中可能需要记录配置文件路径
		configDir, err := utils.EnsureAppConfigDir()
		if err == nil {
			configPath = filepath.Join(configDir, defaultConfigName)
		}
	}

	var err error
	globalConfig, err = LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("重新加载配置失败: %v", err)
	}

	logger.GlobalLogger.Info("配置已重新加载")
	return nil
}
