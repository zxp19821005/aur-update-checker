import { defineStore } from 'pinia'
// 不再导入 message，使用全局的 window.$message
import { getPackageList, getPackageById, addPackage, updatePackage, deletePackage } from '@/services/api'

export const usePackageStore = defineStore('package', {
  state: () => ({
    packages: [],
    loading: false,
    viewMode: 'list', // 'list' 或 'card'
    selectedPackage: null
  }),

  getters: {
    // 获取需要更新的软件包数量
    outdatedPackagesCount: (state) => {
      return state.packages.filter(pkg => {
        if (!pkg.aurVersion || !pkg.upstreamVersion) return false
        return pkg.aurVersion !== pkg.upstreamVersion
      }).length
    },

    // 获取检查失败的软件包数量
    failedPackagesCount: (state) => {
      return state.packages.filter(pkg => {
        return pkg.aurUpdateState === 2 || pkg.upstreamUpdateState === 2
      }).length
    },

    // 获取未检查的软件包数量
    uncheckedPackagesCount: (state) => {
      return state.packages.filter(pkg => {
        return pkg.aurUpdateState === 0 || pkg.upstreamUpdateState === 0
      }).length
    }
  },

  actions: {
    // 获取所有软件包
    async fetchPackages() {
      this.loading = true
      try {
        const response = await getPackageList()
        this.packages = response || []
      } catch (error) {
        window.$message.error('获取软件包列表失败')
        console.error('获取软件包列表失败:', error)
        this.packages = [] // 确保在错误情况下packages也是一个空数组
      } finally {
        this.loading = false
      }
    },

    // 根据ID获取软件包
    async fetchPackageById(id) {
      this.loading = true
      try {
        const response = await getPackageById(id)
        this.selectedPackage = response
        return response
      } catch (error) {
        window.$message.error('获取软件包详情失败')
        console.error('获取软件包详情失败:', error)
        return null
      } finally {
        this.loading = false
      }
    },

    // 添加软件包
    async addPackage(packageData) {
      this.loading = true
      try {
        // 验证输入数据
        if (!packageData) {
          throw new Error('软件包数据不能为空')
        }
        if (!packageData.name) {
          throw new Error('软件包名称不能为空')
        }
        if (!packageData.upstreamUrl) {
          throw new Error('上游URL不能为空')
        }
        
        console.log('准备添加软件包:', packageData)
        
        const response = await addPackage(packageData)
        
        // 验证响应数据
        if (!response) {
          throw new Error('添加软件包失败: 未收到响应数据')
        }
        
        this.packages.push(response)
        window.$message.success('添加软件包成功')
        return response
      } catch (error) {
        // 根据错误类型显示不同的错误信息
        let errorMessage = '添加软件包失败'
        
        if (error.message) {
          if (error.message.includes('已存在同名软件包')) {
            errorMessage = `已存在同名软件包: ${packageData.name}`
          } else if (error.message.includes('软件包名称不能为空')) {
            errorMessage = '软件包名称不能为空'
          } else if (error.message.includes('上游URL不能为空')) {
            errorMessage = '上游URL不能为空'
          } else if (error.message.includes('未收到响应数据')) {
            errorMessage = '添加软件包失败: 未收到响应数据'
          } else {
            errorMessage = `添加软件包失败: ${error.message}`
          }
        }
        
        window.$message.error(errorMessage)
        console.error('添加软件包失败:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    // 更新软件包
    async updatePackage(id, packageData) {
      this.loading = true
      try {
        const response = await updatePackage(id, packageData)
        console.log('更新软件包响应数据:', response)
        const index = this.packages.findIndex(pkg => pkg.id === id)
        if (index !== -1) {
          this.packages[index] = response
          console.log('更新后的 packages[index]:', this.packages[index])
        }
        window.$message.success('更新软件包成功')
        return response
      } catch (error) {
        window.$message.error('更新软件包失败')
        console.error('更新软件包失败:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    // 删除软件包
    async deletePackage(id) {
      this.loading = true
      try {
        await deletePackage(id)
        this.packages = this.packages.filter(pkg => pkg.id !== id)
        window.$message.success('删除软件包成功')
      } catch (error) {
        window.$message.error('删除软件包失败')
        console.error('删除软件包失败:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    // 设置视图模式
    setViewMode(mode) {
      this.viewMode = mode
    },

    // 更新软件包状态
    updatePackageStatus(packageId, type, status) {
      const index = this.packages.findIndex(pkg => pkg.id === packageId)
      if (index !== -1) {
        if (type === 'aur') {
          this.packages[index].aurUpdateState = status
        } else if (type === 'upstream') {
          this.packages[index].upstreamUpdateState = status
        }
      }
    }
  }
})
