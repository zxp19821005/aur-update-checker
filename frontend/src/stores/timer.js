import { defineStore } from 'pinia'
// 不再导入 message，使用全局的 window.$message
import { startTimerTask, stopTimerTask, getTimerTaskStatus } from '@/services/api'

export const useTimerStore = defineStore('timer', {
  state: () => ({
    status: {
      isRunning: false,
      intervalMinutes: 0
    },
    loading: false
  }),

  actions: {
    // 获取定时任务状态
    async fetchTimerStatus() {
      this.loading = true
      try {
        const response = await getTimerTaskStatus()
        this.status = response || {
          isRunning: false,
          intervalMinutes: 0
        }
      } catch (error) {
        window.$message.error('获取定时任务状态失败')
        console.error('获取定时任务状态失败:', error)
        // 确保在错误情况下status也有默认值
        this.status = {
          isRunning: false,
          intervalMinutes: 0
        }
      } finally {
        this.loading = false
      }
    },

    // 启动定时任务
    async startTimerTask(intervalMinutes) {
      this.loading = true
      try {
        await startTimerTask(intervalMinutes)
        this.status.isRunning = true
        this.status.intervalMinutes = intervalMinutes
        window.$message.success('定时任务已启动')
      } catch (error) {
        window.$message.error('启动定时任务失败')
        console.error('启动定时任务失败:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    // 停止定时任务
    async stopTimerTask() {
      this.loading = true
      try {
        await stopTimerTask()
        this.status.isRunning = false
        window.$message.success('定时任务已停止')
      } catch (error) {
        window.$message.error('停止定时任务失败')
        console.error('停止定时任务失败:', error)
        throw error
      } finally {
        this.loading = false
      }
    }
  }
})
