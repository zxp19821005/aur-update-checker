package services

import (
	"aur-update-checker/internal/database"
	"aur-update-checker/internal/utils"
	"time"

	"gorm.io/gorm"
)

// CheckAllAurVersions 检查所有软件包的AUR版本
func (s *AurService) CheckAllAurVersions() ([]database.PackageDetail, error) {
	// 获取所有软件包信息
	packageService := NewPackageService(nil, s.log)
	packages, err := packageService.GetAllPackages()
	if err != nil {
		s.log.Errorf("获取所有软件包信息失败: %v", err)
		return nil, err
	}

	// 提取所有软件包名称
	var packageNames []string
	for _, pkg := range packages {
		packageNames = append(packageNames, pkg.Name)
	}

	// 批量获取AUR软件包信息
	aurPackages, err := s.getAurPackagesInfo(packageNames)
	if err != nil {
		s.log.Errorf("批量获取AUR软件包信息失败: %v", err)
		return nil, err
	}

	// 使用批量操作处理AUR信息更新
	// 将PackageDetail转换为PackageInfo
	var packageInfos []database.PackageInfo
	for _, pkg := range packages {
		packageInfo := database.PackageInfo{
			ID:                pkg.ID,
			Name:              pkg.Name,
			UpstreamUrl:       pkg.UpstreamUrl,
			UpstreamChecker:   pkg.UpstreamChecker,
			VersionExtractKey: pkg.VersionExtractKey,
			CheckTestVersion:  pkg.CheckTestVersion,
			CreatedAt:         pkg.CreatedAt,
			UpdatedAt:         pkg.UpdatedAt,
		}
		packageInfos = append(packageInfos, packageInfo)
	}
	return s.batchUpdateAurInfo(packageInfos, aurPackages)
}

// batchUpdateAurInfo 批量更新AUR信息，减少数据库访问次数
func (s *AurService) batchUpdateAurInfo(packages []database.PackageInfo, aurPackages []AurPackage) ([]database.PackageDetail, error) {
	var results []database.PackageDetail
	var aurInfosToUpdate []database.AurInfo
	var aurInfosToCreate []database.AurInfo
	var failedPackageIDs []int

	// 创建软件包名称到软件包信息的映射
	pkgMap := make(map[string]database.PackageInfo)
	for _, pkg := range packages {
		pkgMap[pkg.Name] = pkg
	}

	// 创建软件包名称到AUR包信息的映射
	aurPkgMap := make(map[string]AurPackage)
	for _, aurPkg := range aurPackages {
		aurPkgMap[aurPkg.Name] = aurPkg
	}

	// 首先获取所有现有的AUR信息
	var existingAurInfos []database.AurInfo
	var packageIDs []int
	for _, pkg := range packages {
		packageIDs = append(packageIDs, pkg.ID)
	}

	if err := s.db.Where("package_id IN ?", packageIDs).Find(&existingAurInfos).Error; err != nil {
		s.log.Errorf("批量查询现有AUR信息失败: %v", err)
		return nil, err
	}

	// 创建PackageID到AurInfo的映射
	aurInfoMap := make(map[int]database.AurInfo)
	for _, aurInfo := range existingAurInfos {
		aurInfoMap[aurInfo.PackageID] = aurInfo
	}

	// 处理每个软件包的AUR信息
	for _, pkg := range packages {
		aurPackage, exists := aurPkgMap[pkg.Name]
		if !exists {
			s.log.Errorf("未找到软件包的AUR信息: %s", pkg.Name)
			failedPackageIDs = append(failedPackageIDs, pkg.ID)
			continue
		}

		// 使用版本解析器处理AUR版本
		parsedVersion := s.versionParser.ParseAndSaveVersion(aurPackage.Version)
		s.log.Infof("解析AUR软件包版本(%s): 完整版本=%s, 解析后版本=%s", pkg.Name, aurPackage.Version, parsedVersion)

		// 生成上游版本提取参考值
		versionRef := utils.GenerateVersionRef(parsedVersion)

		// 检查是否已存在AUR信息
		if existingAurInfo, exists := aurInfoMap[pkg.ID]; exists {
			// 准备更新现有的AUR信息
			existingAurInfo.AurVersion = parsedVersion
			existingAurInfo.UpstreamVersionRef = versionRef
			existingAurInfo.AurCreateDate = time.Unix(aurPackage.FirstSubmitted, 0)
			existingAurInfo.AurUpdateDate = time.Unix(aurPackage.LastModified, 0)
			existingAurInfo.AurUpdateState = 1 // 成功
			existingAurInfo.UpdatedAt = time.Now()

			aurInfosToUpdate = append(aurInfosToUpdate, existingAurInfo)
		} else {
			// 准备创建新的AUR信息
			newAurInfo := database.AurInfo{
				PackageID:          pkg.ID,
				AurVersion:         parsedVersion,
				UpstreamVersionRef: versionRef,
				AurCreateDate:      time.Unix(aurPackage.FirstSubmitted, 0),
				AurUpdateDate:      time.Unix(aurPackage.LastModified, 0),
				AurUpdateState:     1, // 成功
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			}

			aurInfosToCreate = append(aurInfosToCreate, newAurInfo)
		}

		s.log.Infof("成功检查AUR版本(%s): %s (原始版本: %s)", pkg.Name, parsedVersion, aurPackage.Version)
	}

	// 批量创建AUR信息
	if len(aurInfosToCreate) > 0 {
		if err := s.db.CreateInBatches(aurInfosToCreate, 100).Error; err != nil {
			s.log.Errorf("批量创建AUR信息失败: %v", err)
			// 记录失败的软件包ID
			for _, aurInfo := range aurInfosToCreate {
				failedPackageIDs = append(failedPackageIDs, aurInfo.PackageID)
			}
		} else {
			s.log.Infof("成功批量创建%d个AUR信息", len(aurInfosToCreate))
		}
	}

	// 批量更新AUR信息
	if len(aurInfosToUpdate) > 0 {
		// 使用事务进行批量更新
		err := s.db.Transaction(func(tx *gorm.DB) error {
			for _, aurInfo := range aurInfosToUpdate {
				if err := tx.Model(&database.AurInfo{}).Where("id = ?", aurInfo.ID).Updates(map[string]interface{}{
					"aur_version":         aurInfo.AurVersion,
					"upstream_version_ref": aurInfo.UpstreamVersionRef,
					"aur_create_date":      aurInfo.AurCreateDate,
					"aur_update_date":      aurInfo.AurUpdateDate,
					"aur_update_state":     aurInfo.AurUpdateState,
					"updated_at":           aurInfo.UpdatedAt,
				}).Error; err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			s.log.Errorf("批量更新AUR信息失败: %v", err)
			// 记录失败的软件包ID
			for _, aurInfo := range aurInfosToUpdate {
				failedPackageIDs = append(failedPackageIDs, aurInfo.PackageID)
			}
		} else {
			s.log.Infof("成功批量更新%d个AUR信息", len(aurInfosToUpdate))
		}
	}

	// 批量更新失败状态的AUR信息
	if len(failedPackageIDs) > 0 {
		s.batchUpdateAurInfoFailed(failedPackageIDs)
	}

	// 获取所有成功更新的软件包详情
	successfulPackageIDs := make([]int, 0, len(packages))
	for _, pkg := range packages {
		if !contains(failedPackageIDs, pkg.ID) {
			successfulPackageIDs = append(successfulPackageIDs, pkg.ID)
		}
	}

	// 批量获取软件包详情
	if len(successfulPackageIDs) > 0 {
		details, err := s.batchGetPackageDetails(successfulPackageIDs)
		if err != nil {
			s.log.Errorf("批量获取软件包详情失败: %v", err)
			// 如果批量获取失败，则逐个获取
			for _, id := range successfulPackageIDs {
				detail, err := s.getPackageDetailWithAur(id)
				if err != nil {
					s.log.Errorf("获取软件包详情失败(ID: %d): %v", id, err)
					continue
				}
				results = append(results, detail)
			}
		} else {
			results = details
		}
	}

	s.log.Infof("成功检查所有软件包的AUR版本，共%d个", len(results))
	return results, nil
}
