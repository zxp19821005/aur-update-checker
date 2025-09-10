<template>
  <a-space style="width: 100%; justify-content: space-between; margin-bottom: 0;">
    <a-space>
      <a-button type="primary" @click="showAddModal">
        <template #icon><plus-outlined /></template>
      </a-button>
      <a-button 
        @click="() => { checkAllAurVersions().catch(e => handleOperationError(e, '检查AUR版本')) }" 
        :loading="checkingAur" 
        title="检查所有AUR版本"
        :disabled="checkingAllVersions"
      >
        <template #icon><sync-outlined /></template>
      </a-button>
      <a-button 
        @click="() => { checkAllUpstreamVersions().catch(e => handleOperationError(e, '检查上游版本')) }" 
        :loading="checkingUpstream" 
        title="检查所有上游版本"
        :disabled="checkingAllVersions"
      >
        <template #icon><CloudDownloadOutlined /></template>
      </a-button>
      <a-button 
        @click="() => { checkAllVersions().catch(e => handleOperationError(e, '检查所有版本')) }" 
        :loading="checkingAllVersions" 
        title="检查所有AUR版本和上游版本"
      >
        <template #icon><reload-outlined /></template>
      </a-button>
      <a-select
        v-model:value="checkerFilter"
        placeholder="按检查器筛选"
        style="width: 100px"
        allow-clear
        @change="onCheckerFilterChange"
        :dropdownMatchSelectWidth="false"
        :getPopupContainer="getPopupContainer"
        popupClassName="checker-filter-dropdown"
        :virtual="false"
        :dropdownStyle="{ minWidth: '120px', maxWidth: '200px' }"
      >
        <a-select-option value="">检查器</a-select-option>
        <a-select-option value="curl">Curl</a-select-option>
        <a-select-option value="gitee">Gitee</a-select-option>
        <a-select-option value="github">GitHub</a-select-option>
        <a-select-option value="gitlab">GitLab</a-select-option>
        <a-select-option value="http">HTTP</a-select-option>
        <a-select-option value="json">JSON</a-select-option>
        <a-select-option value="npm">NPM</a-select-option>
        <a-select-option value="playwright">Playwright</a-select-option>
        <a-select-option value="pypi">PyPI</a-select-option>
        <a-select-option value="redirect">Redirect</a-select-option>
      </a-select>
      <a-select
        v-model:value="statusFilter"
        placeholder="按状态筛选"
        style="width: 100px"
        allow-clear
        @change="onStatusFilterChange"
        :dropdownMatchSelectWidth="false"
        :getPopupContainer="getPopupContainer"
        popupClassName="status-filter-dropdown"
        :virtual="false"
        :dropdownStyle="{ minWidth: '120px', maxWidth: '200px' }"
      >
        <a-select-option value="">状态</a-select-option>
        <a-select-option value="needUpdate">需要更新</a-select-option>
        <a-select-option value="unchecked">未检查</a-select-option>
        <a-select-option value="failed">检查失败</a-select-option>
      </a-select>
    </a-space>
    <a-space>
      <a-input-search
        v-model:value="searchText"
        placeholder="搜索软件包"
        style="width: 200px"
        @search="onSearch"
      />
    </a-space>
  </a-space>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import {
  PlusOutlined,
  SyncOutlined,
  CloudDownloadOutlined,
  ReloadOutlined
} from '@ant-design/icons-vue'
import { usePackageStore } from '@/stores/package'
import { usePackageActions } from './PackageListActions'

const packageStore = usePackageStore()

// 定义emit
const emit = defineEmits(['showAddModal', 'search', 'filterChange'])

// 使用操作逻辑
const {
  checkingAur,
  checkingUpstream,
  checkAllAurVersions,
  checkAllUpstreamVersions,
  checkAllVersions
} = usePackageActions()

// 错误处理
const handleOperationError = (error, operation) => {
  console.error(`执行${operation}时出错:`, error)
  // 使用 Ant Design Vue 的消息提示
  window.$message?.error?.(`${operation}失败: ${error.message || '未知错误'}`)
}

// 显示添加模态框
const showAddModal = () => {
  emit('showAddModal')
}

// 计算是否正在检查所有版本
const checkingAllVersions = computed(() => checkingAur.value || checkingUpstream.value)

// 搜索相关
const searchText = ref('')

// 添加防抖功能，当用户停止输入500ms后才触发搜索
let searchDebounceTimer = null
watch(searchText, (newValue) => {
  if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  searchDebounceTimer = setTimeout(() => {
    emit('search', newValue)
  }, 500)
})

// 搜索
const onSearch = () => {
  // 通知父组件搜索条件已更新
  emit('search', searchText.value)
}

// 检查器筛选相关
const checkerFilter = ref('')

// 检查器筛选变化
const onCheckerFilterChange = () => {
  // 通知父组件筛选条件已更新
  emit('filterChange', { checker: checkerFilter.value, status: statusFilter.value })
}

// 状态筛选相关
const statusFilter = ref('')

// 状态筛选变化
const onStatusFilterChange = () => {
  // 通知父组件筛选条件已更新
  emit('filterChange', { checker: checkerFilter.value, status: statusFilter.value })
}

// 获取下拉框容器的安全方法
const getPopupContainer = () => {
  return document?.body || document.documentElement
}

// 导出搜索文本和筛选器供父组件使用
// 确保这些变量被使用
console.log('searchText:', searchText.value)
console.log('checkerFilter:', checkerFilter.value)
console.log('statusFilter:', statusFilter.value)

defineExpose({
  searchText,
  checkerFilter,
  statusFilter
})
</script>

<style scoped>
/* 组件局部样式 */
:deep(.checker-filter-dropdown) {
  min-width: 100px !important;
  max-height: none !important;
  overflow: visible !important;
  height: auto !important;
}
:deep(.checker-filter-dropdown .ant-select-dropdown-menu) {
  max-height: none !important;
  overflow: visible !important;
  height: auto !important;
}
:deep(.checker-filter-dropdown .ant-select-dropdown-menu-item) {
  white-space: nowrap;
}
:deep(.checker-filter-dropdown .ant-select-dropdown-menu-list) {
  max-height: none !important;
  overflow: visible !important;
  height: auto !important;
}
:deep(.checker-filter-dropdown .ant-select-item) {
  white-space: nowrap;
}
:deep(.checker-filter-dropdown .rc-virtual-list) {
  max-height: none !important;
  overflow: visible !important;
  height: auto !important;
}
:deep(.checker-filter-dropdown .rc-virtual-list-holder) {
  max-height: none !important;
  overflow: visible !important;
  height: auto !important;
}
</style>
