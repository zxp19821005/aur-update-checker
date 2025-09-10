package services

import (
	checkers "aur-update-checker/internal/interfaces/checkers"
	"aur-update-checker/internal/database"
	"aur-update-checker/internal/logger"
	"aur-update-checker/internal/utils"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// UpstreamVersion 上游版本信息
type UpstreamVersion struct {
	Version string `json:"version"`
	IsPrerelease bool `json:"isPrerelease"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	DownloadURL string `json:"downloadUrl,omitempty"`
}

// UpstreamService 上游服务
type UpstreamService struct {
	db                *gorm.DB
	log               *logger.Logger
	factory           *checkers.CheckerFactory

}

// NewUpstreamService 创建上游服务实例
func NewUpstreamService(db *sql.DB, log *logger.Logger) *UpstreamService {
	gormDB := database.GetDB()
	baseFactory := checkers.NewCheckerFactory()

	return &UpstreamService{
		db:      gormDB,
		log:     log,
		factory: baseFactory,
	}
}

// Stop 停止服务
func (s *UpstreamService) Stop() {
	s.log.Info("上游版本检查服务已停止")
}

// GetUpstreamCheckers 获取上游检查器列表
func (s *UpstreamService) GetUpstreamCheckers() []string {
	return s.factory.GetAllCheckerNames()
}

// CheckUpstreamVersion 检查单个软件包的上游版本
func (s *UpstreamService) CheckUpstreamVersion(packageID int) ([]UpstreamVersion, error) {
	// 获取软件包信息
	var pkg database.PackageInfo
	if err := s.db.Preload("AurInfo").First(&pkg, packageID).Error; err != nil {
		s.log.Errorf("检查上游版本失败，未找到软件包(ID: %d): %v", packageID, err)
		return nil, err
	}

	// 获取上游版本信息
	var versions []UpstreamVersion
	var err error

	if pkg.CheckTestVersion == 1 {
		// 检查测试版本，使用带选项的方法
		versions, err = s.getUpstreamVersionsWithOption(pkg.UpstreamUrl, pkg.VersionExtractKey, pkg.AurInfo.UpstreamVersionRef, pkg.UpstreamChecker, pkg.CheckTestVersion)
	} else {
		// 不检查测试版本，使用简单的方法
		versions, err = s.getUpstreamVersions(pkg.UpstreamUrl, pkg.VersionExtractKey, pkg.AurInfo.UpstreamVersionRef, pkg.UpstreamChecker)
	}
	if err != nil {
		s.log.Errorf("获取上游版本信息失败(%s): %v", pkg.Name, err)

		// 更新上游信息为失败状态
		s.updateUpstreamInfoFailed(packageID)

		return nil, err
	}

	if len(versions) == 0 {
		s.log.Warnf("未找到上游版本信息(%s)", pkg.Name)
		s.updateUpstreamInfoFailed(packageID)
		return nil, fmt.Errorf("未找到上游版本信息")
	}

	// 根据是否检查测试版本获取最新版本
	var latestVersion string
	if pkg.CheckTestVersion == 1 {
		// 检查测试版本，直接使用最新版本（可能是预发布版本）
		latestVersion = versions[0].Version
		s.log.Infof("软件包 %s 配置为检查测试版本，使用最新版本: %s", pkg.Name, latestVersion)
	} else {
		// 不检查测试版本，只获取稳定版本
		for _, v := range versions {
			if !v.IsPrerelease {
				latestVersion = v.Version
				break
			}
		}

		// 如果没有稳定版本，则使用第一个版本（可能是预发布版本）
		if latestVersion == "" && len(versions) > 0 {
			latestVersion = versions[0].Version
			s.log.Warnf("软件包 %s 未找到稳定版本，使用预发布版本: %s", pkg.Name, latestVersion)
		} else if latestVersion == "" {
			s.log.Warnf("软件包 %s 未找到任何版本", pkg.Name)
		} else {
			s.log.Infof("软件包 %s 使用稳定版本: %s", pkg.Name, latestVersion)
		}
	}

	// 获取或创建上游信息
	var upstreamInfo database.UpstreamInfo
	if err := s.db.Where("package_id = ?", packageID).First(&upstreamInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的上游信息
			upstreamInfo = database.UpstreamInfo{
				PackageID:           packageID,
				UpstreamVersion:     latestVersion,
				UpstreamUpdateDate:  utils.ParseReleaseDate(versions[0].ReleaseDate),
				UpstreamUpdateState: 1, // 成功
				CreatedAt:           utils.Now(),
				UpdatedAt:           utils.Now(),
			}

			if err := s.db.Create(&upstreamInfo).Error; err != nil {
				s.log.Errorf("创建上游信息失败(ID: %d): %v", packageID, err)
				return nil, err
			}
		} else {
			s.log.Errorf("查询上游信息失败(ID: %d): %v", packageID, err)
			return nil, err
		}
	} else {
		// 更新现有的上游信息
		upstreamInfo.UpstreamVersion = latestVersion
		upstreamInfo.UpstreamUpdateDate = utils.ParseReleaseDate(versions[0].ReleaseDate)
		upstreamInfo.UpstreamUpdateState = 1 // 成功
		upstreamInfo.UpdatedAt = utils.Now()

		if err := s.db.Save(&upstreamInfo).Error; err != nil {
			s.log.Errorf("更新上游信息失败(ID: %d): %v", packageID, err)
			return nil, err
		}
	}

	s.log.Infof("成功检查上游版本(%s): %s", pkg.Name, latestVersion)

	return versions, nil
}

// CheckAllUpstreamVersions 检查所有软件包的上游版本
func (s *UpstreamService) CheckAllUpstreamVersions() ([]database.PackageDetail, error) {
	// 获取所有软件包信息
	packageService := NewPackageService(nil, s.log)
	packages, err := packageService.GetAllPackages()
	if err != nil {
		s.log.Errorf("获取所有软件包信息失败: %v", err)
		return nil, err
	}

	// 使用逐个检查方式
	var results []database.PackageDetail
	for _, pkg := range packages {
		_, err := s.CheckUpstreamVersion(pkg.ID)
		if err != nil {
			s.log.Errorf("检查软件包上游版本失败(ID: %d): %v", pkg.ID, err)
			continue
		}

		// 获取更新后的软件包详情
		detail, err := packageService.GetPackageByID(pkg.ID)
		if err != nil {
			s.log.Errorf("获取软件包详情失败(ID: %d): %v", pkg.ID, err)
			continue
		}
		results = append(results, detail)
	}

	s.log.Infof("成功检查所有软件包的上游版本，共%d个", len(results))
	return results, nil
}

// getUpstreamVersions 获取上游版本信息
func (s *UpstreamService) getUpstreamVersions(upstreamUrl, versionExtractKey, versionRef string, checkerType string) ([]UpstreamVersion, error) {
	// 使用检查器工厂获取上游版本
	// 使用 CheckWithVersionRef 方法传递版本引用

	version, err := s.factory.CheckWithVersionRef(context.Background(), checkerType, upstreamUrl, versionExtractKey, versionRef, 0)
	if err != nil {
		return nil, fmt.Errorf("使用检查器 '%s' 提取版本失败: %v", checkerType, err)
	}

	// 创建UpstreamVersion对象
	var upstreamVersion UpstreamVersion
	upstreamVersion.Version = version
	upstreamVersion.IsPrerelease = !utils.IsVersionStable(version)

	// 构建下载URL（这里简化处理，实际可能需要根据不同的检查器类型进行特殊处理）
	if strings.Contains(upstreamUrl, "github.com") {
		upstreamVersion.DownloadURL = upstreamUrl + "/releases/download/v" + version + "/release.tar.gz"
	} else if strings.Contains(upstreamUrl, "gitlab.com") {
		upstreamVersion.DownloadURL = upstreamUrl + "/-/archive/v" + version + "/project-v" + version + ".tar.gz"
	} else {
		upstreamVersion.DownloadURL = upstreamUrl + "/downloads/" + version + ".tar.gz"
	}

	return []UpstreamVersion{upstreamVersion}, nil
}

// getUpstreamVersionsWithOption 根据选项获取上游版本信息
func (s *UpstreamService) getUpstreamVersionsWithOption(upstreamUrl, versionExtractKey, versionRef string, checkerType string, checkTestVersion int) ([]UpstreamVersion, error) {
	// 使用检查器工厂获取上游版本
	// 使用 CheckWithVersionRef 方法传递版本引用

	version, err := s.factory.CheckWithVersionRef(context.Background(), checkerType, upstreamUrl, versionExtractKey, versionRef, checkTestVersion)
	if err != nil {
		return nil, fmt.Errorf("使用检查器 '%s' 提取版本失败: %v", checkerType, err)
	}

	// 创建UpstreamVersion对象
	var upstreamVersion UpstreamVersion
	upstreamVersion.Version = version
	upstreamVersion.IsPrerelease = !utils.IsVersionStable(version)

	// 构建下载URL（这里简化处理，实际可能需要根据不同的检查器类型进行特殊处理）
	if strings.Contains(upstreamUrl, "github.com") {
		upstreamVersion.DownloadURL = upstreamUrl + "/releases/download/v" + version + "/release.tar.gz"
	} else if strings.Contains(upstreamUrl, "gitlab.com") {
		upstreamVersion.DownloadURL = upstreamUrl + "/-/archive/v" + version + "/project-v" + version + ".tar.gz"
	} else {
		upstreamVersion.DownloadURL = upstreamUrl + "/downloads/" + version + ".tar.gz"
	}

	return []UpstreamVersion{upstreamVersion}, nil
}

// updateUpstreamInfoFailed 更新上游信息为失败状态
func (s *UpstreamService) updateUpstreamInfoFailed(packageID int) {
	var upstreamInfo database.UpstreamInfo

	// 尝试查找现有的上游信息
	if err := s.db.Where("package_id = ?", packageID).First(&upstreamInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的上游信息，状态为失败
			upstreamInfo = database.UpstreamInfo{
				PackageID:           packageID,
				UpstreamVersion:     "未知", // 设置一个默认值
				UpstreamUpdateState: 2, // 失败
				UpstreamUpdateDate:  utils.Now(), // 更新检查时间
				CreatedAt:           utils.Now(),
				UpdatedAt:           utils.Now(),
			}

			if err := s.db.Create(&upstreamInfo).Error; err != nil {
				s.log.Errorf("创建失败的上游信息失败(ID: %d): %v", packageID, err)
				return
			}
		} else {
			s.log.Errorf("查询上游信息失败(ID: %d): %v", packageID, err)
			return
		}
	} else {
		// 更新现有的上游信息为失败状态
		upstreamInfo.UpstreamUpdateState = 2 // 失败
		upstreamInfo.UpstreamUpdateDate = utils.Now() // 更新检查时间
		upstreamInfo.UpdatedAt = utils.Now()

		if err := s.db.Save(&upstreamInfo).Error; err != nil {
			s.log.Errorf("更新上游信息为失败状态失败(ID: %d): %v", packageID, err)
			return
		}
	}
}
