import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import { message } from "ant-design-vue"
import App from './App.vue'
import router from './router'
import 'ant-design-vue/dist/reset.css'


// 等待 Wails 初始化完成
const wailsInit = () => {
  return new Promise((resolve) => {
    // 检查是否在浏览器环境中
    const isBrowser = typeof window !== 'undefined' && window.location && window.location.hostname;

    if (window.go) {
      // Wails v2 已经初始化完成
      resolve()
    } else if (isBrowser) {
      // 在浏览器环境中，不等待 Wails 初始化
      resolve();
    } else {
      // 监听 Wails 初始化事件
      const checkWails = () => {
        if (window.go) {
          resolve()
        } else {
          setTimeout(checkWails, 100)
        }
      }
      checkWails()
    }
  })
}

// 初始化应用
const initApp = async () => {
  // 检查是否在浏览器环境中
  const isBrowser = typeof window !== 'undefined' && window.location && window.location.hostname;

  // 等待 Wails 初始化
  await wailsInit()
  
  // 如果 Wails API 未加载且不在浏览器环境中，显示警告
  if (!window.go && !isBrowser) {
    console.warn('Wails API 未加载，请确保在 Wails 环境中运行应用程序')
  }

  // 创建应用实例
  const app = createApp(App)
  
  // 使用插件
  app.use(createPinia())
  app.use(router)
  app.use(Antd)

  // 配置全局消息提示位置为底部
  message.config({
    top: '80vh', // 设置消息显示在底部
    duration: 3, // 持续时间3秒
    maxCount: 3  // 最多同时显示3条消息
  })

  const globalMessage = {
    success: (content) => {
      return message.success(content)
    },
    error: (content) => {
      return message.error(content)
    },
    info: (content) => {
      return message.info(content)
    },
    warning: (content) => {
      return message.warning(content)
    }
  }
  
  // 将全局消息对象添加到 app.config.globalProperties 和 window 对象上
  app.config.globalProperties.$message = globalMessage
  window.$message = globalMessage

  // 挂载应用
  app.mount('#app')
}

// 启动应用
initApp()
