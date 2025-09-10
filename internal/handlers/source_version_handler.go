package handlers

import (
	checkers "aur-update-checker/internal/interfaces/checkers"
	"aur-update-checker/internal/logger"
	"aur-update-checker/internal/services"
	"encoding/json"
)

// UpstreamHandler 上游处理器
type UpstreamHandler struct {
	upstreamService *services.UpstreamService
	log             *logger.Logger
	factory         checkers.FactoryProvider
}

// NewUpstreamHandler 创建上游处理器实例
func NewUpstreamHandler(upstreamService *services.UpstreamService, log *logger.Logger, factory checkers.FactoryProvider) *UpstreamHandler {
	return &UpstreamHandler{
		upstreamService: upstreamService,
		log:             log,
		factory:         factory,
	}
}

// CheckUpstreamVersion 检查单个软件包的上游版本
func (h *UpstreamHandler) CheckUpstreamVersion(packageID int) ([]map[string]interface{}, error) {
	versions, err := h.upstreamService.CheckUpstreamVersion(packageID)
	if err != nil {
		h.log.Errorf("检查上游版本失败(ID: %d): %v", packageID, err)
		return nil, err
	}

	// 转换为map切片
	var result []map[string]interface{}
	for _, version := range versions {
		versionMap, err := h.versionToMap(version)
		if err != nil {
			h.log.Errorf("转换版本信息失败: %v", err)
			continue
		}
		result = append(result, versionMap)
	}

	return result, nil
}

// CheckAllUpstreamVersions 检查所有软件包的上游版本
func (h *UpstreamHandler) CheckAllUpstreamVersions() ([]map[string]interface{}, error) {
	packages, err := h.upstreamService.CheckAllUpstreamVersions()
	if err != nil {
		h.log.Errorf("检查所有上游版本失败: %v", err)
		return nil, err
	}

	// 转换为map切片
	var result []map[string]interface{}
	for _, pkg := range packages {
		pkgMap, err := h.packageToMap(pkg)
		if err != nil {
			h.log.Errorf("转换软件包信息失败: %v", err)
			continue
		}
		result = append(result, pkgMap)
	}

	return result, nil
}

// versionToMap 将版本信息转换为map
func (h *UpstreamHandler) versionToMap(version services.UpstreamVersion) (map[string]interface{}, error) {
	// 使用JSON序列化和反序列化进行转换
	jsonData, err := json.Marshal(version)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// packageToMap 将软件包详情转换为map
func (h *UpstreamHandler) packageToMap(pkg interface{}) (map[string]interface{}, error) {
	// 使用JSON序列化和反序列化进行转换
	jsonData, err := json.Marshal(pkg)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetUpstreamCheckers 获取所有可用的上游检查器
func (h *UpstreamHandler) GetUpstreamCheckers() ([]string, error) {
	checkers := h.factory.GetAllCheckers()

	var result []string
	for name := range checkers {
		result = append(result, name)
	}

	return result, nil
}
