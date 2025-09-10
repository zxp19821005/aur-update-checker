<template>
    <a-row class="logs-page" style="margin-bottom: 0; height: calc(100vh - 8px); overflow: hidden; padding-bottom: 0; position: relative;">
      <a-col :span="24" style="position: absolute; top: 0; left: 0; right: 0; bottom: 0; overflow: hidden; margin-bottom: 0; padding-bottom: 0;">
        <a-card style="position: absolute; top: 0; left: 0; right: 0; bottom: 0; overflow: hidden; margin-bottom: 0; padding-bottom: 0;">
      <a-space style="width: 100%; justify-content: space-between; margin-bottom: 0;">
        <a-space>
          <a-radio-group v-model:value="logLevel" button-style="solid">
            <a-radio-button value="all">All</a-radio-button>
            <a-radio-button value="debug">Debug</a-radio-button>
            <a-radio-button value="info">Info</a-radio-button>
            <a-radio-button value="warn">Warn</a-radio-button>
            <a-radio-button value="error">Error</a-radio-button>
          </a-radio-group>
          <a-button @click="refreshLogs" :loading="loading">
            <template #icon><reload-outlined /></template>
            刷新
          </a-button>
        </a-space>
        <a-space>
          <a-switch v-model:checked="autoRefresh" checked-children="自动刷新" un-checked-children="手动刷新" />
          <a-select v-model:value="refreshInterval" style="width: 75px">
            <a-select-option :value="5">5s</a-select-option>
            <a-select-option :value="10">10s</a-select-option>
            <a-select-option :value="30">30s</a-select-option>
            <a-select-option :value="60">60s</a-select-option>
          </a-select>
        </a-space>
      </a-space>

      <div
        v-if="isMounted"
        class="log-container"
        :style="{ fontFamily: 'monospace', backgroundColor: '#f5f5f5', padding: '10px', border: '1px solid #d9d9d9', position: 'absolute', top: '60px', left: '8px', right: '8px', bottom: '0' }"
        ref="logContainer"
      >
        <div 
          class="scroll-container" 
          ref="scrollContainer"
        >
          <div 
            v-if="formattedLogs.length > 0" v-for="(log, index) in formattedLogs" 
            :key="index"
            class="log-item"
            :style="{ color: getLogLevelColor(log.level) }"
          >
            {{ log.formatted }}
          </div>
          <div
            v-else
            class="empty-log-container"
          >
            <a-empty description="没有找到符合条件的日志" />
          </div>
        </div>
      </div>

      <!-- 分页控件已隐藏，现在一次性加载所有日志 -->
      <div class="pagination-container" style="display: none; height: 0; margin: 0; padding: 0;">
        <a-pagination
          v-model:current="currentPage"
          v-model:pageSize="pageSize"
          :total="totalLogs"
          :show-size-changer="true"
          :show-total="total => `共 ${total} 条日志`"
          @change="handlePageChange"
          @showSizeChange="handlePageSizeChange"
        />
      </div>

        </a-card>
      </a-col>
    </a-row>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { message } from 'ant-design-vue'
import { ReloadOutlined } from '@ant-design/icons-vue'
import { getLogs } from '@/services/api'

// 日志级别
const logLevel = ref('all')

// 日志容器引用
const logContainer = ref(null)

// 滚动容器引用
const scrollContainer = ref(null)

// 自动刷新
const autoRefresh = ref(true)

// 刷新间隔（秒）
const refreshInterval = ref(10)

// 加载状态
const loading = ref(false)

// 日志内容
const logContent = ref('')

// 格式化后的日志数据
const logsData = ref([])

// 分页相关
const currentPage = ref(1)
const pageSize = ref(100)
const totalLogs = ref(0)

// 最后一条日志的时间戳，用于增量更新
const lastLogTime = ref('')

// 日志格式化缓存
const formatCache = new Map()

// 格式化日志用于显示
const formattedLogs = computed(() => {
  // 清空缓存当页码或日志级别变化时
  if (formatCache.size > 1000) {
    formatCache.clear()
  }

  // 如果没有日志数据，返回空数组
  if (!logsData.value || logsData.value.length === 0) {
    return []
  }

  return logsData.value.map(log => {
    // 使用缓存键
    const cacheKey = `${log.time}-${log.level}-${log.message}`

    // 检查缓存
    if (formatCache.has(cacheKey)) {
      return formatCache.get(cacheKey)
    }

    // 处理后端返回的JSON格式日志
    const time = log.time || new Date().toISOString().replace('T', ' ').substring(0, 19)
    const level = log.level || 'info'
    let message = log.message || ''

    // 如果消息中包含HTML实体，解码它们
    message = message.replace(/&amp;/g, '&')
                     .replace(/&lt;/g, '<')
                     .replace(/&gt;/g, '>')
                     .replace(/&quot;/g, '"')
                     .replace(/&#39;/g, "'")

    const formatted = `[${time}] [${level.toUpperCase()}] ${message}`

    // 缓存结果
    const result = { level, formatted }
    formatCache.set(cacheKey, result)

    return result
  })
})

// 文本区域自动大小设置
const windowHeight = ref(window.innerHeight)

// 组件是否已挂载
const isMounted = ref(false)

// 定时器ID
let refreshTimer = null

// 刷新日志
const refreshLogs = async () => {
  loading.value = true
  try {
    // 调用后端API获取日志，使用较大的页面大小以获取更多日志
    const largePageSize = 1000 // 设置较大的页面大小
    const result = await getLogs(logLevel.value, 1, largePageSize)

    // 检查返回数据格式
    if (!result || !result.logs || !Array.isArray(result.logs)) {
      logContent.value = '获取到的日志数据格式不正确'
      return
    }

    // 更新日志数据和总数
    logsData.value = result.logs
    totalLogs.value = result.total || 0

    // 记录最新日志时间，用于增量更新
    if (result.logs.length > 0) {
      lastLogTime.value = result.logs[0].time
    } else {
      // 如果没有日志，清空lastLogTime和logsData
      lastLogTime.value = ''
      logsData.value = []
    }

    // 为了兼容性，同时更新logContent
    logContent.value = result.logs.map(log => {
      const time = log.time || new Date().toISOString().replace('T', ' ').substring(0, 19)
      const level = log.level || 'info'
      let message = log.message || ''

      // 如果消息中包含HTML实体，解码它们
      message = message.replace(/&amp;/g, '&')
                       .replace(/&lt;/g, '<')
                       .replace(/&gt;/g, '>')
                       .replace(/&quot;/g, '"')
                       .replace(/&#39;/g, "'")

      return `[${time}] [${level.toUpperCase()}] ${message}`
    }).join('\n')
  } catch (error) {
    message.error(`获取日志失败: ${error.message || '未知错误'}`)
    logContent.value = `获取日志失败: ${error.message || '未知错误'}`
  } finally {
    // 检查是否有日志内容
    if (!logContent.value && logsData.value.length === 0) {
      logContent.value = '没有找到符合条件的日志'
    }
    loading.value = false
    // 调整日志容器高度，确保滚动条正确显示
    nextTick(() => {
      adjustLogContainerHeight()
    })
  }
}

// 获取最新日志（用于自动刷新）
const refreshLatestLogs = async () => {
  if (!autoRefresh.value || !lastLogTime.value) return

  try {
    // 直接刷新日志，而不是获取增量日志
    refreshLogs()
  } catch (error) {
    console.error('获取最新日志失败:', error)
  }
}

// 处理页码变化
const handlePageChange = (page) => {
  currentPage.value = page
  refreshLogs()
}

// 处理每页条数变化
const handlePageSizeChange = (current, size) => {
  pageSize.value = size
  currentPage.value = 1 // 重置到第一页
  refreshLogs()
}

// 获取日志级别颜色
const getLogLevelColor = (level) => {
  switch (level) {
    case 'debug': return '#999999'  // 灰色
    case 'info': return '#52c41a'   // 绿色
    case 'warn': return '#faad14'   // 橙色
    case 'error': return '#ff4d4f'   // 红色
    default: return '#000000'       // 黑色
  }
}

// 监听日志级别变化
watch(logLevel, () => {
  currentPage.value = 1 // 重置到第一页
  refreshLogs()
})

// 监听自动刷新状态变化
watch(autoRefresh, (newValue) => {
  // 清除现有定时器
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }

  if (newValue) {
    // 启动定时刷新，使用用户设置的间隔
    refreshTimer = setInterval(refreshLatestLogs, refreshInterval.value * 1000)
  }
})

// 监听刷新间隔变化
watch(refreshInterval, (newValue) => {
  // 如果自动刷新开启，重新设置定时器
  if (autoRefresh.value) {
    // 清除现有定时器
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }

    // 使用新的间隔设置定时器
    refreshTimer = setInterval(refreshLatestLogs, newValue * 1000)
  }
})

// 窗口大小变化标志
const windowSizeChanged = ref(false)

// 调整日志容器高度
const adjustLogContainerHeight = () => {
  if (logContainer.value && scrollContainer.value) {
    // 获取日志容器的实际高度
    const containerHeight = logContainer.value.clientHeight;

    // 设置滚动容器的高度
    scrollContainer.value.style.height = `${containerHeight}px`;

    // 强制显示滚动条
    scrollContainer.value.style.overflow = 'auto';
  }
}

// 处理滚动事件
const handleScroll = () => {
  // 确保滚动条始终可见
  if (scrollContainer.value) {
    // 强制显示滚动条
    scrollContainer.value.style.overflowY = 'scroll';
  }
}

// 组件挂载时加载数据
onMounted(() => {
  refreshLogs()
  // 添加窗口大小变化监听
  window.addEventListener('resize', handleResize)
  // 使用 nextTick 确保 DOM 更新完成后再设置 isMounted
  nextTick(() => {
    isMounted.value = true
    // 调整日志容器高度
    adjustLogContainerHeight()

    // 添加滚动事件监听
    if (scrollContainer.value) {
      scrollContainer.value.addEventListener('scroll', handleScroll)
    }
  })
})

// 组件卸载时清除定时器
onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
  // 移除窗口大小变化监听
  window.removeEventListener('resize', handleResize)
  // 移除滚动事件监听
  if (scrollContainer.value) {
    scrollContainer.value.removeEventListener('scroll', handleScroll)
  }
})

// 处理窗口大小变化
const handleResize = () => {
  // 更新窗口高度
  windowHeight.value = window.innerHeight
  // 调整日志容器高度
  nextTick(() => {
    adjustLogContainerHeight()
  })
}

// 确保textareaAutoSize计算属性在模板中可用
// 在<script setup>中，顶层声明的变量和函数会自动暴露给模板
</script>

<style scoped>
.log-container {
  overflow: hidden !important;
  border: 1px solid #d9d9d9;
  border-radius: 2px;
  margin-bottom: 0;
  padding-bottom: 0;
  box-sizing: border-box;
}

.scroll-container {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  overflow-y: auto;
  overflow-x: hidden;
  box-sizing: border-box;
  margin-bottom: 0;
  padding-bottom: 0;
}

/* 日志项样式 */
.log-item {
  margin-bottom: 0;
  padding: 2px 0;
  word-break: break-all;
}

.empty-log-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  padding: 20px 0;
}

/* 滚动条样式 */
.scroll-container::-webkit-scrollbar {
  width: 12px;
}

.scroll-container::-webkit-scrollbar-thumb {
  background: #666;
  border-radius: 6px;
}

.scroll-container::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 6px;
}

/* 隐藏页面级别的滚动条已在App.vue中设置 */
</style>
