<template>
  <a-modal
    :open="open"
    :title="title"
    @ok="handleOk"
    @cancel="handleCancel"
    @update:open="handleUpdateOpen"
    :confirmLoading="loading"
    :destroyOnClose="true"
  >
    <a-form
      ref="formRef"
      :model="formState"
      :rules="formRules"
      layout="vertical"
    >
      <a-form-item label="软件包名称" name="name">
        <a-input v-model:value="formState.name" placeholder="请输入软件包名称" />
      </a-form-item>
      <a-form-item label="上游URL" name="upstreamUrl">
        <a-input v-model:value="formState.upstreamUrl" placeholder="请输入上游URL" />
      </a-form-item>
      <a-form-item label="上游检查器" name="upstreamChecker">
        <a-select
          v-model:value="formState.upstreamChecker"
          placeholder="请选择上游检查器"
          :options="checkerOptions"
          style="width: 100%"
        />
      </a-form-item>
      <a-form-item label="版本提取关键字" name="versionExtractKey">
        <a-input v-model:value="formState.versionExtractKey" placeholder="请输入版本提取关键字" />
      </a-form-item>
      <a-form-item name="checkTestVersion">
        <template #label>
          <span>是否检查测试版本</span>
          <a-tooltip title="勾选此项将检查包括alpha、beta等在内的测试版本">
            <question-circle-outlined style="margin-left: 4px; color: rgba(0, 0, 0, 0.45)" />
          </a-tooltip>
        </template>
        <a-checkbox v-model:checked="formState.checkTestVersion">
          检查测试版本（如alpha、beta、rc等）
        </a-checkbox>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { message } from 'ant-design-vue'
import { QuestionCircleOutlined } from '@ant-design/icons-vue'
import { getUpstreamCheckers } from '@/services/api'

// Props 定义
const props = defineProps({
  open: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: '添加软件包'
  },
  isEdit: {
    type: Boolean,
    default: false
  },
  packageData: {
    type: Object,
    default: () => null
  }
})

// Emits 定义
const emit = defineEmits(['update:open', 'ok', 'cancel'])

// 表单引用
const formRef = ref()

// 加载状态
const loading = ref(false)

// 检查器选项
const checkerOptions = ref([
  { label: 'Curl', value: 'curl' },
  { label: 'Gitee', value: 'gitee' },
  { label: 'GitHub', value: 'github' },
  { label: 'GitLab', value: 'gitlab' },
  { label: 'HTTP', value: 'http' },
  { label: 'JSON', value: 'json' },
  { label: 'NPM', value: 'npm' },
  { label: 'Playwright', value: 'playwright' },
  { label: 'PyPI', value: 'pypi' },
  { label: 'Redirect', value: 'redirect' }
])

// 组件创建时设置默认检查器

// 表单数据
const formState = reactive({
  id: null,
  name: '',
  upstreamUrl: '',
  versionExtractKey: '',
  upstreamChecker: '',
  checkTestVersion: false
})

// 表单验证规则
const formRules = {
  name: [
    { required: true, message: '请输入软件包名称', trigger: 'blur' },
  ],
  upstreamUrl: [
    { required: true, message: '请输入上游URL', trigger: 'blur' },
    { type: 'url', message: '请输入有效的URL', trigger: 'blur' }
  ],
  versionExtractKey: [
    // 版本提取关键字不是必填项
  ],
  upstreamChecker: [
    { required: true, message: '请选择上游检查器', trigger: 'change' },
  ]
}

// 加载检查器选项
const loadCheckerOptions = async () => {
  try {
    // 从API获取检查器列表
    const checkers = await getUpstreamCheckers()
    if (checkers && checkers.length > 0) {
      checkerOptions.value = checkers.map(checker => ({
        label: checker,
        value: checker
      }))
    }
  } catch (error) {
    console.error('加载检查器选项失败:', error)
  } finally {
    // 如果没有选择检查器，默认选择第一个
    if (!formState.upstreamChecker && checkerOptions.value.length > 0) {
      formState.upstreamChecker = checkerOptions.value[0].value
    }
  }
}

// 重置表单
const resetForm = () => {
  formState.id = null
  formState.name = ''
  formState.upstreamUrl = ''
  formState.versionExtractKey = ''
  formState.upstreamChecker = ''
  formState.checkTestVersion = false

  // 重置表单验证
  if (formRef.value && typeof formRef.value.clearValidate === 'function') {
    formRef.value.clearValidate()
  }

  // 确保重置后也有默认检查器选项
  if (!formState.upstreamChecker && checkerOptions.value.length > 0) {
    formState.upstreamChecker = checkerOptions.value[0].value
  }
}

// 处理确定按钮点击
const handleOk = async () => {
  try {
    // 验证表单
    if (formRef.value && typeof formRef.value.validate === 'function') {
      await formRef.value.validate()
    }

    // 验证数据
    if (!formState.name) {
      message.error('软件包名称不能为空')
      return
    }
    if (!formState.upstreamUrl) {
      message.error('上游URL不能为空')
      return
    }
    if (!formState.upstreamChecker) {
      message.error('请选择上游检查器')
      return
    }

    // 触发提交事件
    const formData = {
      id: formState.id,
      name: formState.name,
      upstreamUrl: formState.upstreamUrl,
      versionExtractKey: formState.versionExtractKey || '',
      upstreamChecker: formState.upstreamChecker,
      checkTestVersion: formState.checkTestVersion ? 1 : 0
    }
    console.log('表单提交 checkTestVersion:', formState.checkTestVersion, '->', formData.checkTestVersion)

    // 如果是编辑模式，并且有原始数据，则保留 AUR 相关字段
    if (props.isEdit && props.packageData) {
      // 保留 AUR 版本和更新时间
      if (props.packageData.aurVersion) formData.aurVersion = props.packageData.aurVersion
      if (props.packageData.aurUpdateDate) formData.aurUpdateDate = props.packageData.aurUpdateDate
      if (props.packageData.aurUpdateState !== undefined) formData.aurUpdateState = props.packageData.aurUpdateState

      // 保留上游版本和更新时间
      if (props.packageData.upstreamVersion) formData.upstreamVersion = props.packageData.upstreamVersion
      if (props.packageData.upstreamUpdateDate) formData.upstreamUpdateDate = props.packageData.upstreamUpdateDate
      if (props.packageData.upstreamUpdateState !== undefined) formData.upstreamUpdateState = props.packageData.upstreamUpdateState

      // 保留创建和更新时间
      if (props.packageData.createdAt) formData.createdAt = props.packageData.createdAt
      if (props.packageData.updatedAt) formData.updatedAt = props.packageData.updatedAt

    }

    emit('ok', formData)

    // 提交成功后重置表单
    resetForm()
  } catch (error) {
    console.error('表单验证失败:', error)
    message.error('表单验证失败，请检查输入')
  }
}

// 处理取消按钮点击
const handleCancel = () => {
  // 取消时也重置表单
  resetForm()
  emit('cancel')
}

// 处理open属性更新
const handleUpdateOpen = (value) => {
  emit('update:open', value)
}

// 监听 open 属性变化
watch(() => props.open, (newVal, oldVal) => {
  if (newVal) {

    // 先重置表单
    resetForm()

    // 弹窗显示时，加载数据
    loadCheckerOptions().then(() => {
      if (props.isEdit && props.packageData) {
        // 编辑模式，填充表单数据

        // 直接填充表单数据
        formState.id = props.packageData.id
        formState.name = props.packageData.name || ''
        formState.upstreamUrl = props.packageData.upstreamUrl || ''
        formState.versionExtractKey = props.packageData.versionExtractKey || ''
        formState.upstreamChecker = props.packageData.upstreamChecker || ''
        formState.checkTestVersion = Boolean(props.packageData.checkTestVersion)
        console.log('设置 checkTestVersion:', props.packageData.checkTestVersion, '->', formState.checkTestVersion)

        // 确保表单验证状态也被重置
        if (formRef.value) {
          formRef.value.clearValidate()
        }
      } else {
        // 添加模式，确保有默认检查器选项
        if (!formState.upstreamChecker && checkerOptions.value.length > 0) {
          formState.upstreamChecker = checkerOptions.value[0].value
        }
      }
    })
  } else {
    // 弹窗关闭时，重置表单
    resetForm()
  }
})

// 监听 isEdit 和 packageData 的变化
watch(() => [props.isEdit, props.packageData], ([newIsEdit, newPackageData], [oldIsEdit, oldPackageData]) => {
  // 如果弹窗已打开且是编辑模式且有packageData，则填充表单
  if (props.open && newIsEdit && newPackageData) {
    formState.id = newPackageData.id
    formState.name = newPackageData.name || ''
    formState.upstreamUrl = newPackageData.upstreamUrl || ''
    formState.versionExtractKey = newPackageData.versionExtractKey || ''
    formState.upstreamChecker = newPackageData.upstreamChecker || ''
    formState.checkTestVersion = Boolean(newPackageData.checkTestVersion)
    console.log('watch 中设置 checkTestVersion:', newPackageData.checkTestVersion, '->', formState.checkTestVersion)
  }
}, { deep: true })


</script>

<style>
/* 自定义模态框位置，避免太靠下 */
body .ant-modal-wrap {
  top: 20px !important;
  position: fixed !important;
}

body .ant-modal {
  max-height: calc(100vh - 100px) !important;
  overflow-y: auto !important;
  margin-top: 0 !important;
  top: 20px !important;
  /* 隐藏滚动条但保持滚动功能 */
  scrollbar-width: none !important;
  -ms-overflow-style: none !important;
}

/* 隐藏Webkit浏览器的滚动条 */
body .ant-modal::-webkit-scrollbar {
  display: none !important;
  width: 0 !important;
}

/* 减小表单项间距 */
body .ant-modal .ant-form {
  padding: 5px 0 !important;
}

body .ant-modal .ant-form-item {
  margin-bottom: 12px !important;
}

body .ant-modal-body {
  padding: 16px 24px !important;
}

/* 调整底部按钮位置 */
body .ant-modal-footer {
  margin-top: -25px !important;
  padding: 10px 24px !important;
}

/* 自定义下拉框样式，确保不出现滚动条 */
.ant-select-dropdown {
  min-width: 150px !important;
  max-height: none !important;
  overflow: visible !important;
  height: auto !important;
}
.ant-select-dropdown .ant-select-dropdown-menu {
  max-height: none !important;
  overflow: visible !important;
  height: auto !important;
}
.ant-select-dropdown .ant-select-dropdown-menu-item {
  white-space: nowrap;
}
.ant-select-dropdown .ant-select-dropdown-menu-list {
  max-height: none !important;
  overflow: visible !important;
  height: auto !important;
}
.ant-select-dropdown .ant-select-item {
  white-space: nowrap;
}
.ant-select-dropdown .rc-virtual-list {
  max-height: none !important;
  overflow: visible !important;
  height: auto !important;
}
.ant-select-dropdown .rc-virtual-list-holder {
  max-height: none !important;
  overflow: visible !important;
  height: auto !important;
}
.ant-select-dropdown .rc-virtual-list-holder-inner {
  max-height: none !important;
  overflow: visible !important;
  height: auto !important;
}
/* 确保下拉框内容完全显示 */
.ant-select-dropdown .ant-select-item-option {
  height: auto !important;
  line-height: normal !important;
  padding: 8px 12px !important;
  white-space: nowrap;
}
.ant-select-dropdown .ant-select-item-option-content {
  white-space: nowrap;
}
.ant-select-dropdown .ant-select-item-option-state {
  display: none;
}
</style>