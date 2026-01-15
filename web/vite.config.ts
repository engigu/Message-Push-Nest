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
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    define: {
      __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
      'globalThis.__BUILD_TIME__': JSON.stringify(new Date().toISOString()),
    },
  }
})

