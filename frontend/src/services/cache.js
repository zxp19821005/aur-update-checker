
// 缓存服务
class CacheService {
  constructor() {
    this.cache = new Map()
    this.ttl = new Map()
    this.defaultTTL = 5 * 60 * 1000 // 默认缓存5分钟
  }

  // 设置缓存
  set(key, value, ttl = this.defaultTTL) {
    this.cache.set(key, value)
    this.ttl.set(key, Date.now() + ttl)
  }

  // 获取缓存
  get(key) {
    const expiry = this.ttl.get(key)
    if (!expiry) return null

    if (Date.now() > expiry) {
      this.cache.delete(key)
      this.ttl.delete(key)
      return null
    }

    return this.cache.get(key)
  }

  // 删除缓存
  delete(key) {
    this.cache.delete(key)
    this.ttl.delete(key)
  }

  // 清空缓存
  clear() {
    this.cache.clear()
    this.ttl.clear()
  }

  // 获取缓存大小
  size() {
    return this.cache.size
  }

  // 检查是否存在缓存
  has(key) {
    return this.get(key) !== null
  }

  // 获取所有缓存键
  keys() {
    return Array.from(this.cache.keys())
  }

  // 清理过期缓存
  cleanup() {
    const now = Date.now()
    for (const [key, expiry] of this.ttl.entries()) {
      if (now > expiry) {
        this.cache.delete(key)
        this.ttl.delete(key)
      }
    }
  }
}

// 创建缓存实例
const cacheService = new CacheService()

// 定期清理过期缓存
setInterval(() => {
  cacheService.cleanup()
}, 60 * 1000) // 每分钟清理一次

export default cacheService
