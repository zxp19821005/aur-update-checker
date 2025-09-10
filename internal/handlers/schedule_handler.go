package handlers

import (
	"aur-update-checker/internal/services"
	"aur-update-checker/internal/logger"
)

// TimerHandler 定时任务处理器
type TimerHandler struct {
	timerService *services.TimerService
	log          *logger.Logger
}

// NewTimerHandler 创建定时任务处理器实例
func NewTimerHandler(timerService *services.TimerService, log *logger.Logger) *TimerHandler {
	return &TimerHandler{
		timerService: timerService,
		log:          log,
	}
}

// StartTimerTask 启动定时任务
func (h *TimerHandler) StartTimerTask(intervalMinutes int) error {
	h.log.Infof("启动定时任务，间隔时间: %d分钟", intervalMinutes)
	return h.timerService.StartTimerTask(intervalMinutes)
}

// StopTimerTask 停止定时任务
func (h *TimerHandler) StopTimerTask() error {
	h.log.Info("停止定时任务")
	return h.timerService.StopTimerTask()
}

// GetTimerTaskStatus 获取定时任务状态
func (h *TimerHandler) GetTimerTaskStatus() (map[string]interface{}, error) {
	status := h.timerService.GetTimerTaskStatus()
	h.log.Debugf("获取定时任务状态: %v", status)
	return status, nil
}
