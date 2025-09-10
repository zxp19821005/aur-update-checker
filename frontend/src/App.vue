<template>
  <a-config-provider :locale="zhCN">
    <a-layout class="layout" style="min-height: 100vh; --sider-width: 40px;" :style="collapsed ? '--sider-width: 40px;' : '--sider-width: 200px;'">
      <a-layout-sider v-model:collapsed="collapsed" collapsible :collapsed-width="40" :width="120" class="fixed-sider">
        <div class="logo" />
        <a-menu
          v-model:selectedKeys="selectedKeys"
          v-model:openKeys="openKeys"
          theme="dark"
          mode="inline"
        >
          <a-menu-item key="dashboard">
            <dashboard-outlined />
            <span>仪表盘</span>
            <router-link to="/"></router-link>
          </a-menu-item>
          <a-menu-item key="packages">
            <appstore-outlined />
            <span>软件包管理</span>
            <router-link to="/packages"></router-link>
          </a-menu-item>
          <a-menu-item key="settings">
            <setting-outlined />
            <span>设置</span>
            <router-link to="/settings"></router-link>
          </a-menu-item>
          <a-menu-item key="logs">
            <file-text-outlined />
            <span>日志</span>
            <router-link to="/logs"></router-link>
          </a-menu-item>
          <a-menu-item key="about">
            <info-circle-outlined />
            <span>关于</span>
            <router-link to="/about"></router-link>
          </a-menu-item>
        </a-menu>
      </a-layout-sider>
      <a-layout style="flex-direction: column;">

        <a-layout-content style="margin: 0; padding: 0; flex: 1; overflow: hidden;">
          <div class="page-container" style="height: 100%;">
            <router-view />
          </div>
        </a-layout-content>
      </a-layout>
    </a-layout>


  </a-config-provider>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  DashboardOutlined,
  AppstoreOutlined,
  SettingOutlined,
  FileTextOutlined,
  InfoCircleOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined
} from '@ant-design/icons-vue'
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import { usePackageStore } from '@/stores/package'
import { useTimerStore } from '@/stores/timer'

const collapsed = ref(true)
const selectedKeys = ref(['dashboard'])
const openKeys = ref([])


const route = useRoute()
const packageStore = usePackageStore()
const timerStore = useTimerStore()

// 当前页面标题
const currentBreadcrumb = computed(() => {
  const pathMap = {
    '/': '仪表盘',
    '/packages': '软件包管理',
    '/settings': '设置',
    '/logs': '日志',
    '/about': '关于'
  }
  return pathMap[route.path] || '未知页面'
})

// 定时任务状态
const timerStatus = computed(() => timerStore.status)

// 监听路由变化，更新选中的菜单项
watch(
  () => route.path,
  (path) => {
    const pathMap = {
      '/': 'dashboard',
      '/packages': 'packages',
      '/settings': 'settings',
      '/logs': 'logs',
      '/about': 'about'
    }
    selectedKeys.value = [pathMap[path] || 'dashboard']
  },
  { immediate: true }
)



// 组件挂载时加载数据
onMounted(async () => {
  // 加载所有软件包
  await packageStore.fetchPackages()

  // 获取定时任务状态
  await timerStore.fetchTimerStatus()
})
</script>

<style scoped>
.trigger {
  font-size: 18px;
  line-height: 64px;
  padding: 0 24px;
  cursor: pointer;
  transition: color 0.3s;
}

.trigger:hover {
  color: #1890ff;
}

.logo {
  height: 32px;
  margin: 8px;
  background-image: url('./assets/icon.png');
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
}

/* 固定左侧导航栏 */
.fixed-sider {
  position: fixed !important;
  height: 100vh !important;
  left: 0 !important;
  top: 0 !important;
  z-index: 1000 !important;
}

/* 减小左侧导航栏的宽度 */
.layout .ant-layout-sider {
  width: 40px !important;
  min-width: 40px !important;
  max-width: 40px !important;
  flex: 0 0 40px !important;
}

/* 导航栏展开后的宽度 */
.layout .ant-layout-sider:not(.ant-layout-sider-collapsed) {
  width: 200px !important;
  min-width: 200px !important;
  max-width: 200px !important;
  flex: 0 0 200px !important;
}

/* 确保菜单项在窄宽度下也能正确显示 */
.layout .ant-menu-inline-collapsed {
  width: 40px !important;
}

.layout .ant-menu-inline {
  width: 200px !important;
}

.site-layout .site-layout-background {
  background: #fff;
}

.page-container {
  padding: 0;
  background: #fff;
  min-height: auto;
  height: auto;
  border-radius: 0;
  box-shadow: none;
  overflow: hidden;
}

/* 确保整个应用不会出现滚动条，但允许日志页面内部滚动 */
html, body {
  overflow: hidden;
  height: 100%;
}

/* 日志页面特殊处理 */
.logs-page .log-container {
  overflow-y: auto !important;
}

#app {
  height: 100vh;
  overflow: hidden;
}

/* 调整主内容区域，为固定侧边栏留出空间 */
.layout .ant-layout {
  margin-left: 40px; /* 与折叠后的侧边栏宽度相同 */
  transition: margin-left 0.2s;
}

/* 当侧边栏展开时，调整主内容区域的左边距 */
.layout .ant-layout-sider:not(.ant-layout-sider-collapsed) + .ant-layout {
  margin-left: 200px; /* 与展开后的侧边栏宽度相同 */
}
</style>
