package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"aur-update-checker/internal/database"
	"aur-update-checker/internal/logger"
	"aur-update-checker/internal/services"

	"gorm.io/gorm"

	"github.com/gorilla/mux"
)

type APIServer struct {
	packageService  *services.PackageService
	aurService      *services.AurService
	upstreamService *services.UpstreamService
	timerService    *services.TimerService
	logService      *services.LogService
	log             *logger.Logger
}

func NewAPIServer(db interface{}, log *logger.Logger) *APIServer {
	// 将接口转换为具体的数据库连接
	var gormDB *gorm.DB
	switch v := db.(type) {
	case *sql.DB:
		// 如果是 sql.DB，我们需要获取全局的 gorm.DB 实例
		gormDB = database.GetDB()
	case *gorm.DB:
		gormDB = v
	default:
		log.Fatal("数据库类型转换失败")
	}

	// 获取底层的 sql.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("获取底层SQL数据库失败: %v", err)
	}

	// 初始化服务层
	packageService := services.NewPackageService(sqlDB, log)
	aurService := services.NewAurService(sqlDB, log)
	upstreamService := services.NewUpstreamService(sqlDB, log)
	timerService := services.NewTimerService(sqlDB, log, aurService, upstreamService)

	// 创建适配器，将logger.Logger转换为interfaces.LoggerProvider
	logProvider := &LogProviderAdapter{log}
	logService := services.NewLogService(logProvider)

	return &APIServer{
		packageService:  packageService,
		aurService:      aurService,
		upstreamService: upstreamService,
		timerService:    timerService,
		logService:      logService,
		log:             log,
	}
}

// SetupRouter 设置并返回路由器
func (s *APIServer) SetupRouter() http.Handler {
	router := mux.NewRouter()

	// 设置CORS
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 设置CORS头
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 预检请求结果缓存24小时

			// 如果是预检请求，直接返回200
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// 注册路由
	// 软件包相关路由
	router.HandleFunc("/api/packages", s.getPackages).Methods("GET")
	router.HandleFunc("/api/packages/{id:[0-9]+}", s.getPackage).Methods("GET")
	router.HandleFunc("/api/packages", s.addPackage).Methods("POST")
	router.HandleFunc("/api/packages/{id:[0-9]+}", s.updatePackage).Methods("PUT")
	router.HandleFunc("/api/packages/{id:[0-9]+}", s.deletePackage).Methods("DELETE")

	// AUR相关路由
	router.HandleFunc("/api/aur/check/{id:[0-9]+}", s.checkAurVersion).Methods("POST")
	router.HandleFunc("/api/aur/check/all", s.checkAllAurVersions).Methods("POST")

	// 上游相关路由
	router.HandleFunc("/api/upstream/check/{id:[0-9]+}", s.checkUpstreamVersion).Methods("POST")
	router.HandleFunc("/api/upstream/check/all", s.checkAllUpstreamVersions).Methods("POST")
	router.HandleFunc("/api/upstream/checkers", s.getUpstreamCheckers).Methods("GET")

	// 定时任务相关路由
	router.HandleFunc("/api/timer/status", s.getTimerStatus).Methods("GET")
	router.HandleFunc("/api/timer/start", s.startTimer).Methods("POST")
	router.HandleFunc("/api/timer/stop", s.stopTimer).Methods("POST")

	// 日志相关路由
	router.HandleFunc("/api/logs", s.getLogs).Methods("GET")
	router.HandleFunc("/api/logs/latest", s.getLatestLogs).Methods("GET")
	router.HandleFunc("/api/logs/clear", s.clearLogs).Methods("POST")

	// 静态文件服务
	staticDir := "./frontend/dist"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		// 如果生产环境目录不存在，则使用开发环境目录
		staticDir = "./frontend"
	}
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(staticDir))))

	return router
}

func (s *APIServer) Start(port int) {
	// 设置路由
	router := s.SetupRouter()

	// 启动服务器
	addr := fmt.Sprintf(":%d", port)
	s.log.Infof("HTTP API 服务器启动，监听地址: %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
