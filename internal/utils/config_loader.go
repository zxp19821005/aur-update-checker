package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetAppConfigDir 获取应用程序配置目录
// 优先使用 XDG_CONFIG_HOME 环境变量，如果未设置则使用默认的用户配置目录
func GetAppConfigDir() (string, error) {
	// 检查 XDG_CONFIG_HOME 环境变量
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		appDir := filepath.Join(xdgConfigHome, "aur-update-checker")
		return appDir, nil
	}

	// 如果 XDG_CONFIG_HOME 未设置，使用默认的用户配置目录
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("获取用户配置目录失败: %v", err)
	}

	appDir := filepath.Join(userConfigDir, "aur-update-checker")
	return appDir, nil
}

// EnsureAppConfigDir 确保应用程序配置目录存在
func EnsureAppConfigDir() (string, error) {
	appDir, err := GetAppConfigDir()
	if err != nil {
		return "", err
	}

	// 创建目录（如果不存在）
	if err := os.MkdirAll(appDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "[config] 创建应用配置目录失败: %v\n", err)
		return "", fmt.Errorf("创建应用配置目录失败: %v", err)
	}

	return appDir, nil
}

// GetAppDataDir 获取应用程序数据目录
// 统一使用 XDG_CONFIG_HOME 目录
func GetAppDataDir() (string, error) {
	// 检查 XDG_CONFIG_HOME 环境变量
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		appDir := filepath.Join(xdgConfigHome, "aur-update-checker")
		return appDir, nil
	}

	// 如果 XDG_CONFIG_HOME 未设置，使用默认的用户配置目录
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("获取用户配置目录失败: %v", err)
	}

	appDir := filepath.Join(userConfigDir, "aur-update-checker")
	return appDir, nil
}

// EnsureAppDataDir 确保应用程序数据目录存在
func EnsureAppDataDir() (string, error) {
	appDir, err := GetAppDataDir()
	if err != nil {
		return "", err
	}

	// 创建目录（如果不存在）
	if err := os.MkdirAll(appDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "[config] 创建应用数据目录失败: %v\n", err)
		return "", fmt.Errorf("创建应用数据目录失败: %v", err)
	}

	return appDir, nil
}
