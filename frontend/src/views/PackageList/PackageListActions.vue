<script>
import { ref, onMounted, watch, h } from 'vue'
import { usePackageActions } from './PackageListActions.js'
import { message, Modal } from 'ant-design-vue'
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'

export default {
  name: 'PackageListActions',
  setup() {
    // 列宽配置
    const columnWidths = ref({
      name: 200,
      versions: 250,
      actions: 230
    })

    // 列宽调整相关
    const resizingColumn = ref(null)
    const startX = ref(0)
    const startWidth = ref(0)
    
    // 确保这些变量被使用
    const useColumnWidths = () => {
      console.log('columnWidths:', columnWidths.value)
      return columnWidths.value
    }
    
    const useResizingColumn = () => {
      console.log('resizingColumn:', resizingColumn.value)
      return resizingColumn.value
    }

    // 从本地存储加载列宽配置
    onMounted(() => {
      const savedWidths = localStorage.getItem('packageListColumnWidths')
      if (savedWidths) {
        try {
          columnWidths.value = JSON.parse(savedWidths)
        } catch (e) {
          console.error('Failed to parse column widths from localStorage', e)
        }
      }
    })

    // 开始调整列宽
    const startColumnResize = (column, event) => {
      resizingColumn.value = column
      startX.value = event.clientX
      startWidth.value = columnWidths.value[column]
      document.addEventListener('mousemove', handleColumnResize)
      document.addEventListener('mouseup', stopColumnResize)
      event.preventDefault()
    }

    // 处理列宽调整
    const handleColumnResize = (event) => {
      if (!resizingColumn.value) return
      const diff = event.clientX - startX.value
      columnWidths.value[resizingColumn.value] = Math.max(100, startWidth.value + diff)
    }

    // 停止调整列宽
    const stopColumnResize = () => {
      resizingColumn.value = null
      document.removeEventListener('mousemove', handleColumnResize)
      document.removeEventListener('mouseup', stopColumnResize)
      // 保存列宽配置到本地存储
      localStorage.setItem('packageListColumnWidths', JSON.stringify(columnWidths.value))
    }

    // 使用操作逻辑
    const packageActions = usePackageActions()

    // 处理模态框打开状态变化
    const handleModalOpenChange = (open) => {
      packageActions.modalVisible.value = open
    }

    // 确认删除软件包
    const confirmDeletePackage = (pkg) => {
      // 检查 pkg 对象是否存在
      if (!pkg) {
        message.error('无效的软件包数据')
        return
      }

      // 检查 pkg.id 是否存在
      if (!pkg.id) {
        message.error('软件包ID不存在')
        return
      }

      // 使用 Ant Design Vue 的 Modal 组件进行确认
      Modal.confirm({
        title: '确认删除',
        icon: h(ExclamationCircleOutlined),
        content: `确定要删除软件包 "${pkg.name || '未知名称'}" 吗？此操作不可撤销。`,
        okText: '确定',
        cancelText: '取消',
        onOk() {
          packageActions.deletePackage(pkg.id)
        },
        onCancel() {
          console.log('取消删除软件包')
        }
      })
    }

    // 监听包列表变化，提供更好的用户反馈
    watch(() => packageActions.currentPackage.value, (newPackage, oldPackage) => {
      if (newPackage && !oldPackage) {
        message.success('软件包添加成功')
      }
    })

    // 创建一个新对象，包含本地状态和方法
    const result = {
      // 本地状态和方法
      columnWidths,
      resizingColumn,
      startResize: startColumnResize,
      handleResize: handleColumnResize,
      stopResize: stopColumnResize,
      handleModalOpenChange,
      confirmDeletePackage,
      useColumnWidths,
      useResizingColumn
    }

    // 添加 packageActions 中的所有属性和方法
    Object.assign(result, packageActions)

    // 确保返回的对象被使用
    console.log('result:', result)

    return result
  }
}
</script>
