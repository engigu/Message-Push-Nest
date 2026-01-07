<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import ClickableTruncate from '@/components/ui/ClickableTruncate.vue'

import { useRoute } from 'vue-router';
import { request } from '@/api/api';
// @ts-ignore
import { getPageSize } from '@/util/pageUtils';


interface HostedMessageItem {
  id: number
  title: string
  content: string
  url?: string
  created_on: string
  modified_on: string
  status: number
}

const router = useRoute();

let state = reactive({
  tableData: [] as HostedMessageItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  optionValue: '',
})

// 状态过滤
// const selectedStatus = ref('all')
// Sheet 相关状态
const isSheetOpen = ref(false)
const selectedLog = ref('')
const selectedTaskName = ref('')
// 总页数
const totalPages = computed(() => Math.ceil(state.total / state.pageSize))


// 打开消息详情Sheet
const openMessageSheet = (message: HostedMessageItem) => {
  selectedLog.value = formatMessageDisplayHtml(message);
  selectedTaskName.value = message.title
  isSheetOpen.value = true
}

const changePage = async (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    state.currPage = page
    await queryListData(
      state.currPage,
      state.pageSize,
      state.search,
      state.optionValue
    )
  }
}

// 格式化处理显示的消息文本
const formatMessageDisplayHtml = (message: HostedMessageItem) => {
  let content = `标题: ${message.title}\n\n内容: ${message.content}`;
  if (message.url) {
    content += `\n\nURL: ${message.url}`;
  }
  content += `\n\n创建时间: ${message.created_on}`;
  content += `\n修改时间: ${message.modified_on}`;
  return content;
}

//触发过滤筛选
const filterFunc = async () => {
  await queryListData(state.currPage, state.pageSize, state.search, state.optionValue);
}
  

const queryListData = async (page: number, size: number, text = '', query = '') => {
  let params: any = { page: page, size: size, text: text, query: query };
  const rsp = await request.get('/hostedmessages/list', { params: params });
  state.tableData = await rsp.data.data.lists || [];
  state.total = await rsp.data.data.total || 0;
}

onMounted(async () => {
  // 页面加载触发消息显示
  state.search = router.query.name?.toString() || '';
  await queryListData(
    1,
    state.pageSize,
    router.query.name?.toString() || '',
    router.query.query?.toString() || ''
  );
});
</script>

<template>
  <div class="p-4 w-full max-w-6xl mx-auto space-y-2">
    <div class="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4">
      <div class="flex-1 sm:flex-initial">
        <Input v-model="state.search" placeholder="搜索消息..." class="w-full sm:w-64" @keyup.enter="filterFunc"
          @blur="filterFunc" />
      </div>
    </div>

    <!-- 表格 -->
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead class="w-20">ID</TableHead>
          <TableHead>消息标题</TableHead>
          <TableHead>消息内容</TableHead>
          <TableHead class="whitespace-nowrap w-[160px]">创建时间</TableHead>
          <TableHead class="text-center">详情</TableHead>
        </TableRow>
      </TableHeader>

      <TableBody>
        <!-- 空数据展示 -->
        <TableRow v-if="state.tableData.length === 0">
          <TableCell colspan="5" class="text-center py-12">
            <EmptyTableState 
              title="暂无托管消息" 
              description="还没有任何托管消息记录" 
            >
              <template #icon>
                <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
                </svg>
              </template>
            </EmptyTableState>
          </TableCell>
        </TableRow>
        
        <!-- 数据行 -->
        <TableRow v-for="message in state.tableData" :key="message.id">
          <TableCell>{{ message.id }}</TableCell>
          <TableCell>
            <ClickableTruncate :text="message.title" wrapper-class="max-w-[180px] sm:max-w-[240px]" preview-title="消息标题" />
          </TableCell>
          <TableCell>
            <ClickableTruncate :text="message.content" wrapper-class="max-w-[320px] sm:max-w-[480px]" preview-title="消息内容" />
          </TableCell>
          <TableCell class="whitespace-nowrap w-[160px]">{{ message.created_on }}</TableCell>
          <TableCell class="text-center space-x-2">
            <Button size="sm" variant="outline" @click="openMessageSheet(message)">查看</Button>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <!-- 分页 -->
    <Pagination 
      :total="state.total" 
      :current-page="state.currPage" 
      :page-size="state.pageSize" 
      @page-change="changePage" 
    />

    <!-- 消息详情Sheet -->
    <Sheet v-model:open="isSheetOpen">
      <SheetContent class="lg:w-[800px] ">
        <SheetHeader>
          <SheetTitle>{{ selectedTaskName }} - 托管消息详情</SheetTitle>
        </SheetHeader>
        <div class="mt-4">
          <div class="bg-card p-4 rounded-lg border border-border max-h-[82vh] overflow-y-auto break-words">
            <pre class="whitespace-pre-wrap text-sm font-mono text-foreground">{{ selectedLog }}</pre>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>