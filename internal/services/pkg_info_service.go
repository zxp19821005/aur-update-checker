package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"gorm.io/gorm"
	"aur-update-checker/internal/database"
	"aur-update-checker/internal/logger"
)

// PackageService 软件包服务
type PackageService struct {
	db  *gorm.DB
	log *logger.Logger
}

// NewPackageService 创建软件包服务实例
func NewPackageService(db *sql.DB, log *logger.Logger) *PackageService {
	gormDB := database.GetDB()
	return &PackageService{
		db:  gormDB,
		log: log,
	}
}

// GetAllPackages 获取所有软件包
func (s *PackageService) GetAllPackages() ([]database.PackageDetail, error) {
	var packages []database.PackageInfo
	var result []database.PackageDetail

	// 查询所有软件包并预加载关联信息
	if err := s.db.Preload("AurInfo").Preload("UpstreamInfo").Find(&packages).Error; err != nil {
		s.log.Errorf("获取所有软件包失败: %v", err)
		return nil, err
	}

	// 转换为PackageDetail
	for _, pkg := range packages {
		result = append(result, pkg.ToPackageDetail())
	}

	return result, nil
}

// GetPackageByID 根据ID获取软件包
func (s *PackageService) GetPackageByID(id int) (database.PackageDetail, error) {
	var pkg database.PackageInfo

	// 查询软件包并预加载关联信息
	if err := s.db.Preload("AurInfo").Preload("UpstreamInfo").First(&pkg, id).Error; err != nil {
		s.log.Errorf("获取软件包失败(ID: %d): %v", id, err)
		return database.PackageDetail{}, err
	}

	return pkg.ToPackageDetail(), nil
}

// AddPackage 添加软件包
func (s *PackageService) AddPackage(name, upstreamUrl, versionExtractKey, upstreamChecker string, checkTestVersion int) (database.PackageDetail, error) {
	s.log.Infof("尝试添加软件包: 名称=%s, 上游URL=%s, 版本提取键=%s, 上游检查器=%s, 检查测试版本=%d", name, upstreamUrl, versionExtractKey, upstreamChecker, checkTestVersion)
	
	// 验证输入参数
	if name == "" {
		err := fmt.Errorf("软件包名称不能为空")
		s.log.Errorf("添加软件包失败: %v", err)
		return database.PackageDetail{}, err
	}
	
	// 检查是否已存在同名软件包
	var existingPkg database.PackageInfo
	if err := s.db.Where("name = ?", name).First(&existingPkg).Error; err == nil {
		err := fmt.Errorf("已存在同名软件包: %s", name)
		s.log.Errorf("添加软件包失败: %v", err)
		return database.PackageDetail{}, err
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.log.Errorf("检查软件包是否存在时发生错误: %v", err)
		return database.PackageDetail{}, err
	}
	
	// 创建软件包信息
	now := time.Now().Truncate(time.Second) // 截断到秒级
	pkg := database.PackageInfo{
		Name:             name,
		UpstreamUrl:      upstreamUrl,
		VersionExtractKey: versionExtractKey,
		UpstreamChecker:  upstreamChecker,
		CheckTestVersion: checkTestVersion,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	s.log.Infof("准备将软件包信息插入数据库: %+v", pkg)
	
	// 插入数据库
	if err := s.db.Create(&pkg).Error; err != nil {
		s.log.Errorf("添加软件包失败: %v", err)
		s.log.Errorf("数据库错误详情: %+v", err)
		return database.PackageDetail{}, err
	}

	s.log.Infof("成功添加软件包: %s, ID: %d", name, pkg.ID)
	
	// 转换为PackageDetail前再次查询数据库，确保数据完整性
	var refreshedPkg database.PackageInfo
	if err := s.db.Preload("AurInfo").Preload("UpstreamInfo").First(&refreshedPkg, pkg.ID).Error; err != nil {
		s.log.Errorf("刷新软件包信息失败: %v", err)
		// 即使刷新失败，也尝试使用原始数据转换
		return pkg.ToPackageDetail(), nil
	}

	s.log.Infof("准备转换软件包为详情格式: %+v", refreshedPkg)
	detail := refreshedPkg.ToPackageDetail()
	s.log.Infof("成功转换软件包为详情格式: %+v", detail)
	return detail, nil
}

// UpdatePackage 更新软件包
func (s *PackageService) UpdatePackage(id int, name, upstreamUrl, versionExtractKey, upstreamChecker string, checkTestVersion int) (database.PackageDetail, error) {
	// 查询软件包
	var pkg database.PackageInfo
	if err := s.db.First(&pkg, id).Error; err != nil {
		s.log.Errorf("更新软件包失败，未找到软件包(ID: %d): %v", id, err)
		return database.PackageDetail{}, err
	}

	// 更新软件包信息
	pkg.Name = name
	pkg.UpstreamUrl = upstreamUrl
	pkg.VersionExtractKey = versionExtractKey
	pkg.UpstreamChecker = upstreamChecker
	pkg.CheckTestVersion = checkTestVersion
	pkg.UpdatedAt = time.Now()

	// 保存更新
	if err := s.db.Save(&pkg).Error; err != nil {
		s.log.Errorf("更新软件包失败(ID: %d): %v", id, err)
		return database.PackageDetail{}, err
	}

	s.log.Infof("成功更新软件包(ID: %d): %s", id, name)
	return pkg.ToPackageDetail(), nil
}

// DeletePackage 删除软件包
func (s *PackageService) DeletePackage(id int) error {
	// 查询软件包
	var pkg database.PackageInfo
	if err := s.db.First(&pkg, id).Error; err != nil {
		s.log.Errorf("删除软件包失败，未找到软件包(ID: %d): %v", id, err)
		return err
	}

	// 删除软件包（级联删除关联的AUR信息和上游信息）
	if err := s.db.Delete(&pkg).Error; err != nil {
		s.log.Errorf("删除软件包失败(ID: %d): %v", id, err)
		return err
	}

	s.log.Infof("成功删除软件包(ID: %d): %s", id, pkg.Name)
	return nil
}

// GetPackageIDs 获取所有软件包ID列表
func (s *PackageService) GetPackageIDs() ([]int, error) {
	var packages []database.PackageInfo
	var ids []int

	// 查询所有软件包ID
	if err := s.db.Model(&database.PackageInfo{}).Select("id").Find(&packages).Error; err != nil {
		s.log.Errorf("获取软件包ID列表失败: %v", err)
		return nil, err
	}

	// 提取ID
	for _, pkg := range packages {
		ids = append(ids, pkg.ID)
	}

	return ids, nil
}
