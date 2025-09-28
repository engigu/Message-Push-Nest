<script setup lang="ts">
import { ref, computed, reactive, onMounted, nextTick, watch } from 'vue'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import Pagination from '@/components/ui/Pagination.vue'

import { useRoute } from 'vue-router';
import { request } from '@/api/api';
// @ts-ignore
import { getPageSize } from '@/util/pageUtils';


interface LogItem {
  id: number
  task_id: number
  task_name: string
  log: string
  created_on: string
  caller_ip?: string
  status: number
}

const router = useRoute();

let state = reactive({
  tableData: [] as LogItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  optionValue: '',
})

// 状态过滤
const selectedStatus = ref('all')
// Sheet 相关状态
const isSheetOpen = ref(false)
const selectedLog = ref('')
const selectedTaskName = ref('')
// 总页数
const totalPages = computed(() => Math.ceil(state.total / state.pageSize))

const getStatusText = (status: number) => {
  return status === 1 ? '成功' : '失败'
}

// 打开日志详情Sheet
const openLogSheet = (task: LogItem) => {
  selectedLog.value = formatLogDisplayHtml(task);
  selectedTaskName.value = task.task_name
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

// 格式化处理显示的日志文本
const formatLogDisplayHtml = (task: LogItem) => {
  let log = task.log;
  log += '\n';
  if (task.caller_ip) {
    log += `调用来源IP：${task.caller_ip}`;
  };
  return log;
}

//触发过滤筛选
const filterFunc = async () => {
  await queryListData(state.currPage, state.pageSize, state.search, state.optionValue);
}

// 按状态过滤
const filterByStatus = async (value: any) => {
  selectedStatus.value = value;
  state.currPage = 1; // 重置到第一页
  await queryListData(state.currPage, state.pageSize, state.search, state.optionValue);
}

const queryListData = async (page: number, size: number, name = '', taskid = '', query = '', _status = '') => {
  let params: any = { page: page, size: size, name: name, taskid: taskid };
  
  // 优先使用URL传入的query参数（包含日期筛选等）
  if (query) {
    params.query = query;
  } else if (selectedStatus.value !== '' && selectedStatus.value !== 'all') {
    // 如果没有URL query参数，使用当前选择的状态筛选
    params.query = JSON.stringify({
      status: selectedStatus.value
    });
  }

  const rsp = await request.get('/sendlogs/list', { params: params });
  
  // 清空现有数据
  state.tableData = [];
  
  // 使用 nextTick 确保响应式更新
  await nextTick();
  
  // 更新数据
  state.tableData = rsp.data.data.lists || [];
  state.total = rsp.data.data.total;
}

// 解析URL参数并更新筛选状态
const parseUrlParams = async () => {
  state.search = router.query.name?.toString() || '';
  
  // 解析URL中的query参数，设置状态筛选
  const queryParam = router.query.query?.toString() || '';
  if (queryParam) {
    try {
      const queryObj = JSON.parse(decodeURIComponent(queryParam));
      if (queryObj.status !== undefined) {
        selectedStatus.value = queryObj.status.toString();
      }
    } catch (error) {
      console.warn('解析query参数失败:', error);
    }
  } else {
    // 如果没有query参数，重置为全部
    selectedStatus.value = 'all';
  }
  
  await queryListData(
    1,
    state.pageSize,
    router.query.name?.toString() || '',
    router.query.taskid?.toString() || '',
    queryParam
  );
};

// 监听路由变化
watch(() => router.query, () => {
  parseUrlParams();
}, { deep: true });

onMounted(async () => {
  await parseUrlParams();
});
</script>

<template>
  <div class="p-4 w-full max-w-6xl mx-auto space-y-2">
    <div class="flex flex-row sm:flex-row sm:items-center gap-2 sm:gap-4 -mx-4 px-4 sm:mx-0 sm:px-0">
      <div class="flex-[3] sm:flex-initial min-w-0">
        <Input v-model="state.search" placeholder="搜索任务..." class="w-full sm:w-64" @keyup.enter="filterFunc"
          @blur="filterFunc" />
      </div>
    
      <div class="flex-[2] sm:flex-initial min-w-0">
        <Select v-model="selectedStatus" class="w-full" @update:model-value="filterByStatus">
          <SelectTrigger class="w-full">
            <SelectValue placeholder="选择状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">全部</SelectItem>
              <SelectItem value="1">成功</SelectItem>
              <SelectItem value="0">失败</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>
    </div>

    <!-- 表格 -->
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead class="w-20">ID</TableHead>
          <TableHead>任务名称</TableHead>
          <TableHead>发信日志</TableHead>
          <TableHead>发送时间</TableHead>
          <TableHead class="text-center">详情/状态</TableHead>
        </TableRow>
      </TableHeader>

      <TableBody>
        <!-- 空数据展示 -->
        <TableRow v-if="state.tableData.length === 0">
          <TableCell colspan="5" class="text-center py-12">
            <EmptyTableState 
              title="暂无发信日志" 
              description="还没有任何发信日志记录" 
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
        <TableRow v-for="task in state.tableData" :key="task.id">
          <TableCell>{{ task.id }}</TableCell>
          <TableCell>{{ task.task_name }}</TableCell>
          <TableCell class="max-w-xs truncate" :title="formatLogDisplayHtml(task)">{{ task.log }}</TableCell>
          <TableCell>{{ task.created_on }}</TableCell>
          <TableCell class="text-center space-x-2">
            <Button size="sm" variant="outline" @click="openLogSheet(task)">查看</Button>
            <!-- <Button size="sm" variant="destructive">删除</Button> -->
            <Badge :class="task.status === 1 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-600'">
              {{ getStatusText(task.status) }}
            </Badge>
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

    <!-- 日志详情Sheet -->
    <Sheet v-model:open="isSheetOpen" class="lg:w-[900px] ">
      <SheetContent class="lg:w-[900px]">
        <SheetHeader>
          <SheetTitle>{{ selectedTaskName }} - 发信日志详情</SheetTitle>
        </SheetHeader>
        <div class="mt-4">
          <div class="bg-gray-50 p-4 rounded-lg max-h-[82vh] overflow-y-auto break-words">
            <pre class="whitespace-pre-wrap text-sm font-mono">{{ selectedLog }}</pre>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>