package checkers

import (
	"aur-update-checker/internal/logger"
	"aur-update-checker/internal/checkers/common"
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// Cache 缓存接口
type Cache interface {
	// Get 从缓存中获取值
	Get(key string) (string, bool)
	
	// Set 设置缓存值
	Set(key string, value string, ttl time.Duration)
	
	// Delete 删除缓存值
	Delete(key string)
	
	// Clear 清空缓存
	Clear()
}

// EnhancedConcurrentChecker 增强版并发检查器
type EnhancedConcurrentChecker struct {
	factory interface {
		AutoCheck(ctx context.Context, url, versionExtractKey string) (string, string, error)
		CheckWithVersionRef(ctx context.Context, checkerName, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error)
	}
	cache   Cache
	queue   chan common.AsyncCheckRequest
	workers int
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewEnhancedConcurrentChecker 创建增强版并发检查器
func NewEnhancedConcurrentChecker(factory interface {
	AutoCheck(ctx context.Context, url, versionExtractKey string) (string, string, error)
	CheckWithVersionRef(ctx context.Context, checkerName, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error)
}, cache Cache, workers int) *EnhancedConcurrentChecker {
	ctx, cancel := context.WithCancel(context.Background())

	return &EnhancedConcurrentChecker{
		factory: factory,
		cache:   cache,
		queue:   make(chan common.AsyncCheckRequest, 100),
		workers: workers,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start 启动增强版并发检查器
func (e *EnhancedConcurrentChecker) Start() {
	logger.GlobalLogger.Infof("启动增强版并发检查器，工作线程数: %d", e.workers)

	// 启动工作线程
	for i := 0; i < e.workers; i++ {
		e.wg.Add(1)
		go e.worker()
	}
}

// Stop 停止增强版并发检查器
func (e *EnhancedConcurrentChecker) Stop() {
	logger.GlobalLogger.Info("停止增强版并发检查器")
	e.cancel()
	e.wg.Wait()
}

// worker 处理异步检查请求的工作线程
func (e *EnhancedConcurrentChecker) worker() {
	defer e.wg.Done()

	for {
		select {
		case <-e.ctx.Done():
			// 上下文被取消，退出工作线程
			return
		case req := <-e.queue:
			// 处理检查请求
			e.processRequest(req)
		}
	}
}

// processRequest 处理检查请求
func (e *EnhancedConcurrentChecker) processRequest(req common.AsyncCheckRequest) {
	defer func() {
		if r := recover(); r != nil {
			logger.GlobalLogger.Error("处理检查请求时发生panic", "error", r)
		}
	}()

	// 更新请求状态为处理中
	now := time.Now()
	completedAt := now

	// 检查缓存
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(req.URL+req.VersionExtractKey)))
	if cachedVersion, ok := e.cache.Get(cacheKey); ok {
		req.Result = cachedVersion
		req.Status = "completed"
		req.CompletedAt = &completedAt
		if req.Callback != nil {
			req.Callback(common.AsyncCheckResult{
				ID:         req.ID,
				URL:        req.URL,
				Version:    req.Result,
				Error:      nil,
				Status:     "completed",
				CreateTime: req.CreatedAt,
				UpdateTime: time.Now(),
			})
		}
		return
	}

	// 执行检查
	version, _, err := e.factory.AutoCheck(e.ctx, req.URL, req.VersionExtractKey)
	if err != nil {
		req.Error = err
		req.Status = "failed"
	} else {
		req.Result = version
		req.Status = "completed"
		// 缓存结果
		e.cache.Set(cacheKey, version, time.Hour)
	}

	req.CompletedAt = &completedAt

	// 调用回调函数
	if req.Callback != nil {
		var status string
		if req.Error != nil {
			status = "failed"
		} else {
			status = "completed"
		}

		req.Callback(common.AsyncCheckResult{
			ID:         req.ID,
			URL:        req.URL,
			Version:    req.Result,
			Error:      req.Error,
			Status:     status,
			CreateTime: req.CreatedAt,
			UpdateTime: time.Now(),
		})
	}
}

// CheckMultipleWithDynamicConcurrency 动态并发检查多个URL
func (e *EnhancedConcurrentChecker) CheckMultipleWithDynamicConcurrency(
	ctx context.Context,
	requests []common.AsyncCheckRequest,
) []common.AsyncCheckResult {
	results := make([]common.AsyncCheckResult, len(requests))
	var wg sync.WaitGroup

	// 创建一个带缓冲的通道来控制并发数
	concurrencyLimiter := make(chan struct{}, e.workers)

	for i, req := range requests {
		wg.Add(1)
		go func(idx int, r common.AsyncCheckRequest) {
			defer wg.Done()

			// 获取一个并发槽
			concurrencyLimiter <- struct{}{}
			defer func() { <-concurrencyLimiter }()

			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				results[idx] = common.AsyncCheckResult{
					ID:       r.ID,
					URL:      r.URL,
					Error:    ctx.Err(),
					Status:   "failed",
					Duration: 0,
				}
				return
			default:
			}

			// 执行检查
			startTime := time.Now()
			version, _, err := e.factory.AutoCheck(ctx, r.URL, r.VersionExtractKey)
			results[idx] = common.AsyncCheckResult{
				ID:         r.ID,
				URL:        r.URL,
				Version:    version,
				Error:      err,
				Status:     "completed",
				CreateTime: r.CreatedAt,
				UpdateTime: time.Now(),
				Duration:   time.Since(startTime),
			}
		}(i, req)
	}

	// 等待所有检查完成
	wg.Wait()
	return results
}
