package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"aur-update-checker/internal/logger"
	"aur-update-checker/internal/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() (*sql.DB, error) {
	// 首先尝试从当前目录下的data文件夹获取数据库文件（用于打包后的程序）
	localDbPath := filepath.Join(".", "data", "aur_update_checker.db")
	if _, err := os.Stat(localDbPath); err == nil {
		// 本地数据库文件存在，使用它
		dbPath := localDbPath
		
		// 初始化日志系统（如果尚未初始化）
		var log *logger.Logger
		if logger.GlobalLogger != nil {
			log = logger.GlobalLogger
		} else {
			log = logger.InitLogger()
		}
		
		log.Infof("使用本地数据库文件: %s", dbPath)
		
		// 继续使用这个路径初始化数据库
		return initDatabaseWithPath(dbPath)
	}
	
	// 如果本地数据库文件不存在，则使用用户配置目录
	// 获取应用数据目录
	appDir, err := utils.EnsureAppDataDir()
	if err != nil {
		return nil, fmt.Errorf("无法创建应用数据目录: %v", err)
	}

	// 数据库文件路径
	dbPath := filepath.Join(appDir, "aur_update_checker.db")
	
	// 初始化日志系统（如果尚未初始化）
	var log *logger.Logger
	if logger.GlobalLogger != nil {
		log = logger.GlobalLogger
	} else {
		log = logger.InitLogger()
	}
	
	log.Infof("数据库路径: %s", dbPath)

	// 检查数据库文件是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Info("数据库文件不存在，将创建新的数据库文件")
	}
	
	// 继续使用这个路径初始化数据库
	return initDatabaseWithPath(dbPath)
}

// initDatabaseWithPath 使用指定路径初始化数据库
func initDatabaseWithPath(dbPath string) (*sql.DB, error) {
	// 配置GORM
	newLogger := gormLogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormLogger.Config{
			SlowThreshold: time.Second,
			LogLevel:      gormLogger.Silent,
			Colorful:      false,
		},
	)

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	DB = db

	// 获取底层sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层SQL数据库失败: %v", err)
	}

	// 设置连接池 - 优化配置以提高性能
	sqlDB.SetMaxIdleConns(20)          // 增加空闲连接数，减少连接建立开销
	sqlDB.SetMaxOpenConns(200)         // 增加最大连接数，支持更高并发
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大存活时间
	sqlDB.SetConnMaxIdleTime(30 * time.Minute) // 空闲连接最大存活时间，避免长时间占用连接

	return sqlDB, nil
}

// RunMigrations 执行数据库迁移
func RunMigrations(db *gorm.DB) error {
	log := logger.InitLogger()
	log.Info("开始执行数据库迁移...")

	// 自动迁移表结构
	err := db.AutoMigrate(
		&PackageInfo{},
		&AurInfo{},
		&UpstreamInfo{},
	)
	if err != nil {
		log.Errorf("数据库迁移失败: %v", err)
		return err
	}

	log.Info("数据库迁移完成")
	return nil
}

// GetDB 获取GORM数据库实例
func GetDB() *gorm.DB {
	return DB
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
