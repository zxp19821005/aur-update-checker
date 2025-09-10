<template>
  <div class="list-view-container">
    <!-- 固定列表头部 -->
    <div class="list-header" ref="listHeaderRef">
      <div class="header-cell name-column sortable" :style="{ width: name + 'px', fontWeight: 'bold', textAlign: 'center' }" @click="toggleSort('name')">
        软件包名称
        <span class="sort-icon" v-if="sortBy === 'name'">{{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
        <div class="resize-handle" @mousedown="startResize('name', $event)"></div>
      </div>
      <div class="header-cell versions-column" :style="{ width: versions + 'px', fontWeight: 'bold', textAlign: 'center' }">
        版本信息
        <div class="resize-handle" @mousedown="startResize('versions', $event)"></div>
      </div>
      <div class="header-cell actions-column" :style="{ width: actions + 'px', fontWeight: 'bold', textAlign: 'center' }">
        操作
      </div>
    </div>
    <VirtualScroll
        :items="packages"
        :item-height="50"
        :buffer="20"
        class="virtual-scroll-list"
      >
        <template #default="{ item }">
          <div class="package-row"
               :class="{ 'outdated-row': item.aurVersion && item.upstreamVersion && item.aurVersion !== item.upstreamVersion, 'failed-row': item.aurUpdateState === 2 || item.upstreamUpdateState === 2 }"
               @mouseenter="handleRowHover(item)">
            <div class="package-name" :style="{ width: name + 'px' }">
              <span class="package-name-text">{{ item.name }}</span>
              <a-tag v-if="item.aurName && item.aurName !== item.name" color="blue" class="aur-name-tag">
                AUR: {{ item.aurName }}
              </a-tag>
            </div>
            <div class="package-versions" :style="{ width: versions + 'px' }">
              <a-tag :color="getVersionTagColor(item.aurUpdateState)" class="version-tag">
                {{ item.aurVersion || '-' }}
              </a-tag>
              <a-tag :color="getVersionTagColor(item.upstreamUpdateState)" class="version-tag">
                {{ item.upstreamVersion || '-' }}
              </a-tag>

            </div>
            <div class="package-actions" :style="{ width: actions + 'px' }">
              <a-button type="text" size="small" @click="checkPackageAurVersion(item.id)" :disabled="checkingAurForPackage === item.id">
                <a-spin v-if="checkingAurForPackage === item.id" size="small" />
                <sync-outlined v-else />
              </a-button>
              <a-button type="text" size="small" @click="checkPackageUpstreamVersion(item.id)" :disabled="checkingUpstreamForPackage === item.id">
                <a-spin v-if="checkingUpstreamForPackage === item.id" size="small" />
                <cloud-download-outlined v-else />
              </a-button>
              <a-button type="text" size="small" @click="checkPackageAllVersions(item.id)" :disabled="checkingAurForPackage === item.id || checkingUpstreamForPackage === item.id">
                <a-spin v-if="checkingAurForPackage === item.id && checkingUpstreamForPackage === item.id" size="small" />
                <reload-outlined v-else />
              </a-button>
              <a-button type="text" size="small" @click="showEditModal(item)">
                <edit-outlined />
              </a-button>
              <a-button type="text" size="small" danger @click="confirmDeletePackage(item)">
                <delete-outlined />
              </a-button>
            </div>
          </div>
        </template>
      </VirtualScroll>
  </div>
</template>

<script>
import { ref, toRefs, onMounted, h } from 'vue'
import VirtualScroll from "@/components/VirtualScroll.vue"
import { SyncOutlined, CloudDownloadOutlined, ReloadOutlined, EditOutlined, DeleteOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'

export default {
  name: 'PackageListDisplay',
  components: {
    VirtualScroll,
    SyncOutlined,
    CloudDownloadOutlined,
    ReloadOutlined,
    EditOutlined,
    DeleteOutlined,
    ExclamationCircleOutlined
  },
  props: {
    packages: {
      type: Array,
      required: true
    },
    columnWidths: {
      type: Object,
      required: true
    },
    sortBy: {
      type: String,
      required: true
    },
    sortOrder: {
      type: String,
      required: true
    },
    checkingAurForPackage: {
      type: [Number, String],
      default: null
    },
    checkingUpstreamForPackage: {
      type: [Number, String],
      default: null
    }
  },
  emits: ['toggleSort', 'startResize', 'checkPackageAurVersion', 'checkPackageUpstreamVersion', 'checkPackageAllVersions', 'showEditModal', 'confirmDeletePackage'],
  setup(props, { emit }) {
    const listHeaderRef = ref(null)

    // 从props中解构出列宽
    const columnWidthsRefs = toRefs(props.columnWidths)
    const { name, versions, actions } = columnWidthsRefs
    
    // 确保这些变量被使用
    console.log('name:', name.value)
    console.log('versions:', versions.value)
    console.log('actions:', actions.value)

    // 切换排序
    const toggleSort = (field) => {
      emit('toggleSort', field)
    }

    // 开始调整列宽
    const startResize = (column, event) => {
      emit('startResize', { column, event })
    }

    // 获取版本标签颜色
    const getVersionTagColor = (updateState) => {
      // 处理数字状态值
      const state = typeof updateState === 'number' ? updateState : parseInt(updateState) || 0

      // 0:未检查, 1:成功, 2:失败
      switch (state) {
        case 1: // 成功
          return 'green'
        case 2: // 失败
          return 'red'
        case 0: // 未检查
        default:
          return 'default'
      }
    }

    // 处理行悬停效果
    const handleRowHover = (item) => {
      // 可以在这里添加悬停效果，如预览包信息等
    }



    // 检查单个软件包的AUR版本
    const checkPackageAurVersion = (id) => {
      emit('checkPackageAurVersion', id)
    }

    // 检查单个软件包的上游版本
    const checkPackageUpstreamVersion = (id) => {
      emit('checkPackageUpstreamVersion', id)
    }

    // 检查单个软件包的所有版本
    const checkPackageAllVersions = (id) => {
      emit('checkPackageAllVersions', id)
    }

    // 显示编辑模态框
    const showEditModal = (pkg) => {
      emit('showEditModal', pkg)
    }

    // 确认删除软件包
    const confirmDeletePackage = (pkg) => {
      emit('confirmDeletePackage', pkg)
    }

    // 创建返回对象
    const result = {
      listHeaderRef, // 在组件中使用
      name, // 在模板中用于列宽
      versions, // 在模板中用于列宽
      actions, // 在模板中用于列宽
      toggleSort, // 在模板中用于排序
      startResize, // 在模板中用于调整列宽
      getVersionTagColor, // 在模板中用于版本标签颜色
      handleRowHover, // 在模板中用于行悬停效果
      checkPackageAurVersion, // 在模板中用于检查AUR版本
      checkPackageUpstreamVersion, // 在模板中用于检查上游版本
      checkPackageAllVersions, // 在模板中用于检查所有版本
      showEditModal, // 在模板中用于显示编辑模态框
      confirmDeletePackage // 在模板中用于确认删除软件包
    }
    
    // 确保返回的对象被使用
    console.log('result:', result)
    
    return result
  }
}
</script>

<style scoped>
.list-view-container {
  flex: 1;
  overflow: hidden;
  position: relative;
  height: calc(100vh - 120px);
}

.list-header {
  position: sticky;
  top: 0;
  z-index: 5;
  background: #fafafa;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
}

.header-cell {
  padding: 12px 8px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}

.name-column {
  flex-shrink: 0;
  justify-content: space-between;
}

.versions-column {
  flex-shrink: 0;
}

.actions-column {
  flex-shrink: 0;
}

.sort-icon {
  margin-left: 4px;
  font-size: 12px;
}

.resize-handle {
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  cursor: col-resize;
  background: transparent;
}

.resize-handle:hover {
  background: #1890ff;
}

.virtual-scroll-list {
  height: calc(100% - 50px);
  overflow-y: auto;
  overflow-x: hidden;
}

.package-row {
  display: flex;
  border-bottom: 1px solid #f0f0f0;
  height: 50px;
  align-items: center;
}

.package-name {
  padding: 0 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 0;
}

.package-versions {
  padding: 0 8px;
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.package-actions {
  padding: 0 8px;
  display: flex;
  gap: 4px;
  justify-content: flex-end;
  flex-shrink: 0;
}

/* 悬停效果 */
.package-row:hover {
  background-color: #f5f5f5;
  transition: background-color 0.3s;
}

/* 过期行样式 */
.outdated-row {
  background-color: #fff1f0;
}

.outdated-row:hover {
  background-color: #ffe7e6;
}

/* 失败行样式 */
.failed-row {
  background-color: #fff2e8;
}

.failed-row:hover {
  background-color: #ffe7d6;
}

/* 包名称样式 */
.package-name-text {
  font-weight: 500;
}

.aur-name-tag {
  margin-left: 8px;
  font-size: 10px;
}

.version-tag {
  min-width: 120px;
  text-align: center;
  display: inline-flex;
  justify-content: center;
}

/* 优化滚动条样式 */
.virtual-scroll-list::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.virtual-scroll-list::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.virtual-scroll-list::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.virtual-scroll-list::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}
</style>
