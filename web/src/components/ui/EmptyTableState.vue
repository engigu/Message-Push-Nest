<template>
  <div class="flex flex-col items-center justify-center space-y-3">
    <div class="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center">
      <slot name="icon">
        <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2M4 13h2m13-8V4a1 1 0 00-1-1H6a1 1 0 00-1 1v1m16 0V4a1 1 0 00-1-1H6a1 1 0 00-1 1v1"></path>
        </svg>
      </slot>
    </div>
    <div class="text-gray-500 text-center">
      <p class="text-lg font-medium">{{ title }}</p>
      <p class="text-sm">{{ description }}</p>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'EmptyTableState',
  props: {
    title: {
      type: String,
      default: '暂无数据'
    },
    description: {
      type: String,
      default: '当前没有任何数据'
    }
  }
})
</script>