<script>
import { ref, computed } from 'vue'
import { usePackageStore } from '@/stores/package'

export default {
  name: 'PackageListFilter',
  props: {
    headerRef: {
      type: Object,
      required: true
    }
  },
  setup(props) {
    const packageStore = usePackageStore()
    const currentStore = packageStore

    // 排序配置
    const sortBy = ref('name')
    const sortOrder = ref('asc')

    // 搜索相关
    const searchText = computed(() => {
      const text = props.headerRef.value?.searchText || ''
      return text
    })
    const checkerFilter = computed(() => {
      const filter = props.headerRef.value?.checkerFilter || ''
      return filter
    })
    const statusFilter = computed(() => {
      const filter = props.headerRef.value?.statusFilter || ''
      return filter
    })

    // 获取过滤后的软件包列表
    const filterByChecker = (packages, checker) => {
      if (!checker) return packages
      return packages.filter(pkg => pkg.upstreamChecker === checker)
    }

    const filterByStatus = (packages, status) => {
      if (!status) return packages
      if (status === 'needUpdate') {
        return packages.filter(pkg =>
          pkg.aurVersion &&
          pkg.upstreamVersion &&
          pkg.aurVersion !== pkg.upstreamVersion
        )
      } else if (status === 'unchecked') {
        return packages.filter(pkg =>
          (pkg.aurUpdateState ?? 0) === 0 ||
          (pkg.upstreamUpdateState ?? 0) === 0
        )
      } else if (status === 'failed') {
        return packages.filter(pkg => {
          const aurState = pkg.aurUpdateState ?? 0
          const upstreamState = pkg.upstreamUpdateState ?? 0
          return aurState === 2 || upstreamState === 2
        })
      }
      return packages
    }

    const filterBySearchText = (packages, text) => {
      if (!text) return packages
      const searchLower = text.toLowerCase()
      return packages.filter(pkg =>
        pkg.name.toLowerCase().includes(searchLower) ||
        pkg.aurName.toLowerCase().includes(searchLower) ||
        (pkg.upstreamProject && pkg.upstreamProject.toLowerCase().includes(searchLower))
      )
    }

    const sortPackages = (packages, sortField, sortDirection) => {
      return [...packages].sort((a, b) => {
        let compareValue = 0
        if (sortField === 'name') {
          compareValue = a.name.localeCompare(b.name)
        } else if (sortField === 'aurVersion') {
          compareValue = a.aurVersion?.localeCompare(b.aurVersion) || 0
        } else if (sortField === 'upstreamVersion') {
          compareValue = a.upstreamVersion?.localeCompare(b.upstreamVersion) || 0
        }
        return sortDirection === 'asc' ? compareValue : -compareValue
      })
    }

    const filteredPackages = computed(() => {
      let result = currentStore.packages

      // 应用检查器筛选
      if (checkerFilter.value) {
        result = result.filter(pkg => pkg.upstreamChecker === checkerFilter.value)
      }

      // 应用状态筛选
      if (statusFilter.value) {
        if (statusFilter.value === 'needUpdate') {
          // 需要更新：AUR版本和上游版本都存在但不相同
          result = result.filter(pkg =>
            pkg.aurVersion &&
            pkg.upstreamVersion &&
            pkg.aurVersion !== pkg.upstreamVersion
          )
        } else if (statusFilter.value === 'unchecked') {
          // 未检查：检查状态为未检查(0)
          result = result.filter(pkg =>
            (pkg.aurUpdateState ?? 0) === 0 ||
            (pkg.upstreamUpdateState ?? 0) === 0
          )
        } else if (statusFilter.value === 'failed') {
          // 检查失败：检查状态为失败
          result = result.filter(pkg => {
            // 如果 AUR 检查失败或上游检查失败，则显示
            const aurState = pkg.aurUpdateState ?? 0
            const upstreamState = pkg.upstreamUpdateState ?? 0
            return aurState === 2 || upstreamState === 2
          })
        }
      }

      // 应用搜索文本筛选
      if (searchText.value) {
        const searchLower = searchText.value.toLowerCase()
        result = result.filter(pkg =>
          pkg.name.toLowerCase().includes(searchLower) ||
          (pkg.aurName && pkg.aurName.toLowerCase().includes(searchLower)) ||
          (pkg.upstreamProject && pkg.upstreamProject.toLowerCase().includes(searchLower))
        )
      }

      // 应用排序
      result = [...result].sort((a, b) => {
        let compareValue = 0

        if (sortBy.value === 'name') {
          compareValue = a.name.localeCompare(b.name)
        } else if (sortBy.value === 'aurVersion') {
          compareValue = a.aurVersion?.localeCompare(b.aurVersion) || 0
        } else if (sortBy.value === 'upstreamVersion') {
          compareValue = a.upstreamVersion?.localeCompare(b.upstreamVersion) || 0
        }

        return sortOrder.value === 'asc' ? compareValue : -compareValue
      })

      return result
    })

    // 排序切换
    const toggleSort = (field) => {
      if (sortBy.value === field) {
        sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
      } else {
        // 如果按新字段排序，则设置新字段并默认升序
        sortBy.value = field
        sortOrder.value = 'asc'
      }
    }

    // 确保这些变量被使用
    console.log('sortBy:', sortBy.value)
    console.log('sortOrder:', sortOrder.value)
    console.log('filteredPackages:', filteredPackages.value)
    console.log('toggleSort:', toggleSort)
    
    // 创建返回对象
    const result = {
      sortBy, // 在 index.vue 中用于排序
      sortOrder, // 在 index.vue 中用于排序
      filteredPackages, // 在 index.vue 中用于显示过滤后的软件包列表
      toggleSort // 在 index.vue 中用于切换排序
    }
    
    // 确保返回的对象被使用
    console.log('result:', result)
    
    return result
  }
}
</script>
