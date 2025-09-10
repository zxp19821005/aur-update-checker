package services

import (
	"aur-update-checker/internal/database"
	"aur-update-checker/internal/logger"
	"database/sql"

	"gorm.io/gorm"
)

// AUR API响应结构
type AurResponse struct {
	Resultcount int        `json:"resultcount"`
	Results     []AurPackage `json:"results"`
}

// AurPackage AUR包信息
type AurPackage struct {
	Name        string `json:"Name"`
	Version     string `json:"Version"`
	Description string `json:"Description"`
	OutOfDate   int    `json:"OutOfDate"`
	FirstSubmitted  int64 `json:"FirstSubmitted"`
	LastModified   int64 `json:"LastModified"`
	URLPath    string `json:"URLPath"`
}

// AurService AUR服务
type AurService struct {
	db            *gorm.DB
	log           *logger.Logger
	versionParser *AurVersionParser
}

// NewAurService 创建AUR服务实例
func NewAurService(db *sql.DB, log *logger.Logger) *AurService {
	gormDB := database.GetDB()
	return &AurService{
		db:            gormDB,
		log:           log,
		versionParser: NewAurVersionParser(),
	}
}
