package checkers

import (
	"aur-update-checker/internal/checkers/common"
	"aur-update-checker/internal/logger"
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// CheckResult 检查结果
// 使用 common 包中定义的 CheckResult 类型
type CheckResult = common.CheckResult

// CheckCache 检查缓存
type CheckCache struct {
	cache    map[string]CacheEntry
	mutex    sync.RWMutex
	ttl      time.Duration
	maxSize  int
	evictionPolicy string // "lru" (最近最少使用) 或 "fifo" (先进先出)
	keyQueue []string    // 用于跟踪缓存键的顺序
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Version    string
	ExpiryTime time.Time
	LastAccess time.Time // 用于LRU策略
}

// NewCheckCache 创建检查缓存
func NewCheckCache(ttl time.Duration, maxSize int, evictionPolicy string) *CheckCache {
	// 验证淘汰策略
	if evictionPolicy != "lru" && evictionPolicy != "fifo" {
		evictionPolicy = "lru" // 默认使用LRU
	}

	return &CheckCache{
		cache:         make(map[string]CacheEntry),
		ttl:           ttl,
		maxSize:       maxSize,
		evictionPolicy: evictionPolicy,
		keyQueue:      make([]string, 0),
	}
}

// Get 从缓存中获取结果
func (c *CheckCache) Get(key string) (string, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, exists := c.cache[key]
	if !exists {
		return "", false
	}

	// 检查是否过期
	if time.Now().After(entry.ExpiryTime) {
		// 过期条目，从缓存中删除
		c.removeKey(key)
		return "", false
	}

	// 更新最后访问时间（用于LRU策略）
	if c.evictionPolicy == "lru" {
		entry.LastAccess = time.Now()
		c.cache[key] = entry
	}

	return entry.Version, true
}

// Set 设置缓存
func (c *CheckCache) Set(key, version string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 如果缓存已满，执行淘汰策略
	if len(c.cache) >= c.maxSize {
		c.evict()
	}

	// 添加新条目
	c.cache[key] = CacheEntry{
		Version:    version,
		ExpiryTime: time.Now().Add(c.ttl),
		LastAccess: time.Now(),
	}

	// 更新键队列
	c.keyQueue = append(c.keyQueue, key)
}

// evict 执行缓存淘汰策略
func (c *CheckCache) evict() {
	if len(c.cache) == 0 {
		return
	}

	if c.evictionPolicy == "lru" {
		// LRU: 淘汰最近最少使用的条目
		var oldestKey string
		var oldestTime time.Time
		first := true

		for key, entry := range c.cache {
			if first || entry.LastAccess.Before(oldestTime) {
				oldestKey = key
				oldestTime = entry.LastAccess
				first = false
			}
		}

		if oldestKey != "" {
			c.removeKey(oldestKey)
		}
	} else {
		// FIFO: 淘汰最早添加的条目
		if len(c.keyQueue) > 0 {
			oldestKey := c.keyQueue[0]
			c.keyQueue = c.keyQueue[1:]
			delete(c.cache, oldestKey)
		}
	}
}

// removeKey 从缓存中删除指定键
func (c *CheckCache) removeKey(key string) {
	delete(c.cache, key)

	// 从键队列中移除
	for i, k := range c.keyQueue {
		if k == key {
			c.keyQueue = append(c.keyQueue[:i], c.keyQueue[i+1:]...)
			break
		}
	}
}

// ClearExpired 清除所有过期的缓存条目
func (c *CheckCache) ClearExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	keysToRemove := make([]string, 0)

	for key, entry := range c.cache {
		if now.After(entry.ExpiryTime) {
			keysToRemove = append(keysToRemove, key)
		}
	}

	for _, key := range keysToRemove {
		c.removeKey(key)
	}

	logger.GlobalLogger.Debugf("清除了 %d 个过期的缓存条目", len(keysToRemove))
}

// GetSize 获取当前缓存大小
func (c *CheckCache) GetSize() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.cache)
}

// ConcurrentChecker 并发检查器
type ConcurrentChecker struct {
	factory interface {
		AutoCheck(ctx context.Context, url, versionExtractKey string) (string, string, error)
		CheckWithVersionRef(ctx context.Context, checkerName, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error)
	}
	cache   *CheckCache
}

// NewConcurrentChecker 创建并发检查器
func NewConcurrentChecker(factory interface {
	AutoCheck(ctx context.Context, url, versionExtractKey string) (string, string, error)
	CheckWithVersionRef(ctx context.Context, checkerName, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error)
}, cacheTTL time.Duration) *ConcurrentChecker {
	// 默认缓存大小为1000，使用LRU淘汰策略
	return NewConcurrentCheckerWithCacheSettings(factory, cacheTTL, 1000, "lru")
}

// NewConcurrentCheckerWithCacheSettings 创建带自定义缓存设置的并发检查器
func NewConcurrentCheckerWithCacheSettings(factory interface {
	AutoCheck(ctx context.Context, url, versionExtractKey string) (string, string, error)
	CheckWithVersionRef(ctx context.Context, checkerName, url, versionExtractKey, versionRef string, checkTestVersion int) (string, error)
}, cacheTTL time.Duration, cacheMaxSize int, cacheEvictionPolicy string) *ConcurrentChecker {
	return &ConcurrentChecker{
		factory: factory,
		cache:   NewCheckCache(cacheTTL, cacheMaxSize, cacheEvictionPolicy),
	}
}

// CheckSingle 检查单个URL
func (c *ConcurrentChecker) CheckSingle(ctx context.Context, url, versionExtractKey string, checkTestVersion int) (string, error) {
	// 生成缓存键
	cacheKey := c.generateCacheKey(url, versionExtractKey, checkTestVersion)

	// 尝试从缓存获取
	if version, found := c.cache.Get(cacheKey); found {
		logger.GlobalLogger.Debugf("从缓存中获取版本: %s -> %s", url, version)
		return version, nil
	}

	// 自动选择检查器并检查
	_, checkerName, err := c.factory.AutoCheck(ctx, url, versionExtractKey)
	if err != nil {
		return "", err
	}

	// 使用选定的检查器检查版本
	version, err := c.factory.CheckWithVersionRef(ctx, checkerName, url, versionExtractKey, "", checkTestVersion)
	if err != nil {
		return "", err
	}

	// 将结果存入缓存
	c.cache.Set(cacheKey, version)

	return version, nil
}

// CheckMultiple 并发检查多个URL
func (c *ConcurrentChecker) CheckMultiple(ctx context.Context, urls []string, versionExtractKey string, checkTestVersion int) []CheckResult {
	return c.CheckMultipleWithConcurrency(ctx, urls, versionExtractKey, checkTestVersion, 10)
}

// CheckMultipleWithConcurrency 使用指定并发数检查多个URL
func (c *ConcurrentChecker) CheckMultipleWithConcurrency(ctx context.Context, urls []string, versionExtractKey string, checkTestVersion int, maxConcurrency int) []CheckResult {
	results := make([]CheckResult, len(urls))
	var wg sync.WaitGroup

	// 使用带缓冲的通道控制并发数
	semaphore := make(chan struct{}, maxConcurrency)

	// 创建错误处理通道
	errChan := make(chan error, len(urls))

	// 创建上下文取消函数，用于在出现严重错误时取消所有操作
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 启动错误监控协程
	go func() {
		for err := range errChan {
			// 如果发生严重错误，取消所有操作
			if isCriticalError(err) {
				logger.GlobalLogger.Errorf("发生严重错误，取消所有检查: %v", err)
				cancel()
				return
			}
		}
	}()

	for i, url := range urls {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			// 如果上下文已取消，设置剩余结果为取消状态
			for j := i; j < len(urls); j++ {
				results[j] = CheckResult{
					URL:   urls[j],
					Error: ctx.Err(),
				}
			}
			wg.Wait()
			return results
		default:
		}

		wg.Add(1)
		go func(idx int, u string) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 添加超时控制
			checkCtx, checkCancel := context.WithTimeout(ctx, 30*time.Second)
			defer checkCancel()

			startTime := time.Now()

			// 检查单个URL，使用带超时的上下文
			version, err := c.CheckSingle(checkCtx, u, versionExtractKey, checkTestVersion)

			// 如果发生错误，发送到错误通道
			if err != nil {
				errChan <- err
			}

			results[idx] = CheckResult{
				URL:      u,
				Version:  version,
				Error:    err,
				Duration: time.Since(startTime),
			}

			if err != nil {
				logger.GlobalLogger.Errorf("检查URL '%s' 失败: %v", u, err)
			} else {
				logger.GlobalLogger.Infof("成功检查URL '%s', 版本: %s, 耗时: %v", u, version, time.Since(startTime))
			}
		}(i, url)
	}

	wg.Wait()
	close(errChan)
	return results
}

// isCriticalError 判断是否为严重错误
// 使用 async_check_utils.go 中定义的 isCriticalError 函数

// generateCacheKey 生成缓存键
func (c *ConcurrentChecker) generateCacheKey(url, versionExtractKey string, checkTestVersion int) string {
	// 使用更可靠的哈希算法生成缓存键，避免特殊字符导致的问题
	h := sha256.New()

	// 写入URL
	h.Write([]byte(url))

	// 写入分隔符
	h.Write([]byte("|"))

	// 写入版本提取键
	h.Write([]byte(versionExtractKey))

	// 写入分隔符
	h.Write([]byte("|"))

	// 写入测试版本标志
	h.Write([]byte(strconv.Itoa(checkTestVersion)))

	// 返回十六进制表示的哈希值
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ClearCache 清除缓存
func (c *ConcurrentChecker) ClearCache() {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	c.cache.cache = make(map[string]CacheEntry)
	logger.GlobalLogger.Info("已清除检查缓存")
}
