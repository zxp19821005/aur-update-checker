
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    // 启用代码分割
    rollupOptions: {
      output: {
        manualChunks: {
          // 将第三方库分离到单独的chunk
          'vendor': ['vue', 'pinia', 'ant-design-vue', '@ant-design/icons-vue'],
          // 将路由相关代码分离到单独的chunk
          'router': ['./src/router/index.js'],
          // 将API相关代码分离到单独的chunk
          'api': ['./src/services/api.js', './src/services/cachedApi.js', './src/services/cache.js'],
          // 将store相关代码分离到单独的chunk
          'store': ['./src/stores/package.js', './src/stores/timer.js']
        },
        // 优化chunk文件名
        chunkFileNames: (chunkInfo) => {
          const facadeModuleId = chunkInfo.facadeModuleId
            ? chunkInfo.facadeModuleId.split('/').pop()
            : 'chunk'
          return `js/${facadeModuleId}-[hash].js`
        },
        // 优化资源文件名
        assetFileNames: (assetInfo) => {
          const info = assetInfo.name.split('.')
          const ext = info[info.length - 1]
          if (/\.(mp4|webm|ogg|mp3|wav|flac|aac)(\?.*)?$/i.test(assetInfo.name)) {
            return `media/[name]-[hash].${ext}`
          }
          if (/\.(png|jpe?g|gif|svg)(\?.*)?$/i.test(assetInfo.name)) {
            return `img/[name]-[hash].${ext}`
          }
          if (/\.(woff2?|eot|ttf|otf)(\?.*)?$/i.test(assetInfo.name)) {
            return `fonts/[name]-[hash].${ext}`
          }
          if (/\.css$/.test(assetInfo.name)) {
            return `css/[name]-[hash].css`
          }
          return `assets/[name]-[hash].${ext}`
        }
      }
    },
    // 启用源码映射
    sourcemap: true,
    // 压缩选项
    minify: 'terser',
    terserOptions: {
      compress: {
        // 删除console
        drop_console: true,
        // 删除debugger
        drop_debugger: true
      }
    }
  },
  // 优化依赖预构建
  optimizeDeps: {
    include: [
      'vue',
      'pinia',
      'ant-design-vue',
      '@ant-design/icons-vue',
      'axios',
      'dayjs'
    ]
  },
  // CSS优化
  css: {
    // 启用CSS代码分割
    codeSplit: true,
    // 预处理器选项
    preprocessorOptions: {
      less: {
        javascriptEnabled: true,
        modifyVars: {
          // 主题变量
          'primary-color': '#1890ff'
        }
      }
    }
  }
})

