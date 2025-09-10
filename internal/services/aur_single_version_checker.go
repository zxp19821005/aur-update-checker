package services

import (
	"aur-update-checker/internal/database"
	"aur-update-checker/internal/utils"
	"time"

	"gorm.io/gorm"
)

// CheckAurVersion 检查单个软件包的AUR版本
func (s *AurService) CheckAurVersion(packageID int) (database.PackageDetail, error) {
	// 获取软件包信息
	var pkg database.PackageInfo
	if err := s.db.First(&pkg, packageID).Error; err != nil {
		s.log.Errorf("检查AUR版本失败，未找到软件包(ID: %d): %v", packageID, err)
		return database.PackageDetail{}, err
	}

	// 调用AUR API获取软件包信息
	aurPackage, err := s.getAurPackageInfo(pkg.Name)
	if err != nil {
		s.log.Errorf("获取AUR软件包信息失败(%s): %v", pkg.Name, err)

		// 更新AUR信息为失败状态
		s.updateAurInfoFailed(packageID)

		return database.PackageDetail{}, err
	}

	// 使用版本解析器处理AUR版本
	parsedVersion := s.versionParser.ParseAndSaveVersion(aurPackage.Version)
	s.log.Infof("解析AUR软件包版本(%s): 完整版本=%s, 解析后版本=%s", pkg.Name, aurPackage.Version, parsedVersion)

	// 生成上游版本提取参考值
	versionRef := utils.GenerateVersionRef(parsedVersion)

	// 获取或创建AUR信息
	var aurInfo database.AurInfo
	if err := s.db.Where("package_id = ?", packageID).First(&aurInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的AUR信息
			aurInfo = database.AurInfo{
				PackageID:          packageID,
				AurVersion:         parsedVersion, // 使用处理后的版本
				UpstreamVersionRef: versionRef,
				AurCreateDate:      time.Unix(aurPackage.FirstSubmitted, 0),
				AurUpdateDate:      time.Unix(aurPackage.LastModified, 0),
				AurUpdateState:     1, // 成功
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			}

			if err := s.db.Create(&aurInfo).Error; err != nil {
				s.log.Errorf("创建AUR信息失败(ID: %d): %v", packageID, err)
				return database.PackageDetail{}, err
			}
		} else {
			s.log.Errorf("查询AUR信息失败(ID: %d): %v", packageID, err)
			return database.PackageDetail{}, err
		}
	} else {
		// 更新现有的AUR信息
		aurInfo.AurVersion = parsedVersion // 使用处理后的版本
		aurInfo.UpstreamVersionRef = versionRef
		aurInfo.AurCreateDate = time.Unix(aurPackage.FirstSubmitted, 0)
		aurInfo.AurUpdateDate = time.Unix(aurPackage.LastModified, 0)
		aurInfo.AurUpdateState = 1 // 成功
		aurInfo.UpdatedAt = time.Now()

		if err := s.db.Save(&aurInfo).Error; err != nil {
			s.log.Errorf("更新AUR信息失败(ID: %d): %v", packageID, err)
			return database.PackageDetail{}, err
		}
	}

	s.log.Infof("成功检查AUR版本(%s): %s (原始版本: %s)", pkg.Name, parsedVersion, aurPackage.Version)

	// 返回完整的软件包信息
	return s.getPackageDetailWithAur(packageID)
}
