<template>
  <a-card>
    <a-tabs v-model:activeKey="activeKey">
      <a-tab-pane key="timer" tab="定时任务">
        <a-form
          :model="timerForm"
          layout="vertical"
        >
          <a-form-item label="定时任务状态" v-if="timerStore.status.isRunning">
            <a-tag color="green">
              运行中
            </a-tag>
          </a-form-item>

          <a-form-item label="检查间隔（分钟）" v-if="!timerStore.status.isRunning">
            <a-input-number
              v-model:value="timerForm.intervalMinutes"
              :min="1"
              :max="1440"
              style="width: 100%"
            />
          </a-form-item>

          <a-form-item label="下次检查时间" v-if="timerStore.status.isRunning">
            <a>{{ nextCheckTime }}</a>
          </a-form-item>

          <a-form-item>
            <a-space>
              <a-button
                v-if="!timerStore.status.isRunning"
                type="primary"
                @click="startTimerTask"
                :loading="timerStore.loading"
              >
                启动定时任务
              </a-button>
              <a-button
                v-else
                danger
                @click="stopTimerTask"
                :loading="timerStore.loading"
              >
                停止定时任务
              </a-button>
              <a-button @click="checkNow" :loading="checking">
                立即检查
              </a-button>
            </a-space>
          </a-form-item>
        </a-form>

        <a-divider />

        <a-alert
          message="定时任务说明"
          description="定时任务将在后台运行，定期检查所有软件包的AUR版本和上游版本。检查结果将自动更新到软件包列表中。"
          type="info"
          show-icon
        />
      </a-tab-pane>

      <a-tab-pane key="app" tab="应用设置">
        <a-form
          :model="appSettings"
          layout="vertical"
        >
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="日志级别">
                <a-select v-model:value="appSettings.logLevel" style="width: 100%">
                  <a-select-option value="debug">Debug</a-select-option>
                  <a-select-option value="info">Info</a-select-option>
                  <a-select-option value="warn">Warn</a-select-option>
                  <a-select-option value="error">Error</a-select-option>
                </a-select>
              </a-form-item>

              <a-form-item label="并发检查数">
                <a-input-number
                  v-model:value="appSettings.maxConcurrentChecks"
                  :min="1"
                  :max="50"
                  style="width: 100%"
                />
                <div class="ant-form-item-explain">设置同时进行的版本检查数量</div>
              </a-form-item>

              <a-form-item label="检查超时时间（秒）">
                <a-input-number
                  v-model:value="appSettings.checkerTimeout"
                  :min="5"
                  :max="120"
                  style="width: 100%"
                />
                <div class="ant-form-item-explain">设置版本检查的超时时间</div>
              </a-form-item>
            </a-col>
            
            <a-col :span="12">
              <a-form-item label="重试次数">
                <a-input-number
                  v-model:value="appSettings.retryCount"
                  :min="0"
                  :max="5"
                  style="width: 100%"
                />
                <div class="ant-form-item-explain">设置版本检查失败时的重试次数</div>
              </a-form-item>

              <a-form-item label="缓存时间（分钟）">
                <a-input-number
                  v-model:value="appSettings.cacheTTL"
                  :min="0"
                  :max="1440"
                  style="width: 100%"
                />
                <div class="ant-form-item-explain">设置版本检查结果的缓存时间，0表示不缓存</div>
              </a-form-item>

              <a-form-item label="自动检查">
                <a-switch v-model:checked="appSettings.autoCheck" />
                <span style="margin-left: 8px;">启动应用时自动检查所有软件包版本</span>
              </a-form-item>
            </a-col>
          </a-row>

          <a-divider />

          <a-form-item>
            <a-button type="primary" @click="saveAppSettings" :loading="saving">
              保存设置
            </a-button>
          </a-form-item>
        </a-form>
      </a-tab-pane>

      <a-tab-pane key="checker" tab="检查器设置">
        <a-form
          :model="checkerSettings"
          layout="vertical"
        >
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="默认检查器">
                <a-select
                  v-model:value="checkerSettings.defaultChecker"
                  style="width: 100%"
                  :dropdownMatchSelectWidth="false"
                  :getPopupContainer="(trigger) => trigger.parentNode"
                  :virtual="false"
                  popupClassName="checker-filter-dropdown"
                  :listHeight="1000"
                  :dropdownStyle="{ maxHeight: 'none', overflow: 'visible' }"
                >
                  <a-select-option value="auto">自动选择</a-select-option>
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
              </a-form-item>
            </a-col>
            
            <a-col :span="12">
              <a-form-item label="GitHub API Token">
                <a-input-password v-model:value="checkerSettings.githubToken" placeholder="输入GitHub API Token（提高API限制）" />
              </a-form-item>

              <a-form-item label="GitLab API Token">
                <a-input-password v-model:value="checkerSettings.gitlabToken" placeholder="输入GitLab API Token（提高API限制）" />
              </a-form-item>
            </a-col>
          </a-row>

          <a-divider />

          <a-form-item>
            <a-button type="primary" @click="saveCheckerSettings" :loading="savingCheckerSettings">
              保存检查器设置
            </a-button>
          </a-form-item>
        </a-form>
      </a-tab-pane>
    </a-tabs>
  </a-card>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import { useTimerStore } from '@/stores/timer'
import { usePackageStore } from '@/stores/package'
import { checkAllAurVersions, checkAllUpstreamVersions } from '@/services/api'

const timerStore = useTimerStore()
const packageStore = usePackageStore()

// 当前激活的标签页
const activeKey = ref('timer')

// 定时任务表单
const timerForm = reactive({
  intervalMinutes: 60
})

// 应用设置
const appSettings = reactive({
  logLevel: 'info',
  maxConcurrentChecks: 10,
  checkerTimeout: 30,
  retryCount: 3,
  cacheTTL: 5,
  autoCheck: true
})

// 检查器设置
const checkerSettings = reactive({
  defaultChecker: 'auto',
  githubToken: '',
  gitlabToken: ''
})

// 加载状态
const checking = ref(false)
const saving = ref(false)
const savingCheckerSettings = ref(false)

// 下次检查时间
const nextCheckTime = computed(() => {
  if (!timerStore.status.isRunning) return '-'

  // 这里简化处理，实际应用中应该记录上次检查时间并计算下次检查时间
  return dayjs().add(timerStore.status.intervalMinutes, 'minute').format('YYYY-MM-DD HH:mm:ss')
})

// 启动定时任务
const startTimerTask = async () => {
  try {
    await timerStore.startTimerTask(timerForm.intervalMinutes)
    message.success('定时任务已启动')
  } catch (error) {
    // 错误已在store中处理
  }
}

// 停止定时任务
const stopTimerTask = async () => {
  try {
    await timerStore.stopTimerTask()
    message.success('定时任务已停止')
  } catch (error) {
    // 错误已在store中处理
  }
}

// 立即检查
const checkNow = async () => {
  checking.value = true
  try {
    // 检查AUR版本
    await checkAllAurVersions()
    // 检查上游版本
    await checkAllUpstreamVersions()
    // 刷新软件包列表
    await packageStore.fetchPackages()
    message.success('版本检查完成')
  } catch (error) {
    message.error('版本检查失败')
    console.error('版本检查失败:', error)
  } finally {
    checking.value = false
  }
}

// 保存应用设置
const saveAppSettings = async () => {
  saving.value = true
  try {
    // 这里应该调用API保存设置到后端或本地存储
    // 为了简化示例，我们只保存到localStorage
    localStorage.setItem('appSettings', JSON.stringify(appSettings))

    message.success('设置已保存')
  } catch (error) {
    message.error('保存设置失败')
    console.error('保存设置失败:', error)
  } finally {
    saving.value = false
  }
}

// 加载应用设置
const loadAppSettings = () => {
  try {
    const settings = localStorage.getItem('appSettings')
    if (settings) {
      const parsedSettings = JSON.parse(settings)
      Object.assign(appSettings, parsedSettings)
    }
  } catch (error) {
    console.error('加载应用设置失败:', error)
  }
}

// 保存检查器设置
const saveCheckerSettings = async () => {
  savingCheckerSettings.value = true
  try {
    // 这里应该调用API保存设置到后端或本地存储
    // 为了简化示例，我们只保存到localStorage
    localStorage.setItem('checkerSettings', JSON.stringify(checkerSettings))

    message.success('检查器设置已保存')
  } catch (error) {
    message.error('保存检查器设置失败')
    console.error('保存检查器设置失败:', error)
  } finally {
    savingCheckerSettings.value = false
  }
}

// 加载检查器设置
const loadCheckerSettings = () => {
  try {
    const settings = localStorage.getItem('checkerSettings')
    if (settings) {
      const parsedSettings = JSON.parse(settings)
      Object.assign(checkerSettings, parsedSettings)
    }
  } catch (error) {
    console.error('加载检查器设置失败:', error)
  }
}

// 组件挂载时加载数据
onMounted(async () => {
  // 获取定时任务状态
  await timerStore.fetchTimerStatus()

  // 加载应用设置
  loadAppSettings()

  // 加载检查器设置
  loadCheckerSettings()
})
</script>

<style scoped>
.ant-form-item-explain {
  color: rgba(0, 0, 0, 0.45);
  font-size: 12px;
  margin-top: 4px;
}
</style>
