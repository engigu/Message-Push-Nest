<template>
  <div class="flex items-center justify-between">
    <div class="text-sm text-muted-foreground">
      <span class="block sm:inline">共 {{ total }} 条记录</span>
      <span class="hidden sm:inline"> · </span>
      <span class="block sm:inline">每页 {{ pageSize }} 条</span>
    </div>
    
    <div class="flex justify-end">
      <Pagination
        :total="total"
        :items-per-page="pageSize"
        :sibling-count="1"
        :show-edges="true"
        :default-page="currentPage"
        :page="currentPage"
        @update:page="handlePageChange"
        class="justify-end"
      >
      <PaginationContent>
        <PaginationItem :value="currentPage - 1">
          <PaginationPrevious :disabled="currentPage <= 1" class="mr-3" @click.prevent.stop="handlePageChange(currentPage - 1)">
            上一页
          </PaginationPrevious>
        </PaginationItem>

        <template v-for="(page, index) in items" :key="index">
          <PaginationItem v-if="page.type === 'page'" :value="page.value" :is-active="page.value === currentPage">
            <button
              :class="[
                'h-9 w-9 rounded-md border transition-colors',
                page.value === currentPage
                  ? 'bg-primary text-primary-foreground border-primary'
                  : 'border-input bg-background hover:bg-accent hover:text-accent-foreground'
              ]"
              @click="handlePageChange(page.value)"
            >
              {{ page.value }}
            </button>
          </PaginationItem>
          <PaginationEllipsis v-else-if="page.type === 'ellipsis'" :index="index" />
        </template>

        <PaginationItem :value="currentPage + 1">
          <PaginationNext :disabled="currentPage >= totalPages" class="ml-3" @click.prevent.stop="handlePageChange(currentPage + 1)">
            下一页
          </PaginationNext>
        </PaginationItem>
      </PaginationContent>
      </Pagination>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'

interface Props {
  total: number
  currentPage: number
  pageSize: number
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'page-change': [page: number]
}>()

const totalPages = computed(() => {
  return Math.ceil(props.total / props.pageSize)
})

// 分页项目类型
type PaginationItem = 
  | { type: 'page'; value: number }
  | { type: 'ellipsis' }

// 生成分页项目
const items = computed((): PaginationItem[] => {
  const pages: PaginationItem[] = []
  const total = totalPages.value
  const current = props.currentPage
  
  if (total <= 7) {
    // 如果总页数小于等于7，显示所有页码
    for (let i = 1; i <= total; i++) {
      pages.push({ type: 'page', value: i })
    }
  } else {
    // 复杂分页逻辑
    if (current <= 2) {
      // 当前页在前面
      for (let i = 1; i <= 3; i++) {
        pages.push({ type: 'page', value: i })
      }
      pages.push({ type: 'ellipsis' })
      pages.push({ type: 'page', value: total })
    } else if (current >= total - 3) {
      // 当前页在后面
      pages.push({ type: 'page', value: 1 })
      pages.push({ type: 'ellipsis' })
      for (let i = total - 4; i <= total; i++) {
        pages.push({ type: 'page', value: i })
      }
    } else {
      // 当前页在中间
      pages.push({ type: 'page', value: 1 })
      pages.push({ type: 'ellipsis' })
      for (let i = current - 1; i <= current + 1; i++) {
        pages.push({ type: 'page', value: i })
      }
      pages.push({ type: 'ellipsis' })
      pages.push({ type: 'page', value: total })
    }
  }
  
  return pages
})

const handlePageChange = (page: number) => {
  if (page !== props.currentPage && page >= 1 && page <= totalPages.value) {
    emit('page-change', page)
  }
}

// no-op: previous/next handled inline to avoid double updates
</script>

<script lang="ts">
export default {
  name: 'Pagination'
}
</script>