package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// getPackages 获取所有软件包
func (s *APIServer) getPackages(w http.ResponseWriter, r *http.Request) {
	packages, _ := s.packageService.GetAllPackages()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(packages)
}

// getPackage 获取单个软件包
func (s *APIServer) getPackage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	packageData, err := s.packageService.GetPackageByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(packageData)
}

// addPackage 添加软件包
func (s *APIServer) addPackage(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Name              string `json:"name"`
		UpstreamUrl       string `json:"upstreamUrl"`
		VersionExtractKey string `json:"versionExtractKey"`
		UpstreamChecker   string `json:"upstreamChecker"`
		CheckTestVersion  int    `json:"checkTestVersion"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.packageService.AddPackage(data.Name, data.UpstreamUrl, data.VersionExtractKey, data.UpstreamChecker, data.CheckTestVersion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// updatePackage 更新软件包
func (s *APIServer) updatePackage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var data struct {
		Name              string `json:"name"`
		UpstreamUrl       string `json:"upstreamUrl"`
		VersionExtractKey string `json:"versionExtractKey"`
		UpstreamChecker   string `json:"upstreamChecker"`
		CheckTestVersion  int    `json:"checkTestVersion"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.packageService.UpdatePackage(id, data.Name, data.UpstreamUrl, data.VersionExtractKey, data.UpstreamChecker, data.CheckTestVersion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// deletePackage 删除软件包
func (s *APIServer) deletePackage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	err := s.packageService.DeletePackage(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
