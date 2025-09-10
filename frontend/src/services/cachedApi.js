import { api } from "./api";
import { cache } from "./cache";

/**
 * 带缓存的API服务
 * 使用缓存来减少对后端的请求，提高性能
 */

// 缓存键前缀
const CACHE_KEYS = {
  PACKAGES: "packages",
  PACKAGE_DETAIL: "package_detail",
  PACKAGE_UPSTREAM: "package_upstream",
  SYSTEM_INFO: "system_info",
  CONFIG: "config"
};

// 缓存时间（毫秒）
const CACHE_TTL = {
  PACKAGES: 5 * 60 * 1000, // 5分钟
  PACKAGE_DETAIL: 10 * 60 * 1000, // 10分钟
  PACKAGE_UPSTREAM: 30 * 60 * 1000, // 30分钟
  SYSTEM_INFO: 60 * 60 * 1000, // 1小时
  CONFIG: 24 * 60 * 60 * 1000 // 24小时
};

/**
 * 获取软件包列表（带缓存）
 * @param {Object} params - 查询参数
 * @returns {Promise<Object>} 软件包列表
 */
export async function getPackagesWithCache(params) {
  const cacheKey = `${CACHE_KEYS.PACKAGES}_${JSON.stringify(params)}`;
  const cachedData = cache.get(cacheKey);

  if (cachedData) {
    return cachedData;
  }

  const result = await api.getPackages(params);
  cache.set(cacheKey, result, CACHE_TTL.PACKAGES);
  return result;
}

/**
 * 获取软件包详情（带缓存）
 * @param {number} id - 软件包ID
 * @returns {Promise<Object>} 软件包详情
 */
export async function getPackageDetailWithCache(id) {
  const cacheKey = `${CACHE_KEYS.PACKAGE_DETAIL}_${id}`;
  const cachedData = cache.get(cacheKey);

  if (cachedData) {
    return cachedData;
  }

  const result = await api.getPackageDetail(id);
  cache.set(cacheKey, result, CACHE_TTL.PACKAGE_DETAIL);
  return result;
}

/**
 * 获取软件包上游信息（带缓存）
 * @param {number} id - 软件包ID
 * @returns {Promise<Object>} 软件包上游信息
 */
export async function getPackageUpstreamWithCache(id) {
  const cacheKey = `${CACHE_KEYS.PACKAGE_UPSTREAM}_${id}`;
  const cachedData = cache.get(cacheKey);

  if (cachedData) {
    return cachedData;
  }

  const result = await api.getPackageUpstream(id);
  cache.set(cacheKey, result, CACHE_TTL.PACKAGE_UPSTREAM);
  return result;
}

/**
 * 获取系统信息（带缓存）
 * @returns {Promise<Object>} 系统信息
 */
export async function getSystemInfoWithCache() {
  const cacheKey = CACHE_KEYS.SYSTEM_INFO;
  const cachedData = cache.get(cacheKey);

  if (cachedData) {
    return cachedData;
  }

  const result = await api.getSystemInfo();
  cache.set(cacheKey, result, CACHE_TTL.SYSTEM_INFO);
  return result;
}

/**
 * 获取配置（带缓存）
 * @returns {Promise<Object>} 配置信息
 */
export async function getConfigWithCache() {
  const cacheKey = CACHE_KEYS.CONFIG;
  const cachedData = cache.get(cacheKey);

  if (cachedData) {
    return cachedData;
  }

  const result = await api.getConfig();
  cache.set(cacheKey, result, CACHE_TTL.CONFIG);
  return result;
}

/**
 * 清除所有缓存
 */
export function clearAllCache() {
  cache.clear();
}

/**
 * 清除特定类型的缓存
 * @param {string} type - 缓存类型
 */
export function clearCacheByType(type) {
  if (!CACHE_KEYS[type]) {
    console.warn(`Unknown cache type: ${type}`);
    return;
  }

  // 清除匹配前缀的所有缓存
  cache.removeByPrefix(CACHE_KEYS[type]);
}
