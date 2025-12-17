<template>
  <div class="flex items-center justify-between gap-2 w-full">
    <!-- 统计信息 -->
    <div class="text-xs sm:text-sm text-muted-foreground shrink-0">
      <span>共 {{ total }} 条</span>
      <span class="hidden sm:inline"> · 每页 {{ pageSize }} 条</span>
    </div>
    
    <!-- 分页控件 -->
    <div class="flex justify-end shrink-0">
      <Pagination
        :total="total"
        :items-per-page="pageSize"
        :sibling-count="1"
        :show-edges="true"
        :default-page="currentPage"
        :page="currentPage"
        @update:page="handlePageChange"
      >
      <PaginationContent class="gap-0.5 sm:gap-1">
        <PaginationItem :value="currentPage - 1">
          <PaginationPrevious 
            :disabled="currentPage <= 1" 
            class="h-8 w-8 sm:h-9 sm:w-auto sm:px-3 text-xs sm:text-sm p-0 sm:p-2" 
            @click.prevent.stop="handlePageChange(currentPage - 1)"
          >
            <span class="hidden sm:inline">上一页</span>
            <span class="sm:hidden">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
              </svg>
            </span>
          </PaginationPrevious>
        </PaginationItem>

        <template v-for="(page, index) in displayItems" :key="index">
          <PaginationItem v-if="page.type === 'page'" :value="page.value" :is-active="page.value === currentPage">
            <button
              :class="[
                'h-8 w-8 sm:h-9 sm:w-9 rounded-md border transition-colors text-xs sm:text-sm',
                page.value === currentPage
                  ? 'bg-primary text-primary-foreground border-primary'
                  : 'border-input bg-background hover:bg-accent hover:text-accent-foreground'
              ]"
              @click="handlePageChange(page.value)"
            >
              {{ page.value }}
            </button>
          </PaginationItem>
          <PaginationEllipsis v-else-if="page.type === 'ellipsis'" :index="index" class="h-8 w-5 sm:h-9 sm:w-9" />
        </template>

        <PaginationItem :value="currentPage + 1">
          <PaginationNext 
            :disabled="currentPage >= totalPages" 
            class="h-8 w-8 sm:h-9 sm:w-auto sm:px-3 text-xs sm:text-sm p-0 sm:p-2" 
            @click.prevent.stop="handlePageChange(currentPage + 1)"
          >
            <span class="hidden sm:inline">下一页</span>
            <span class="sm:hidden">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </span>
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

// 响应式显示项目（根据屏幕大小使用不同逻辑）
const displayItems = computed((): PaginationItem[] => {
  const pages: PaginationItem[] = []
  const total = totalPages.value
  const current = props.currentPage
  
  // 使用媒体查询判断是否为小屏
  const isSmallScreen = typeof window !== 'undefined' && window.innerWidth < 640
  
  if (isSmallScreen) {
    // 小屏逻辑：只显示关键页码
    if (total <= 5) {
      // 总页数少，显示所有
      for (let i = 1; i <= total; i++) {
        pages.push({ type: 'page', value: i })
      }
    } else {
      // 只显示当前页和首尾页
      if (current === 1) {
        pages.push({ type: 'page', value: 1 })
        pages.push({ type: 'page', value: 2 })
        pages.push({ type: 'ellipsis' })
        pages.push({ type: 'page', value: total })
      } else if (current === total) {
        pages.push({ type: 'page', value: 1 })
        pages.push({ type: 'ellipsis' })
        pages.push({ type: 'page', value: total - 1 })
        pages.push({ type: 'page', value: total })
      } else {
        pages.push({ type: 'page', value: 1 })
        if (current > 2) {
          pages.push({ type: 'ellipsis' })
        }
        pages.push({ type: 'page', value: current })
        if (current < total - 1) {
          pages.push({ type: 'ellipsis' })
        }
        pages.push({ type: 'page', value: total })
      }
    }
  } else {
    // 大屏逻辑：显示更多页码
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