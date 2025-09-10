package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"aur-update-checker/internal/checkers"
	"aur-update-checker/internal/config"
	"aur-update-checker/internal/database"
	"aur-update-checker/internal/logger"
	"aur-update-checker/internal/server"
	"aur-update-checker/internal/services"
	"aur-update-checker/internal/utils"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "", "配置文件路径")
	port := flag.Int("port", 8080, "HTTP服务器端口")
	flag.Parse()

	// 处理配置文件路径
	var actualConfigPath string
	if *configPath != "" {
		// 使用用户指定的配置文件路径
		actualConfigPath = *configPath
	} else {
		// 使用默认配置文件路径
		configDir, err := utils.EnsureAppConfigDir()
		if err != nil {
			fmt.Printf("获取配置目录失败: %v\n", err)
			os.Exit(1)
		}
		actualConfigPath = filepath.Join(configDir, "config.json")
	}

	// 初始化日志系统
	log := logger.InitLogger()

	// 初始化全局错误处理器
	utils.InitGlobalErrorHandler(log)

	// 加载配置
	_, err := config.LoadConfig(actualConfigPath)
	if err != nil {
		utils.HandleError(err, "加载配置失败")
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	log.Info("AUR更新检查器启动中...")
	log.Infof("使用配置文件: %s", actualConfigPath)

	// 初始化数据库连接
	db, err := database.InitDatabase()
	if err != nil {
		utils.HandleError(err, "数据库初始化失败")
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer db.Close()

	// 记录数据库路径
	log.Infof("数据库连接成功")

	// 执行数据库迁移
	err = database.RunMigrations(database.GetDB())
	if err != nil {
		utils.HandleError(err, "数据库迁移失败")
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 创建API服务器实例
	apiServer := server.NewAPIServer(database.GetDB(), log)

	// 创建HTTP服务器
	apiAddr := fmt.Sprintf(":%d", *port)
	httpServer := &http.Server{
		Addr:    apiAddr,
		Handler: apiServer.SetupRouter(),
	}

	// 初始化检查器注册表
	// 这会触发 checkers 包的 init 函数，确保所有检查器都被注册
	log.Info("初始化检查器注册表...")
	_ = checkers.GetRegistry()

	// 初始化服务层
	aurService := services.NewAurService(db, log)
	upstreamService := services.NewUpstreamService(db, log)
	timerService := services.NewTimerService(db, log, aurService, upstreamService)

	// 启动定时任务，默认间隔为60分钟
	if err := timerService.StartTimerTask(60); err != nil {
		log.Errorf("启动定时任务失败: %v", err)
	} else {
		log.Info("定时任务已启动")
	}

	// 启动HTTP API服务器
	go func() {
		log.Infof("HTTP API 服务器启动，监听地址: %s", apiAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP API 服务器启动失败: %v", err)
		}
	}()

	// 设置信号处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭服务器...")

	// 创建关闭上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 停止定时任务
	if err := timerService.StopTimerTask(); err != nil {
		log.Errorf("停止定时任务失败: %v", err)
	} else {
		log.Info("定时任务已停止")
	}

	// 关闭HTTP API服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Errorf("关闭HTTP API服务器失败: %v", err)
	} else {
		log.Info("HTTP API服务器已关闭")
	}

	log.Info("服务器已关闭")
}