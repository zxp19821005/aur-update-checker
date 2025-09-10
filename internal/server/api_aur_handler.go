package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// checkAurVersion 检查AUR版本
func (s *APIServer) checkAurVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	// 直接调用aurService的方法
	result, err := s.aurService.CheckAurVersion(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// checkAllAurVersions 检查所有AUR版本
func (s *APIServer) checkAllAurVersions(w http.ResponseWriter, r *http.Request) {
	// 直接调用aurService的方法
	results, err := s.aurService.CheckAllAurVersions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}


