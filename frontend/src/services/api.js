import axios from 'axios'

// 创建axios实例
const api = axios.create({
  baseURL: '/api', // 使用相对路径，让 Vite 的代理处理请求
  timeout: 30000
})

// 软件包相关API
export const getPackageList = () => {
  return api.get('/packages').then(response => response.data);
}

export const getPackageById = (id) => {
  return api.get(`/packages/${id}`).then(response => response.data);
}

export const addPackage = (data) => {
  try {
    // 确保所有必需参数都有值
    if (!data.name) {
      throw new Error('软件包名称不能为空');
    }
    if (!data.upstreamUrl) {
      throw new Error('上游URL不能为空');
    }
    
    // 确保检查器有默认值
    const checker = data.upstreamChecker || '';
    
    return api.post('/packages', {
      name: data.name,
      upstreamUrl: data.upstreamUrl,
      versionExtractKey: data.versionExtractKey || '',
      upstreamChecker: checker,
      checkTestVersion: data.checkTestVersion || 0
    }).then(response => response.data);
  } catch (error) {
    console.error('添加软件包失败:', error);
    throw error;
  }
}

export const updatePackage = (id, data) => {
  return api.put(`/packages/${id}`, {
    name: data.name,
    upstreamUrl: data.upstreamUrl,
    versionExtractKey: data.versionExtractKey,
    upstreamChecker: data.upstreamChecker || '',
    checkTestVersion: data.checkTestVersion || 0
  }).then(response => response.data);
}

export const deletePackage = (id) => {
  return api.delete(`/packages/${id}`).then(response => response.data);
}

// AUR相关API
export const checkAurVersion = (packageId) => {
  return api.post(`/aur/check/${packageId}`).then(response => response.data);
}

export const checkAllAurVersions = () => {
  return api.post('/aur/check/all').then(response => response.data);
}

// 上游相关API
export const checkUpstreamVersion = (packageId) => {
  return api.post(`/upstream/check/${packageId}`).then(response => response.data);
}

export const checkAllUpstreamVersions = () => {
  return api.post('/upstream/check/all').then(response => response.data);
}

export const getUpstreamCheckers = () => {
  return api.get('/upstream/checkers').then(response => response.data);
}

// 日志相关API
export const getLogs = (level = 'all', page = 1, pageSize = 100) => {
  return api.get('/logs', {
    params: {
      level,
      page,
      pageSize
    }
  }).then(response => response.data);
}

export const getLatestLogs = (sinceTime, level = 'all') => {
  return api.get('/logs/latest', {
    params: {
      sinceTime,
      level
    }
  }).then(response => response.data);
}

export const clearLogs = () => {
  return api.post('/logs/clear').then(response => response.data);
}

// 定时任务相关API
export const startTimerTask = (intervalMinutes) => {
  return api.post('/timer/start', { intervalMinutes }).then(response => response.data);
}

export const stopTimerTask = () => {
  return api.post('/timer/stop').then(response => response.data);
}

export const getTimerTaskStatus = () => {
  return api.get('/timer/status').then(response => response.data);
}
