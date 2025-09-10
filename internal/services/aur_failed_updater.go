package services

import (
	"aur-update-checker/internal/database"
	"time"

	"gorm.io/gorm"
)

// batchUpdateAurInfoFailed 批量更新AUR信息为失败状态
func (s *AurService) batchUpdateAurInfoFailed(packageIDs []int) {
	if len(packageIDs) == 0 {
		return
	}

	// 获取这些软件包的现有AUR信息
	var existingAurInfos []database.AurInfo
	if err := s.db.Where("package_id IN ?", packageIDs).Find(&existingAurInfos).Error; err != nil {
		s.log.Errorf("批量查询AUR信息失败: %v", err)
		// 如果查询失败，则为每个包ID单独创建失败记录
		for _, id := range packageIDs {
			s.updateAurInfoFailed(id)
		}
		return
	}

	// 创建PackageID到AurInfo的映射
	existingAurInfoMap := make(map[int]database.AurInfo)
	for _, aurInfo := range existingAurInfos {
		existingAurInfoMap[aurInfo.PackageID] = aurInfo
	}

	var aurInfosToUpdate []database.AurInfo
	var aurInfosToCreate []database.AurInfo

	now := time.Now()

	// 处理每个软件包ID
	for _, id := range packageIDs {
		if existingAurInfo, exists := existingAurInfoMap[id]; exists {
			// 更新现有的AUR信息为失败状态
			existingAurInfo.AurUpdateState = 2 // 失败
			existingAurInfo.UpdatedAt = now
			aurInfosToUpdate = append(aurInfosToUpdate, existingAurInfo)
		} else {
			// 创建新的失败状态AUR信息
			newAurInfo := database.AurInfo{
				PackageID:      id,
				AurUpdateState: 2, // 失败
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			aurInfosToCreate = append(aurInfosToCreate, newAurInfo)
		}
	}

	// 批量创建失败的AUR信息
	if len(aurInfosToCreate) > 0 {
		if err := s.db.CreateInBatches(aurInfosToCreate, 100).Error; err != nil {
			s.log.Errorf("批量创建失败的AUR信息失败: %v", err)
		}
	}

	// 批量更新失败的AUR信息
	if len(aurInfosToUpdate) > 0 {
		// 使用事务进行批量更新
		err := s.db.Transaction(func(tx *gorm.DB) error {
			for _, aurInfo := range aurInfosToUpdate {
				if err := tx.Model(&database.AurInfo{}).Where("id = ?", aurInfo.ID).Updates(map[string]interface{}{
					"aur_update_state": aurInfo.AurUpdateState,
					"updated_at":       aurInfo.UpdatedAt,
				}).Error; err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			s.log.Errorf("批量更新AUR信息为失败状态失败: %v", err)
		}
	}
}
