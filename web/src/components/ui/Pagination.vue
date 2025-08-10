<template>
  <div class="flex items-center justify-between space-y-6">
    <div class="text-sm text-gray-500">
      共 {{ total }} 条记录，第 {{ currentPage }} / {{ totalPages }} 页
    </div>
    <div class="flex items-center space-x-2">
      <Button size="sm" variant="outline" :disabled="currentPage <= 1" @click="$emit('page-change', currentPage - 1)">
        上一页
      </Button>
      <Button size="sm" variant="outline" :disabled="currentPage >= totalPages"
        @click="$emit('page-change', currentPage + 1)">
        下一页
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@/components/ui/button'

interface Props {
  total: number
  currentPage: number
  pageSize: number
}

const props = defineProps<Props>()

// const _emit = defineEmits<{
//   'page-change': [page: number]
// }>()

const totalPages = computed(() => {
  return Math.ceil(props.total / props.pageSize)
})
</script>

<script lang="ts">
export default {
  name: 'Pagination'
}
</script>