package common

import (
	"aur-update-checker/internal/logger"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// AsyncCheckerImpl AsyncChecker的实现
type AsyncCheckerImpl struct {
	factory          FactoryProvider
	concurrentChecker ConcurrentCheckerInterface
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
func NewAsyncChecker(factory FactoryProvider, workerCount int) AsyncCheckerInterface {
	return NewAsyncCheckerWithSettings(factory, workerCount, 1000)
}

// NewAsyncCheckerWithSettings 创建带设置的异步检查器
func NewAsyncCheckerWithSettings(factory FactoryProvider, workerCount, maxPending int, resultChanSize ...int) AsyncCheckerInterface {
	ctx, cancel := context.WithCancel(context.Background())

	// 计算结果通道大小
	chanSize := 100
	if len(resultChanSize) > 0 && resultChanSize[0] > 0 {
		chanSize = resultChanSize[0]
	}

	checker := &AsyncCheckerImpl{
		factory:       factory,
		requests:      make(map[string]*AsyncCheckRequest),
		resultChan:    make(chan AsyncCheckResult, chanSize),
		workerCount:   workerCount,
		maxPending:    maxPending,
		ctx:           ctx,
		cancel:        cancel,
		stats: AsyncCheckerStats{
			TotalRequests:     0,
			CompletedRequests: 0,
			FailedRequests:   0,
			AverageTime:      0,
		},
	}

	return checker
}

// Start 启动异步检查器
func (c *AsyncCheckerImpl) Start() {
	logger.GlobalLogger.Infof("启动异步检查器，工作线程数: %d", c.workerCount)

	// 启动工作线程
	for i := 0; i < c.workerCount; i++ {
		go c.worker(i)
	}

	// 启动结果处理器
	go c.resultProcessor()

	logger.GlobalLogger.Info("异步检查器启动完成")
}

// Stop 停止异步检查器
func (c *AsyncCheckerImpl) Stop() {
	logger.GlobalLogger.Info("停止异步检查器")

	// 取消上下文
	c.cancel()

	// 等待所有请求完成
	c.mutex.Lock()
	for len(c.requests) > 0 {
		c.mutex.Unlock()
		time.Sleep(100 * time.Millisecond)
		c.mutex.Lock()
	}
	c.mutex.Unlock()

	logger.GlobalLogger.Info("异步检查器已停止")
}

// Check 异步检查版本
func (c *AsyncCheckerImpl) Check(request *AsyncCheckRequest) (string, error) {
	// 默认使用HTTP检查器
	return c.CheckWithCheckerName(request, "http")
}

// CheckWithCheckerName 使用指定检查器名称进行异步检查
func (c *AsyncCheckerImpl) CheckWithCheckerName(request *AsyncCheckRequest, checkerName string) (string, error) {
	// 生成请求ID
	requestID := fmt.Sprintf("%s-%d", checkerName, time.Now().UnixNano())

	// 添加请求到队列
	c.mutex.Lock()
	if len(c.requests) >= c.maxPending {
		c.mutex.Unlock()
		return "", fmt.Errorf("too many pending requests")
	}

	request.ID = requestID
	request.Status = "pending"
	request.CreatedAt = time.Now()
	c.requests[requestID] = request

	// 更新统计信息
	c.statsMutex.Lock()
	c.stats.TotalRequests++
	c.statsMutex.Unlock()

	c.mutex.Unlock()

	return requestID, nil
}

// GetResult 获取检查结果
func (c *AsyncCheckerImpl) GetResult(requestID string) (*AsyncCheckResult, error) {
	c.mutex.RLock()
	request, exists := c.requests[requestID]
	c.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("request not found")
	}

	if request.Status == "pending" {
		return &AsyncCheckResult{
			ID:     requestID,
			Status: "pending",
		}, nil
	}

	return &AsyncCheckResult{
		ID:        requestID,
		Status:    request.Status,
		Version:   request.Result,
		Error:     request.Error,
		CreatedAt: request.CreatedAt,
		UpdateTime: func() time.Time { 
				if request.CompletedAt != nil { 
					return *request.CompletedAt 
				} 
				return time.Time{} 
			}(),
	}, nil
}

// GetStats 获取统计信息
func (c *AsyncCheckerImpl) GetStats() AsyncCheckerStats {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()
	return c.stats
}

// Submit 提交异步检查请求
func (c *AsyncCheckerImpl) Submit(url, versionExtractKey string, checkTestVersion int, callback func(result AsyncCheckResult)) (string, error) {
	// 创建请求
	request := &AsyncCheckRequest{
		URL:               url,
		VersionExtractKey: versionExtractKey,
		CheckTestVersion:  checkTestVersion,
		Callback:          callback,
	}
	
	// 调用Check方法
	// 注意：这里需要修改Check方法，不再依赖CheckerName字段
	// 而是通过URL或其他方式确定使用哪个检查器
	return c.CheckWithCheckerName(request, "http") // 默认使用HTTP检查器
}

// Clear 清除所有异步检查请求
func (c *AsyncCheckerImpl) Clear() {
	logger.GlobalLogger.Info("清除所有异步检查请求")
	
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// 取消所有待处理的请求
	for id, request := range c.requests {
		if request.Status == "pending" {
			request.Status = "cancelled"
			now := time.Now(); request.CompletedAt = &now
			
			// 更新统计信息
			c.statsMutex.Lock()
			c.stats.FailedRequests++
			c.statsMutex.Unlock()
			
			// 如果有回调函数，调用它
			if request.Callback != nil {
				request.Callback(AsyncCheckResult{
					ID:        id,
					Status:    "cancelled",
					Error:     fmt.Errorf("request cancelled"),
					CreatedAt: request.CreatedAt,
				})
			}
		}
	}
	
	// 清空请求映射
	c.requests = make(map[string]*AsyncCheckRequest)
	
	logger.GlobalLogger.Info("所有异步检查请求已清除")
}

// SetMaxPending 设置最大待处理请求数
func (c *AsyncCheckerImpl) SetMaxPending(max int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.maxPending = max
	logger.GlobalLogger.Infof("设置最大待处理请求数为: %d", max)
}

// SetWorkerCount 设置工作线程数
func (c *AsyncCheckerImpl) SetWorkerCount(count int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// 如果检查器正在运行，需要先停止
	wasRunning := c.ctx != nil
	if wasRunning {
		c.Stop()
	}
	
	// 更新工作线程数
	c.workerCount = count
	
	// 如果之前在运行，重新启动
	if wasRunning {
		c.Start()
	}
	
	logger.GlobalLogger.Infof("设置工作线程数为: %d", count)
}

// GetStatus 获取检查状态
func (c *AsyncCheckerImpl) GetStatus(id string) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	request, exists := c.requests[id]
	if !exists {
		return "", fmt.Errorf("request not found")
	}
	
	return request.Status, nil
}

// Remove 移除检查请求
func (c *AsyncCheckerImpl) Remove(id string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	request, exists := c.requests[id]
	if !exists {
		return fmt.Errorf("request not found")
	}
	
	// 如果请求正在处理中，取消它
	if request.Status == "pending" {
		request.Status = "cancelled"
		now := time.Now(); request.CompletedAt = &now
		
		// 更新统计信息
		c.statsMutex.Lock()
		c.stats.FailedRequests++
		c.statsMutex.Unlock()
		
		// 如果有回调函数，调用它
		if request.Callback != nil {
			request.Callback(AsyncCheckResult{
				ID:        id,
				Status:    "cancelled",
				Error:     fmt.Errorf("request cancelled"),
				CreatedAt: request.CreatedAt,
			})
		}
	}
	
	// 从映射中移除请求
	delete(c.requests, id)
	
	return nil
}

// worker 工作线程
func (c *AsyncCheckerImpl) worker(id int) {
	logger.GlobalLogger.Debugf("异步检查器工作线程 %d 启动", id)

	for {
		select {
		case <-c.ctx.Done():
			logger.GlobalLogger.Debugf("异步检查器工作线程 %d 停止", id)
			return
		default:
			// 获取下一个请求
			c.mutex.Lock()
			var request *AsyncCheckRequest
			for _, req := range c.requests {
				if req.Status == "pending" {
					request = req
					request.Status = "processing"
					break
				}
			}
			c.mutex.Unlock()

			if request == nil {
				// 没有待处理的请求，稍等片刻
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 处理请求
			logger.GlobalLogger.Debugf("工作线程 %d 处理请求: %s", id, request.ID)

			// 从请求ID中提取检查器名称
			// 请求ID格式为 "checkerName-timestamp"
			parts := strings.Split(request.ID, "-")
			if len(parts) < 2 {
				c.resultChan <- AsyncCheckResult{
					ID:        request.ID,
					Status:    "failed",
					Error:     fmt.Errorf("invalid request ID format"),
					CreatedAt: request.CreatedAt,
					UpdateTime: time.Now(),
				}
				continue
			}
			checkerName := parts[0]
			
			// 使用工厂获取检查器
			checker, err := c.factory.GetChecker(checkerName)
			if err != nil {
				c.resultChan <- AsyncCheckResult{
					ID:        request.ID,
					Status:    "failed",
					Error:     err,
					CreatedAt: request.CreatedAt,
					UpdateTime: time.Now(),
				}
				continue
			}

			// 执行检查
			version, err := checker.Check(c.ctx, request.URL, request.VersionExtractKey)

			// 发送结果
			status := "completed"
			
			if err != nil {
				status = "failed"
				
			}

			c.resultChan <- AsyncCheckResult{
				ID:        request.ID,
				Status:    status,
				Version:   version,
				Error:     func() error { 
				if err != nil { 
					return err 
				} 
				return nil 
			}(),
				CreatedAt: request.CreatedAt,
				UpdateTime: time.Now(),
			}
		}
	}
}

// resultProcessor 结果处理器
func (c *AsyncCheckerImpl) resultProcessor() {
	logger.GlobalLogger.Debug("异步检查器结果处理器启动")

	for {
		select {
		case <-c.ctx.Done():
			logger.GlobalLogger.Debug("异步检查器结果处理器停止")
			return
		case result := <-c.resultChan:
			// 更新请求状态
			c.mutex.Lock()
			if request, exists := c.requests[result.ID]; exists {
				request.Status = result.Status
				request.Result = result.Version
				request.Error = result.Error
				request.CompletedAt = &result.UpdateTime

				// 更新统计信息
				c.statsMutex.Lock()
				if result.Status == "completed" {
					c.stats.CompletedRequests++
				} else {
					c.stats.FailedRequests++
				}
				c.statsMutex.Unlock()
			}
			c.mutex.Unlock()
		}
	}
}


