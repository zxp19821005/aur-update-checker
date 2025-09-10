package database

import (
	"time"
)

// PackageInfo 基本信息表
type PackageInfo struct {
	ID               int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name             string `gorm:"type:text;not null;unique;index" json:"name"`
	UpstreamUrl      string `gorm:"type:text;not null" json:"upstreamUrl"`
	UpstreamChecker  string `gorm:"type:text;not null;index" json:"upstreamChecker"`
	VersionExtractKey string `gorm:"type:text;not null" json:"versionExtractKey"`
	CheckTestVersion int    `gorm:"default:0;index" json:"checkTestVersion"` // 0:不检查测试版本,1:检查测试版本

	AurInfo          *AurInfo      `gorm:"foreignKey:PackageID" json:"aurInfo"`
	UpstreamInfo     *UpstreamInfo `gorm:"foreignKey:PackageID" json:"upstreamInfo"`

	CreatedAt        time.Time   `json:"createdAt"`
	UpdatedAt        time.Time   `json:"updatedAt"`
}

// AurInfo AUR信息表
type AurInfo struct {
	ID                 int           `gorm:"primaryKey;autoIncrement" json:"id"`
	PackageID          int           `gorm:"not null;index" json:"packageId"`
	AurVersion         string        `gorm:"type:text;not null" json:"aurVersion"`
	UpstreamVersionRef string        `gorm:"type:text;not null" json:"upstreamVersionRef"`
	AurCreateDate      time.Time     `json:"aurCreateDate"`
	AurUpdateDate      time.Time     `json:"aurUpdateDate"`
	AurUpdateState     int           `gorm:"default:0;index" json:"aurUpdateState"` // 0:未检查,1:成功,2:失败

	PackageInfo        *PackageInfo   `gorm:"foreignKey:PackageID" json:"-"`

	CreatedAt          time.Time     `json:"createdAt"`
	UpdatedAt          time.Time     `json:"updatedAt"`
}

// UpstreamInfo 上游信息表
type UpstreamInfo struct {
	ID                  int        `gorm:"primaryKey;autoIncrement" json:"id"`
	PackageID           int        `gorm:"not null;index" json:"packageId"`
	UpstreamVersion     string     `gorm:"type:text;not null" json:"upstreamVersion"`
	UpstreamUpdateDate  time.Time  `json:"upstreamUpdateDate"`
	UpstreamUpdateState int        `gorm:"default:0;index" json:"upstreamUpdateState"` // 0:未检查,1:成功,2:失败

	PackageInfo         *PackageInfo `gorm:"foreignKey:PackageID" json:"-"`

	CreatedAt           time.Time   `json:"createdAt"`
	UpdatedAt           time.Time   `json:"updatedAt"`
}

// PackageDetail 软件包详细信息，用于前后端交互
type PackageDetail struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	UpstreamUrl        string `json:"upstreamUrl"`
	UpstreamChecker    string `json:"upstreamChecker"`
	VersionExtractKey  string `json:"versionExtractKey"`
	CheckTestVersion   int    `json:"checkTestVersion"` // 0:不检查测试版本,1:检查测试版本

	// AUR信息
	AurVersion         string    `json:"aurVersion"`
	UpstreamVersionRef string    `json:"upstreamVersionRef"`
	AurCreateDate      time.Time `json:"aurCreateDate"`
	AurUpdateDate      time.Time `json:"aurUpdateDate"`
	AurUpdateState     int       `json:"aurUpdateState"`

	// 上游信息
	UpstreamVersion    string    `json:"upstreamVersion"`
	UpstreamUpdateDate time.Time `json:"upstreamUpdateDate"`
	UpstreamUpdateState int      `json:"upstreamUpdateState"`

	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// ToPackageDetail 将PackageInfo转换为PackageDetail
func (p *PackageInfo) ToPackageDetail() PackageDetail {
	detail := PackageDetail{
		ID:                p.ID,
		Name:              p.Name,
		UpstreamUrl:       p.UpstreamUrl,
		UpstreamChecker:   p.UpstreamChecker,
		VersionExtractKey: p.VersionExtractKey,
		CheckTestVersion:  p.CheckTestVersion,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}

	if p.AurInfo != nil && p.AurInfo.ID != 0 {
		detail.AurVersion = p.AurInfo.AurVersion
		detail.UpstreamVersionRef = p.AurInfo.UpstreamVersionRef
		detail.AurCreateDate = p.AurInfo.AurCreateDate
		detail.AurUpdateDate = p.AurInfo.AurUpdateDate
		detail.AurUpdateState = p.AurInfo.AurUpdateState
	}

	if p.UpstreamInfo != nil && p.UpstreamInfo.ID != 0 {
		detail.UpstreamVersion = p.UpstreamInfo.UpstreamVersion
		detail.UpstreamUpdateDate = p.UpstreamInfo.UpstreamUpdateDate
		detail.UpstreamUpdateState = p.UpstreamInfo.UpstreamUpdateState
	}

	return detail
}

// UpdateStateText 获取更新状态的文本描述
func UpdateStateText(state int) string {
	switch state {
	case 0:
		return "未检查"
	case 1:
		return "成功"
	case 2:
		return "失败"
	default:
		return "未知"
	}
}
