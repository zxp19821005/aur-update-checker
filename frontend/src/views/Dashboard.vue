<template>
    <a-row :gutter="16">
      <a-col :span="17">
        <a-row :gutter="16" style="margin-bottom: 4px;">
          <a-col :span="12">
            <a-card>
              <a-statistic
                title="软件包总数"
                :value="packageStore.packages.length"
                :value-style="{ color: '#3f8600' }"
              >
                <template #prefix>
                  <appstore-outlined />
                </template>
              </a-statistic>
            </a-card>
          </a-col>
          <a-col :span="12">
            <a-card>
              <a-statistic
                title="需要更新"
                :value="packageStore.outdatedPackagesCount"
                :value-style="{ color: '#cf1322' }"
              >
                <template #prefix>
                  <exclamation-circle-outlined />
                </template>
              </a-statistic>
            </a-card>
          </a-col>
          <a-col :span="12">
            <a-card>
              <a-statistic
                title="检查失败"
                :value="packageStore.failedPackagesCount"
                :value-style="{ color: '#d4380d' }"
              >
                <template #prefix>
                  <close-circle-outlined />
                </template>
              </a-statistic>
            </a-card>
          </a-col>
          <a-col :span="12">
            <a-card>
              <a-statistic
                title="未检查"
                :value="packageStore.uncheckedPackagesCount"
                :value-style="{ color: '#d48806' }"
              >
                <template #prefix>
                  <clock-circle-outlined />
                </template>
              </a-statistic>
            </a-card>
          </a-col>
        </a-row>
      </a-col>
      <a-col :span="7">
        <a-card title="快速操作" style="height: 100%">
          <a-space direction="vertical" style="width: 100%">
            <div style="display: flex; justify-content: space-between;">
              <a-button @click="checkAllAurVersions" :loading="checkingAur" title="检查所有AUR版本">
                <template #icon><sync-outlined /></template>
              </a-button>
              <a-button @click="checkAllUpstreamVersions" :loading="checkingUpstream" title="检查所有上游版本">
                <template #icon><cloud-download-outlined /></template>
              </a-button>
              <a-button @click="checkAllVersions" :loading="checkingAll" title="检查所有AUR版本和上游版本">
                <template #icon><reload-outlined /></template>
              </a-button>
            </div>
            <a-button block @click="$router.push('/packages')">
              <template #icon><appstore-outlined /></template>
              管理软件包
            </a-button>
            <a-button block @click="$router.push('/settings')">
              <template #icon><setting-outlined /></template>
              设置定时任务
            </a-button>
          </a-space>
        </a-card>
      </a-col>
    </a-row>



    <a-row style="margin-top: 20px;">
      <a-col :span="24">
        <a-card title="需要更新的软件包">
          <a-table
            :columns="columns"
            :data-source="outdatedPackages"
            :loading="packageStore.loading"
            rowKey="id"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                {{ record.name }}
              </template>
              <template v-else-if="column.key === 'aurVersion'">
                <a-tag :color="getVersionTagColor(record.aurUpdateState)">
                  {{ record.aurVersion || '-' }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'upstreamVersion'">
                <a-tag :color="getVersionTagColor(record.upstreamUpdateState)">
                  {{ record.upstreamVersion || '-' }}
                </a-tag>
              </template>
              <template v-else-if="column.key === 'action'">
                <a-space>
                  <a-button
                    type="text"
                    size="small"
                    @click="checkPackageAurVersion(record.id)"
                    :disabled="checkingAurForPackage === record.id"
                  >
                    <a-spin v-if="checkingAurForPackage === record.id" size="small" />
                    <sync-outlined v-else />
                  </a-button>
                  <a-button
                    type="text"
                    size="small"
                    @click="checkPackageUpstreamVersion(record.id)"
                    :disabled="checkingUpstreamForPackage === record.id"
                  >
                    <a-spin v-if="checkingUpstreamForPackage === record.id" size="small" />
                    <cloud-download-outlined v-else />
                  </a-button>
                  <a-button
                    type="text"
                    size="small"
                    @click="checkPackageAllVersions(record.id)"
                    :disabled="checkingAurForPackage === record.id || checkingUpstreamForPackage === record.id"
                  >
                    <a-spin v-if="checkingAurForPackage === record.id && checkingUpstreamForPackage === record.id" size="small" />
                    <reload-outlined v-else />
                  </a-button>
                </a-space>
              </template>
            </template>
          </a-table>
        </a-card>
      </a-col>
    </a-row>

    <!-- 定时任务设置对话框 -->
    <a-modal
      v-model:open="timerModalVisible"
      title="设置定时任务"
      @ok="startTimerTask"
      @cancel="timerModalVisible = false"
      :confirmLoading="timerStore.loading"
    >
      <a-form :model="timerForm" layout="vertical">
        <a-form-item label="检查间隔（分钟）" name="intervalMinutes" :rules="[{ required: true, message: '请输入检查间隔' }]">
          <a-input-number
            v-model:value="timerForm.intervalMinutes"
            :min="1"
            :max="1440"
            style="width: 100%"
          />
        </a-form-item>
        <a-alert
          message="定时任务将在后台运行，定期检查所有软件包的AUR版本和上游版本。"
          type="info"
          show-icon
        />
      </a-form>
    </a-modal>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  AppstoreOutlined,
  ExclamationCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  SettingOutlined,
  CloudDownloadOutlined,
  ReloadOutlined
} from '@ant-design/icons-vue'
import { usePackageStore } from '@/stores/package'
import { useTimerStore } from '@/stores/timer'
import { checkAllAurVersions, checkAllUpstreamVersions, checkAurVersion, checkUpstreamVersion } from '@/services/api'

const packageStore = usePackageStore()
const timerStore = useTimerStore()

// 定时任务相关
const timerModalVisible = ref(false)
const timerForm = ref({
  intervalMinutes: 60
})

// 加载状态
const checkingAur = ref(false)
const checkingUpstream = ref(false)
const checkingAll = ref(false)
const checkingAurForPackage = ref(null)
const checkingUpstreamForPackage = ref(null)

// 表格列定义
const columns = [
  {
    title: '软件包名称',
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: 'AUR版本',
    dataIndex: 'aurVersion',
    key: 'aurVersion',
  },
  {
    title: '上游版本',
    dataIndex: 'upstreamVersion',
    key: 'upstreamVersion',
  },
  {
    title: '操作',
    key: 'action',
    width: 220,
  },
]

// 需要更新的软件包
const outdatedPackages = computed(() => {
  return packageStore.packages.filter(pkg => {
    if (!pkg.aurVersion || !pkg.upstreamVersion) return false
    return pkg.aurVersion !== pkg.upstreamVersion
  })
})

// 获取版本标签颜色
const getVersionTagColor = (state) => {
  switch (state) {
    case 0: return 'default'  // 未检查
    case 1: return 'green'    // 成功
    case 2: return 'red'      // 失败
    default: return 'default'
  }
}

// 显示定时任务设置对话框
const showTimerModal = () => {
  timerModalVisible.value = true
}

// 启动定时任务
const startTimerTask = async () => {
  try {
    await timerStore.startTimerTask(timerForm.value.intervalMinutes)
    timerModalVisible.value = false
  } catch (error) {
    // 错误已在store中处理
  }
}

// 停止定时任务
const stopTimerTask = async () => {
  try {
    await timerStore.stopTimerTask()
  } catch (error) {
    // 错误已在store中处理
  }
}

// 检查所有AUR版本
const checkAllAurVersionsHandler = async () => {
  checkingAur.value = true
  try {
    await checkAllAurVersions()
    await packageStore.fetchPackages()
    message.success('AUR版本检查完成')
  } catch (error) {
    message.error('AUR版本检查失败')
    console.error('AUR版本检查失败:', error)
  } finally {
    checkingAur.value = false
  }
}

// 检查所有上游版本
const checkAllUpstreamVersionsHandler = async () => {
  checkingUpstream.value = true
  try {
    await checkAllUpstreamVersions()
    await packageStore.fetchPackages()
    message.success('上游版本检查完成')
  } catch (error) {
    message.error('上游版本检查失败')
    console.error('上游版本检查失败:', error)
  } finally {
    checkingUpstream.value = false
  }
}

// 检查所有版本 (AUR和上游)
const checkAllVersions = async () => {
  checkingAll.value = true
  try {
    // 先检查AUR版本
    checkingAur.value = true
    try {
      await checkAllAurVersions()
      message.success('AUR版本检查完成')
    } catch (error) {
      message.error('AUR版本检查失败')
      console.error('AUR版本检查失败:', error)
    } finally {
      checkingAur.value = false
    }

    // 然后检查上游版本
    checkingUpstream.value = true
    try {
      await checkAllUpstreamVersions()
      message.success('上游版本检查完成')
    } catch (error) {
      message.error('上游版本检查失败')
      console.error('上游版本检查失败:', error)
    } finally {
      checkingUpstream.value = false
    }

    // 刷新软件包列表
    await packageStore.fetchPackages()
    message.success('所有版本检查完成')
  } catch (error) {
    message.error('版本检查过程中发生错误')
    console.error('版本检查过程中发生错误:', error)
  } finally {
    checkingAll.value = false
  }
}

// 检查单个软件包的AUR版本
const checkPackageAurVersion = async (packageId) => {
  checkingAurForPackage.value = packageId
  try {
    await checkAurVersion(packageId)
    await packageStore.fetchPackages()
    message.success('AUR版本检查完成')
  } catch (error) {
    message.error('AUR版本检查失败')
    console.error('AUR版本检查失败:', error)
  } finally {
    checkingAurForPackage.value = null
  }
}

// 检查单个软件包的上游版本
const checkPackageUpstreamVersion = async (packageId) => {
  checkingUpstreamForPackage.value = packageId
  try {
    await checkUpstreamVersion(packageId)
    await packageStore.fetchPackages()
    message.success('上游版本检查完成')
  } catch (error) {
    message.error('上游版本检查失败')
    console.error('上游版本检查失败:', error)
  } finally {
    checkingUpstreamForPackage.value = null
  }
}

// 检查单个软件包的所有版本
const checkPackageAllVersions = async (packageId) => {
  checkingAurForPackage.value = packageId
  checkingUpstreamForPackage.value = packageId
  let aurCheckSuccess = false
  let upstreamCheckSuccess = false

  try {
    // 尝试检查AUR版本
    await checkAurVersion(packageId)
    aurCheckSuccess = true
  } catch (error) {
    console.error('AUR版本检查失败:', error)
    message.warning('AUR版本检查失败，将尝试检查上游版本')
  }

  try {
    // 尝试检查上游版本
    await checkUpstreamVersion(packageId)
    upstreamCheckSuccess = true
  } catch (error) {
    console.error('上游版本检查失败:', error)
    message.warning('上游版本检查失败')
  }

  // 无论成功与否，都刷新软件包列表
  await packageStore.fetchPackages()

  // 根据检查结果提供反馈
  if (aurCheckSuccess && upstreamCheckSuccess) {
    message.success('所有版本检查完成')
  } else if (aurCheckSuccess || upstreamCheckSuccess) {
    message.warning('部分版本检查完成')
  } else {
    message.error('所有版本检查均失败')
  }

  checkingAurForPackage.value = null
  checkingUpstreamForPackage.value = null
}

// 组件挂载时加载数据
onMounted(async () => {
  // 获取定时任务状态
  await timerStore.fetchTimerStatus()
})
</script>

<style scoped>
.ant-row {
  margin: 0 !important;
}

.ant-card {
  margin-bottom: 16px;
}

.ant-table-wrapper {
  overflow-x: auto;
}

:deep(.ant-table) {
  width: 100%;
}

:deep(.ant-card-body) {
  padding: 12px;
}

:deep(.ant-statistic-content) {
  display: flex;
  align-items: center;
}

/* 使卡片标题居中 */
:deep(.ant-card-head) {
  text-align: center;
}

:deep(.ant-card-head-title) {
  display: block;
  width: 100%;
  text-align: center;
}

/* 使表格内容居中 */
:deep(.ant-table-thead > tr > th),
:deep(.ant-table-tbody > tr > td) {
  text-align: center;
}

/* 使操作按钮居中 */
:deep(.ant-space) {
  justify-content: center;
}

/* 使快速操作内容居中 */
:deep(.ant-space-vertical) {
  align-items: center;
}
</style>