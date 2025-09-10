package services

import (
	"aur-update-checker/internal/database"
	"time"

	"gorm.io/gorm"
)

// updateAurInfoFailed 更新AUR信息为失败状态
func (s *AurService) updateAurInfoFailed(packageID int) {
	var aurInfo database.AurInfo

	// 尝试查找现有的AUR信息
	if err := s.db.Where("package_id = ?", packageID).First(&aurInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的AUR信息，状态为失败
			aurInfo = database.AurInfo{
				PackageID:      packageID,
				AurUpdateState: 2, // 失败
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			if err := s.db.Create(&aurInfo).Error; err != nil {
				s.log.Errorf("创建失败的AUR信息失败(ID: %d): %v", packageID, err)
				return
			}
		} else {
			s.log.Errorf("查询AUR信息失败(ID: %d): %v", packageID, err)
			return
		}
	} else {
		// 更新现有的AUR信息为失败状态
		aurInfo.AurUpdateState = 2 // 失败
		aurInfo.UpdatedAt = time.Now()

		if err := s.db.Save(&aurInfo).Error; err != nil {
			s.log.Errorf("更新AUR信息为失败状态失败(ID: %d): %v", packageID, err)
			return
		}
	}
}

// getPackageDetailWithAur 获取包含AUR信息的软件包详情
func (s *AurService) getPackageDetailWithAur(packageID int) (database.PackageDetail, error) {
	var pkg database.PackageInfo

	// 查询软件包并预加载关联信息
	if err := s.db.Preload("AurInfo").Preload("UpstreamInfo").First(&pkg, packageID).Error; err != nil {
		s.log.Errorf("获取软件包详情失败(ID: %d): %v", packageID, err)
		return database.PackageDetail{}, err
	}

	return pkg.ToPackageDetail(), nil
}

// batchGetPackageDetails 批量获取软件包详情
func (s *AurService) batchGetPackageDetails(packageIDs []int) ([]database.PackageDetail, error) {
	var packages []database.PackageInfo
	var results []database.PackageDetail

	// 批量查询软件包并预加载关联信息
	if err := s.db.Preload("AurInfo").Preload("UpstreamInfo").Where("id IN ?", packageIDs).Find(&packages).Error; err != nil {
		return nil, err
	}

	// 转换为PackageDetail
	for _, pkg := range packages {
		results = append(results, pkg.ToPackageDetail())
	}

	return results, nil
}

// contains 检查切片是否包含某个元素
func contains(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
