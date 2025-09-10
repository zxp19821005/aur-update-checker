package checkers

import (
	"aur-update-checker/internal/checkers/common"
	"aur-update-checker/internal/logger"
	"context"
	"fmt"
	"sync"
	"time"
)

// 类型别名，引用common包中定义的类型
type AsyncCheckRequest = common.AsyncCheckRequest
type AsyncCheckResult = common.AsyncCheckResult
type AsyncCheckerStats = common.AsyncCheckerStats

// AsyncChecker 异步检查器
type AsyncChecker struct {
	factory          common.FactoryProvider
	concurrentChecker common.ConcurrentCheckerInterface
	requests         map[string]*AsyncCheckRequest
	mutex            sync.RWMutex
	resultChan       chan AsyncCheckResult
	workerCount      int
	maxPending       int
	ctx              context.Context
	cancel           context.CancelFunc
	stats            AsyncCheckerStats
	statsMutex       sync.Mutex
}

// NewAsyncChecker 创建异步检查器
func NewAsyncChecker(factory common.FactoryProvider, workerCount int) *AsyncChecker {
	return NewAsyncCheckerWithSettings(factory, workerCount, 1000, 100)
}

// NewAsyncCheckerWithSettings 创建带自定义设置的异步检查器
func NewAsyncCheckerWithSettings(factory common.FactoryProvider, workerCount, maxPending, resultChanSize int) *AsyncChecker {
	ctx, cancel := context.WithCancel(context.Background())

	return &AsyncChecker{
		factory:          factory,
		concurrentChecker: factory.GetConcurrentChecker(),
		requests:         make(map[string]*AsyncCheckRequest),
		resultChan:       make(chan AsyncCheckResult, resultChanSize),
		workerCount:      workerCount,
		maxPending:       maxPending,
		ctx:              ctx,
		cancel:           cancel,
		stats: AsyncCheckerStats{
			TotalRequests:     0,
			CompletedRequests: 0,
			FailedRequests:   0,
			AverageTime:      0,
		},
	}
}

// Start 启动异步检查器
func (a *AsyncChecker) Start() {
	logger.GlobalLogger.Infof("启动异步检查器，工作线程数: %d", a.workerCount)

	// 启动结果处理协程
	go a.processResults()

	// 启动工作协程
	for i := 0; i < a.workerCount; i++ {
		go a.worker(i)
	}
}

// Stop 停止异步检查器
func (a *AsyncChecker) Stop() {
	logger.GlobalLogger.Info("停止异步检查器")
	a.cancel()
}

// Submit 提交异步检查请求
func (a *AsyncChecker) Submit(url, versionExtractKey string, checkTestVersion int, callback func(result AsyncCheckResult)) (string, error) {
	// 生成唯一ID
	id := generateAsyncCheckID(url, versionExtractKey, checkTestVersion)

	a.mutex.Lock()
	defer a.mutex.Unlock()

	// 检查是否已经存在相同的请求
	if _, exists := a.requests[id]; exists {
		logger.GlobalLogger.Debugf("已存在相同的检查请求: %s", id)
		return id, nil
	}

	// 检查待处理请求数量是否超过限制
	if len(a.requests) >= a.maxPending {
		return "", fmt.Errorf("待处理请求数量已超过限制: %d", a.maxPending)
	}

	// 创建新的检查请求
	request := &AsyncCheckRequest{
		ID:                id,
		URL:               url,
		VersionExtractKey: versionExtractKey,
		CheckTestVersion:  checkTestVersion,
		Status:            "pending",
		CreatedAt:         time.Now(),
		Callback:          callback,
	}

	// 存储请求
	a.requests[id] = request

	// 更新统计信息
	a.updateStats(func(s *AsyncCheckerStats) {
		s.TotalRequests++
	})

	logger.GlobalLogger.Infof("提交异步检查请求: ID=%s, URL=%s", id, url)

	return id, nil
}

// updateStats 更新统计信息
func (a *AsyncChecker) updateStats(updateFunc func(*AsyncCheckerStats)) {
	a.statsMutex.Lock()
	defer a.statsMutex.Unlock()

	updateFunc(&a.stats)
}

// GetStats 获取统计信息
func (a *AsyncChecker) GetStats() AsyncCheckerStats {
	a.statsMutex.Lock()
	defer a.statsMutex.Unlock()

	return a.stats
}

// GetPendingCount 获取待处理请求数量
func (a *AsyncChecker) GetPendingCount() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	count := 0
	for _, req := range a.requests {
		if req.Status == "pending" || req.Status == "processing" {
			count++
		}
	}
	return count
}

// GetStatus 获取检查状态
func (a *AsyncChecker) GetStatus(id string) (string, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	request, exists := a.requests[id]
	if !exists {
		return "", fmt.Errorf("未找到ID为 '%s' 的检查请求", id)
	}

	return request.Status, nil
}

// GetResult 获取检查结果
func (a *AsyncChecker) GetResult(id string) (*AsyncCheckResult, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	request, exists := a.requests[id]
	if !exists {
		return nil, fmt.Errorf("未找到ID为 '%s' 的检查请求", id)
	}

	if request.Status != "completed" && request.Status != "failed" {
		return nil, fmt.Errorf("检查请求 '%s' 尚未完成，当前状态: %s", id, request.Status)
	}

	result := &AsyncCheckResult{
		ID:        request.ID,
		URL:       request.URL,
		Version:   request.Result,
		Error:     request.Error,
		Status:    request.Status,
		CreatedAt: request.CreatedAt,
	}

	if request.CompletedAt != nil {
		result.Duration = request.CompletedAt.Sub(request.CreatedAt)
	}

	return result, nil
}

// Remove 移除检查请求
func (a *AsyncChecker) Remove(id string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, exists := a.requests[id]; !exists {
		return fmt.Errorf("未找到ID为 '%s' 的检查请求", id)
	}

	delete(a.requests, id)
	logger.GlobalLogger.Debugf("已移除检查请求: %s", id)

	return nil
}

// Clear 清除所有检查请求
func (a *AsyncChecker) Clear() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.requests = make(map[string]*AsyncCheckRequest)
	logger.GlobalLogger.Info("已清除所有检查请求")
}

// worker 工作协程
func (a *AsyncChecker) worker(workerID int) {
	logger.GlobalLogger.Debugf("启动异步检查工作协程: %d", workerID)

	// 工作线程本地状态
	localStats := struct {
		processedCount int
		errorCount     int
		totalTime      time.Duration
	}{}

	// 工作线程健康检查
	healthCheckTicker := time.NewTicker(30 * time.Second)
	defer healthCheckTicker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			logger.GlobalLogger.Debugf("停止异步检查工作协程: %d, 处理了 %d 个请求，其中 %d 个失败",
				workerID, localStats.processedCount, localStats.errorCount)
			return

		case <-healthCheckTicker.C:
			// 定期报告工作线程状态
			avgTime := time.Duration(0)
			if localStats.processedCount > 0 {
				avgTime = localStats.totalTime / time.Duration(localStats.processedCount)
			}
			if localStats.processedCount > 0 {
				logger.GlobalLogger.Infof("工作协程 %d 状态: 已处理 %d 个请求，失败 %d 个，平均耗时 %v",
					workerID, localStats.processedCount, localStats.errorCount, avgTime)
			}

		default:
			// 获取待处理的请求
			request := a.getNextPendingRequest()
			if request == nil {
				// 没有待处理的请求，动态休眠时间，避免CPU空转
				pendingCount := a.GetPendingCount()
				if pendingCount == 0 {
					// 完全没有待处理请求，休眠更长时间
					time.Sleep(1 * time.Second)
				} else {
					// 有待处理请求但当前没有获取到，可能是竞争，短时间休眠
					time.Sleep(100 * time.Millisecond)
				}
				continue
			}

			// 处理请求
			startTime := time.Now()
			a.processRequest(request)
			duration := time.Since(startTime)

			// 更新本地统计
			localStats.processedCount++
			localStats.totalTime += duration
			if request.Status == "failed" {
				localStats.errorCount++
			}

			// 根据处理时间和错误率动态调整工作策略
			if localStats.processedCount%10 == 0 {
				// 每处理10个请求检查一次性能
				errorRate := float64(localStats.errorCount) / float64(localStats.processedCount)
				avgTime := localStats.totalTime / time.Duration(localStats.processedCount)

				// 如果错误率过高或处理时间过长，记录警告
				if errorRate > 0.3 {
					logger.GlobalLogger.Warnf("工作协程 %d 错误率过高: %.2f%%", workerID, errorRate*100)
				}

				if avgTime > 30*time.Second {
					logger.GlobalLogger.Warnf("工作协程 %d 平均处理时间过长: %v", workerID, avgTime)
				}
			}
		}
	}
}

// AdjustWorkerCount 动态调整工作线程数量
func (a *AsyncChecker) AdjustWorkerCount(newCount int) {
	if newCount <= 0 {
		logger.GlobalLogger.Warnf("无效的工作线程数量: %d，保持不变", newCount)
		return
	}

	if newCount == a.workerCount {
		return // 数量未变化
	}

	logger.GlobalLogger.Infof("调整工作线程数量: %d -> %d", a.workerCount, newCount)

	if newCount > a.workerCount {
		// 增加工作线程
		for i := a.workerCount; i < newCount; i++ {
			go a.worker(i)
		}
	} else {
		// 减少工作线程
		// 由于工作线程是无限循环的，我们只能通过取消上下文来停止它们
		// 这里我们不做处理，因为工作线程会在上下文取消时自动退出
		// 实际减少工作线程需要重启整个异步检查器
		logger.GlobalLogger.Warnf("减少工作线程数量需要重启异步检查器，当前数量: %d", a.workerCount)
		return
	}

	a.workerCount = newCount
}

// getNextPendingRequest 获取下一个待处理的请求
func (a *AsyncChecker) getNextPendingRequest() *AsyncCheckRequest {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for _, request := range a.requests {
		if request.Status == "pending" {
			request.Status = "processing"
			return request
		}
	}

	return nil
}

// processRequest 处理检查请求
func (a *AsyncChecker) processRequest(request *AsyncCheckRequest) {
	logger.GlobalLogger.Debugf("开始处理检查请求: %s", request.ID)

	startTime := time.Now()

	// 设置最大重试次数
	maxRetries := 3
	var lastErr error

	for retry := 0; retry <= maxRetries; retry++ {
		if retry > 0 {
			logger.GlobalLogger.Infof("重试检查请求: %s, 第 %d 次重试", request.ID, retry)
			// 指数退避
			time.Sleep(time.Duration(retry*retry) * time.Second)
		}

		// 使用并发检查器执行检查
		version, err := a.concurrentChecker.CheckSingle(a.ctx, request.URL, request.VersionExtractKey, request.CheckTestVersion)

		if err == nil {
			// 成功获取结果
			completedAt := time.Now()

			a.mutex.Lock()
			defer a.mutex.Unlock()

			// 更新请求状态
			request.Status = "completed"
			request.Result = version
			request.CompletedAt = &completedAt

			// 更新统计信息
			a.updateStats(func(s *AsyncCheckerStats) {
				s.CompletedRequests++
				// 更新平均处理时间
				if s.CompletedRequests > 0 {
					totalDuration := s.AverageTime * time.Duration(s.CompletedRequests-1) + completedAt.Sub(startTime)
					s.AverageTime = totalDuration / time.Duration(s.CompletedRequests)
				} else {
					s.AverageTime = completedAt.Sub(startTime)
				}
			})

			logger.GlobalLogger.Infof("检查请求处理成功: %s, 版本: %s", request.ID, version)

			// 发送结果
			a.sendResult(request, completedAt.Sub(startTime), nil, version)
			return
		}

		// 记录错误
		lastErr = err
		logger.GlobalLogger.Errorf("检查请求处理失败: %s, 错误: %v", request.ID, err)

		// 如果是严重错误，不进行重试
		if isCriticalError(err) {
			break
		}
	}

	// 所有重试都失败
	completedAt := time.Now()

	a.mutex.Lock()
	defer a.mutex.Unlock()

	// 更新请求状态
	request.Status = "failed"
	request.Error = lastErr
	request.CompletedAt = &completedAt

	// 更新统计信息
	a.updateStats(func(s *AsyncCheckerStats) {
		s.FailedRequests++
	})

	// 发送结果
	a.sendResult(request, completedAt.Sub(startTime), lastErr, "")
}

// sendResult 发送结果到结果通道和回调函数
func (a *AsyncChecker) sendResult(request *AsyncCheckRequest, duration time.Duration, err error, version string) {
	// 发送结果到结果通道
	result := AsyncCheckResult{
		ID:        request.ID,
		URL:       request.URL,
		Version:   version,
		Error:     err,
		Duration:  duration,
		Status:    request.Status,
		CreatedAt: request.CreatedAt,
	}

	select {
	case a.resultChan <- result:
		// 结果已发送
	default:
		// 结果通道已满，记录警告
		logger.GlobalLogger.Warnf("异步检查结果通道已满，丢弃结果: %s", request.ID)
	}

	// 如果有回调函数，调用它
	if request.Callback != nil {
		// 使用单独的goroutine调用回调，避免阻塞处理流程
		go request.Callback(result)
	}
}

// processResults 处理检查结果
func (a *AsyncChecker) processResults() {
	logger.GlobalLogger.Debug("启动异步检查结果处理协程")

	for {
		select {
		case <-a.ctx.Done():
			logger.GlobalLogger.Debug("停止异步检查结果处理协程")
			return
		case result := <-a.resultChan:
			// 这里可以添加结果的后处理逻辑
			logger.GlobalLogger.Debugf("处理异步检查结果: %s, 状态: %s", result.ID, result.Status)
		}
	}
}
