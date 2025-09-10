<template>
  <div class="package-list-container">
    <a-card class="package-list-card" :body-style="{ padding: 0, height: '100%', display: 'flex', flexDirection: 'column' }">
      <!-- 固定操作区域 -->
      <div class="fixed-header">
        <div class="header-content">
          <PackageListHeader ref="headerRef" @showAddModal="showAddModal" />
        </div>
      </div>

      <!-- 可滚动区域 -->
      <PackageListDisplay
        :packages="filteredPackages"
        :column-widths="columnWidths"
        :sort-by="sortBy"
        :sort-order="sortOrder"
        :checking-aur-for-package="checkingAurForPackage"
        :checking-upstream-for-package="checkingUpstreamForPackage"
        @toggle-sort="toggleSort"
        @startResize="startResize"
        @checkPackageAurVersion="checkPackageAurVersion"
        @checkPackageUpstreamVersion="checkPackageUpstreamVersion"
        @checkPackageAllVersions="checkPackageAllVersions"
        @showEditModal="showEditModal"
        @confirmDeletePackage="confirmDeletePackage"
      />
    </a-card>

    <!-- 添加/编辑软件包模态框 -->
    <PackageModal
      :open="modalVisible"
      :title="modalTitle"
      :isEdit="isEdit"
      :packageData="currentPackage"
      @update:open="handleModalOpenChange"
      @ok="handleModalOk"
    />
  </div>
</template>

<script>
import { ref, onMounted, watch } from 'vue'
import { usePackageStore } from '@/stores/package'
import { message } from 'ant-design-vue'
import PackageListHeader from './PackageListHeader.vue'
import PackageListDisplay from './PackageListDisplay.vue'
import PackageModal from "./PackageModal.vue"
import PackageListFilter from './PackageListFilter.vue'
import PackageListActions from './PackageListActions.vue'

export default {
  name: 'PackageList',
  components: {
    PackageListHeader,
    PackageListDisplay,
    PackageModal,
    PackageListFilter,
    PackageListActions
  },
  setup() {
    const packageStore = usePackageStore()
    const headerRef = ref(null)

    // 生命周期
    onMounted(async () => {
      try {
        // 初始化时强制加载数据
        await packageStore.fetchPackages()
      } catch (error) {
        console.error('加载软件包列表失败:', error)
        message.error('加载软件包列表失败，请刷新页面重试')
      }
    })

    // 监听包列表变化，提供更好的用户反馈
    watch(() => packageStore.packages.length, (newLength, oldLength) => {
      if (newLength > oldLength) {
        const addedCount = newLength - oldLength
        message.success(`成功添加了 ${addedCount} 个软件包`)
      }
    })

    // 使用过滤器逻辑
    const packageFilter = PackageListFilter.setup({ headerRef })
    const { sortBy, sortOrder, filteredPackages, toggleSort } = packageFilter

    // 使用操作逻辑
    const packageActions = PackageListActions.setup()
    const {
      columnWidths,
      currentPackage,
      checkingAurForPackage,
      checkingUpstreamForPackage,
      modalVisible,
      modalTitle,
      isEdit,
      showAddModal,
      showEditModal,
      handleModalOk,
      handleModalOpenChange,
      checkPackageAurVersion,
      checkPackageUpstreamVersion,
      checkPackageAllVersions,
      startResize,
      handleResize,
      stopResize
    } = packageActions

    // 确认删除软件包
    const confirmDeletePackage = (pkg) => {
      // 使用已经解构出来的方法
      packageActions.confirmDeletePackage(pkg)
    }

    // 以下变量都在模板中被使用
    return {
      headerRef, // 在模板中用于 PackageListHeader
      columnWidths, // 在模板中用于 PackageListDisplay
      sortBy, // 在模板中用于 PackageListDisplay
      sortOrder, // 在模板中用于 PackageListDisplay
      filteredPackages, // 在模板中用于 PackageListDisplay
      currentPackage, // 在模板中用于 PackageModal
      checkingAurForPackage, // 在模板中用于 PackageListDisplay
      checkingUpstreamForPackage, // 在模板中用于 PackageListDisplay
      modalVisible, // 在模板中用于 PackageModal
      modalTitle, // 在模板中用于 PackageModal
      isEdit, // 在模板中用于 PackageModal
      showAddModal, // 在模板中用于 PackageListHeader
      showEditModal, // 在模板中用于 PackageListDisplay
      handleModalOk, // 在模板中用于 PackageModal
      handleModalOpenChange, // 在模板中用于 PackageModal
      checkPackageAurVersion, // 在模板中用于 PackageListDisplay
      checkPackageUpstreamVersion, // 在模板中用于 PackageListDisplay
      checkPackageAllVersions, // 在模板中用于 PackageListDisplay
      toggleSort, // 在模板中用于 PackageListDisplay
      startResize, // 在模板中用于 PackageListDisplay
      handleResize, // 在模板中用于 PackageListDisplay
      stopResize, // 在模板中用于 PackageListDisplay
      confirmDeletePackage // 在模板中用于 PackageListDisplay
    }
  }
}
</script>

<style scoped>
.package-list-container {
  height: 100%;
  overflow: hidden !important;
}

.package-list-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.fixed-header {
  position: sticky;
  top: 0;
  z-index: 10;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  padding: 12px 0;
}

.header-content {
  padding: 0 16px;
}
</style>
