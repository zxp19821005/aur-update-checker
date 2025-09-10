package handlers

import (
	"encoding/json"
	"aur-update-checker/internal/services"
	"aur-update-checker/internal/logger"
)

// AurHandler AUR处理器
type AurHandler struct {
	aurService *services.AurService
	log        *logger.Logger
}

// NewAurHandler 创建AUR处理器实例
func NewAurHandler(aurService *services.AurService, log *logger.Logger) *AurHandler {
	return &AurHandler{
		aurService: aurService,
		log:        log,
	}
}

// CheckAurVersion 检查单个软件包的AUR版本
func (h *AurHandler) CheckAurVersion(packageID int) (map[string]interface{}, error) {
	pkg, err := h.aurService.CheckAurVersion(packageID)
	if err != nil {
		h.log.Errorf("检查AUR版本失败(ID: %d): %v", packageID, err)
		return nil, err
	}

	return h.packageToMap(pkg)
}

// CheckAllAurVersions 检查所有软件包的AUR版本
func (h *AurHandler) CheckAllAurVersions() ([]map[string]interface{}, error) {
	packages, err := h.aurService.CheckAllAurVersions()
	if err != nil {
		h.log.Errorf("检查所有AUR版本失败: %v", err)
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

// packageToMap 将软件包详情转换为map
func (h *AurHandler) packageToMap(pkg interface{}) (map[string]interface{}, error) {
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
