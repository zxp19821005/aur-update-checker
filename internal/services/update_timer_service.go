package services

import (
	"database/sql"
	"sync"
	"time"
	"aur-update-checker/internal/logger"
)

// TimerService 定时任务服务
type TimerService struct {
	db             *sql.DB
	log            *logger.Logger
	aurService     *AurService
	upstreamService *UpstreamService
	timer          *time.Timer
	isRunning      bool
	interval       time.Duration
	stopChan       chan struct{}
	mu             sync.Mutex
}

// NewTimerService 创建定时任务服务实例
func NewTimerService(db *sql.DB, log *logger.Logger, aurService *AurService, upstreamService *UpstreamService) *TimerService {
	return &TimerService{
		db:             db,
		log:            log,
		aurService:     aurService,
		upstreamService: upstreamService,
		isRunning:      false,
		stopChan:       make(chan struct{}),
	}
}

// StartTimerTask 启动定时任务
func (s *TimerService) StartTimerTask(intervalMinutes int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		s.log.Warn("定时任务已在运行中")
		return nil
	}

	// 设置定时任务间隔
	s.interval = time.Duration(intervalMinutes) * time.Minute
	s.isRunning = true

	// 启动定时任务
	s.timer = time.AfterFunc(s.interval, s.runTimerTask)

	s.log.Infof("定时任务已启动，间隔时间: %d分钟", intervalMinutes)
	return nil
}

// StopTimerTask 停止定时任务
func (s *TimerService) StopTimerTask() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		s.log.Warn("定时任务未在运行")
		return nil
	}

	// 停止定时器
	if s.timer != nil {
		s.timer.Stop()
	}

	// 发送停止信号
	close(s.stopChan)

	// 重置状态
	s.isRunning = false
	s.stopChan = make(chan struct{})

	s.log.Info("定时任务已停止")
	return nil
}

// GetTimerTaskStatus 获取定时任务状态
func (s *TimerService) GetTimerTaskStatus() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	status := map[string]interface{}{
		"isRunning": s.isRunning,
	}

	if s.isRunning {
		status["intervalMinutes"] = int(s.interval.Minutes())
	}

	return status
}

// runTimerTask 执行定时任务
func (s *TimerService) runTimerTask() {
	s.mu.Lock()

	if !s.isRunning {
		s.mu.Unlock()
		return
	}

	// 创建新的停止通道
	oldStopChan := s.stopChan
	s.stopChan = make(chan struct{})

	// 设置下一次执行
	s.timer = time.AfterFunc(s.interval, s.runTimerTask)

	s.mu.Unlock()

	s.log.Info("开始执行定时任务...")

	// 执行AUR版本检查
	s.log.Info("检查所有软件包的AUR版本...")
	_, err := s.aurService.CheckAllAurVersions()
	if err != nil {
		s.log.Errorf("检查AUR版本失败: %v", err)
	} else {
		s.log.Info("AUR版本检查完成")
	}

	// 检查是否已停止
	select {
	case <-oldStopChan:
		s.log.Info("定时任务已停止，取消后续操作")
		return
	default:
	}

	// 执行上游版本检查
	s.log.Info("检查所有软件包的上游版本...")
	_, err = s.upstreamService.CheckAllUpstreamVersions()
	if err != nil {
		s.log.Errorf("检查上游版本失败: %v", err)
	} else {
		s.log.Info("上游版本检查完成")
	}

	s.log.Info("定时任务执行完成")
}
