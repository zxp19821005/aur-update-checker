
<template>
  <div class="virtual-scroll-container" ref="containerRef" @scroll="handleScroll">
    <div class="virtual-scroll-content">
      <div class="virtual-scroll-spacer" :style="{ height: totalHeight + 'px' }"></div>
      <div class="virtual-scroll-items" :style="{ transform: `translateY(${offsetY}px)` }">
        <div 
          v-for="(item, index) in visibleItems" 
          :key="item.id || index" 
          class="virtual-scroll-item" 
          :style="{ height: `${itemHeight}px` }"
        >
          <slot :item="item" :index="startIndex + index"></slot>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'

// 属性定义
const props = defineProps({
  items: {
    type: Array,
    default: () => []
  },
  itemHeight: {
    type: Number,
    default: 100
  },
  buffer: {
    type: Number,
    default: 5 // 预渲染的额外项目数量
  }
})

// 引用
const containerRef = ref(null)
const scrollTop = ref(0)
const containerHeight = ref(0)

// 计算属性
const totalHeight = computed(() => props.items.length * props.itemHeight)
const startIndex = computed(() => {
  if (!containerHeight.value || props.itemHeight <= 0) return 0
  return Math.max(0, Math.floor(scrollTop.value / props.itemHeight) - props.buffer)
})
const endIndex = computed(() => {
  if (!containerHeight.value || props.itemHeight <= 0 || props.items.length === 0) return 0
  const visibleCount = Math.ceil(containerHeight.value / props.itemHeight)
  return Math.min(props.items.length - 1, startIndex.value + visibleCount + props.buffer * 2)
})
const visibleItems = computed(() => {
  if (props.items.length === 0) return []
  const start = startIndex.value
  const end = Math.min(props.items.length, endIndex.value + 1)
  return props.items.slice(start, end)
})
const offsetY = computed(() => {
  return startIndex.value * props.itemHeight
})

// 方法
const handleScroll = () => {
  if (containerRef.value) {
    scrollTop.value = containerRef.value.scrollTop
  }
}

const updateContainerHeight = () => {
  if (containerRef.value) {
    containerHeight.value = containerRef.value.clientHeight
  }
}

// 生命周期
onMounted(() => {
  updateContainerHeight()
  window.addEventListener('resize', updateContainerHeight)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateContainerHeight)
})

// 监听容器引用变化
watch(containerRef, () => {
  updateContainerHeight()
})
</script>

<style scoped>
.virtual-scroll-container {
  height: 100%;
  overflow-y: auto !important;
  position: relative;
  width: 100%;
  box-sizing: border-box;
  scrollbar-width: thin; /* Firefox */
  /* 确保滚动条可见 */
  margin-right: 0;
  padding-right: 0;
}

/* 显示滚动条 */
.virtual-scroll-container::-webkit-scrollbar {
  width: 8px !important;
  display: block !important;
}

.virtual-scroll-container::-webkit-scrollbar-track {
  background: #f1f1f1 !important;
}

.virtual-scroll-container::-webkit-scrollbar-thumb {
  background: #888 !important;
  border-radius: 4px;
}

.virtual-scroll-container::-webkit-scrollbar-thumb:hover {
  background: #555 !important;
}

.virtual-scroll-content {
  position: relative;
  width: 100%;
  /* 确保内容区域不超出容器 */
  box-sizing: border-box;
  overflow: hidden;
}

.virtual-scroll-spacer {
  width: 1px;
  float: left;
}

.virtual-scroll-items {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  /* 确保项目区域不超出容器 */
  box-sizing: border-box;
  overflow: hidden;
}

.virtual-scroll-item {
  width: 100%;
  box-sizing: border-box;
  overflow: hidden;
}
</style>
