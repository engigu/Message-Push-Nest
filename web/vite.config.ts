// import { defineConfig } from 'vite'
// import vue from '@vitejs/plugin-vue'

// // https://vite.dev/config/
// export default defineConfig({
//   plugins: [vue()],
// })



import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

export default defineConfig(() => {
  return {
    // 使用相对路径，这样可以在任何路径下部署
    base: './',
    plugins: [vue(), tailwindcss()],
    server: {
      port: 5173,
      proxy: {
        '/api': {
          target: 'http://localhost:8000',
          changeOrigin: true,
          configure: (proxy) => {
            proxy.on('proxyReq', (proxyReq, req) => {
              const host = req.headers.host;
              if (host) {
                proxyReq.setHeader('X-Forwarded-Host', host);
                proxyReq.setHeader('X-Forwarded-Proto', (req.socket as any)?.encrypted ? 'https' : 'http');
              }
            });
          },
        },
      },
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    build: {
      chunkSizeWarningLimit: 1000,
      rollupOptions: {
        onwarn(warning, defaultWarn) {
          // 过滤 node_modules 三方库中无害的 PURE 注释位置警告
          if (warning.code === 'INVALID_ANNOTATION') return
          defaultWarn(warning)
        },
        output: {
          manualChunks(id) {
            if (id.includes('node_modules')) {
              if (id.includes('vue') || id.includes('pinia') || id.includes('vue-router')) {
                return 'vendor-core'
              }
              if (id.includes('lucide-vue-next') || id.includes('reka-ui')) {
                return 'vendor-ui'
              }
              if (id.includes('apexcharts') || id.includes('vue3-apexcharts')) {
                return 'vendor-charts'
              }
            }
          },
        },
      },
    },
    define: {
      __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
      'globalThis.__BUILD_TIME__': JSON.stringify(new Date().toISOString()),
    },
  }
})

