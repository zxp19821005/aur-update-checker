package server

import (
	"encoding/json"
	"net/http"
)

// getTimerStatus 获取定时器状态
func (s *APIServer) getTimerStatus(w http.ResponseWriter, r *http.Request) {
	status := s.timerService.GetTimerTaskStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// startTimer 启动定时器
func (s *APIServer) startTimer(w http.ResponseWriter, r *http.Request) {
	var data struct {
		IntervalMinutes int `json:"intervalMinutes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.timerService.StartTimerTask(data.IntervalMinutes)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// stopTimer 停止定时器
func (s *APIServer) stopTimer(w http.ResponseWriter, r *http.Request) {
	s.timerService.StopTimerTask()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
