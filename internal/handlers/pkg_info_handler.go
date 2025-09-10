package handlers

import (
	"encoding/json"
	"aur-update-checker/internal/services"
	"aur-update-checker/internal/logger"
)

// PackageHandler 软件包处理器
type PackageHandler struct {
	packageService *services.PackageService
	log           *logger.Logger
}

// NewPackageHandler 创建软件包处理器实例
func NewPackageHandler(packageService *services.PackageService, log *logger.Logger) *PackageHandler {
	return &PackageHandler{
		packageService: packageService,
		log:           log,
	}
}

// GetAllPackages 获取所有软件包
func (h *PackageHandler) GetAllPackages() ([]map[string]interface{}, error) {
	packages, err := h.packageService.GetAllPackages()
	if err != nil {
		h.log.Errorf("获取所有软件包失败: %v", err)
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

// GetPackageByID 根据ID获取软件包
func (h *PackageHandler) GetPackageByID(id int) (map[string]interface{}, error) {
	pkg, err := h.packageService.GetPackageByID(id)
	if err != nil {
		h.log.Errorf("获取软件包失败(ID: %d): %v", id, err)
		return nil, err
	}

	return h.packageToMap(pkg)
}

// AddPackage 添加软件包
func (h *PackageHandler) AddPackage(name, upstreamUrl, versionExtractKey, upstreamChecker string, checkTestVersion int) (map[string]interface{}, error) {
	h.log.Infof("处理器接收到添加软件包请求: 名称=%s, 上游URL=%s, 版本提取键=%s, 上游检查器=%s", name, upstreamUrl, versionExtractKey, upstreamChecker)
	
	pkg, err := h.packageService.AddPackage(name, upstreamUrl, versionExtractKey, upstreamChecker, checkTestVersion)
	if err != nil {
		h.log.Errorf("处理器添加软件包失败: %v", err)
		h.log.Errorf("错误类型: %T", err)
		return nil, err
	}

	h.log.Infof("处理器成功添加软件包: %+v", pkg)
	result, err := h.packageToMap(pkg)
	if err != nil {
		h.log.Errorf("转换软件包为Map失败: %v", err)
		return nil, err
	}
	
	return result, nil
}

// UpdatePackage 更新软件包
func (h *PackageHandler) UpdatePackage(id int, name, upstreamUrl, versionExtractKey, upstreamChecker string, checkTestVersion int) (map[string]interface{}, error) {
	pkg, err := h.packageService.UpdatePackage(id, name, upstreamUrl, versionExtractKey, upstreamChecker, checkTestVersion)
	if err != nil {
		h.log.Errorf("更新软件包失败(ID: %d): %v", id, err)
		return nil, err
	}

	return h.packageToMap(pkg)
}

// DeletePackage 删除软件包
func (h *PackageHandler) DeletePackage(id int) error {
	return h.packageService.DeletePackage(id)
}

// packageToMap 将软件包详情转换为map
func (h *PackageHandler) packageToMap(pkg interface{}) (map[string]interface{}, error) {
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
