package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// checkUpstreamVersion 检查上游版本
func (s *APIServer) checkUpstreamVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	results, err := s.upstreamService.CheckUpstreamVersion(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// checkAllUpstreamVersions 检查所有上游版本
func (s *APIServer) checkAllUpstreamVersions(w http.ResponseWriter, r *http.Request) {
	results, err := s.upstreamService.CheckAllUpstreamVersions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// getUpstreamCheckers 获取上游检查器列表
func (s *APIServer) getUpstreamCheckers(w http.ResponseWriter, r *http.Request) {
	checkers := s.upstreamService.GetUpstreamCheckers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(checkers)
}
