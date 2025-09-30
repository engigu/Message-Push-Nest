<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import AddCronMessages from './AddCronMessages.vue'
import EditCronMessages from './EditCronMessages.vue'
import { toast } from 'vue-sonner'

import { useRoute, useRouter } from 'vue-router';
import { request } from '@/api/api';
// @ts-ignore
import { getPageSize } from '@/util/pageUtils';


interface CronMessageItem {
  id: string
  title: string
  name: string
  content: string
  cron: string
  cron_expression: string
  task_id: string
  task_name: string
  enable: number
  status: boolean
  created_on: string
  modified_on: string
  next_time?: string
  url: string
}

const route = useRoute();
const router = useRouter();

let state = reactive({
  tableData: [] as CronMessageItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  optionValue: '',
})

// 状态过滤
const selectedStatus = ref('all')


// 新增定时消息 Dialog 状态
const isAddCronMessageDialogOpen = ref(false)

// 编辑定时消息 Dialog 状态
const isEditCronMessageDialogOpen = ref(false)
const editCronMessageData = ref<CronMessageItem | null>(null)

// 处理保存新定时消息
const handleSaveCronMessage = (_data: any) => {
  // 保存成功后刷新列表
  queryListDataWithStatus()
}

// 总页数
const totalPages = computed(() => Math.ceil(state.total / state.pageSize))

// 打开编辑定时消息Dialog
const openEditCronMessageDialog = (cronMessage: CronMessageItem) => {
  editCronMessageData.value = cronMessage
  isEditCronMessageDialogOpen.value = true
}

// 处理编辑定时消息保存
const handleEditCronMessage = (_data: any) => {
  // 保存成功后刷新列表
  queryListDataWithStatus()
}

// 处理查看日志
const handleViewLogs = (cronMessage: CronMessageItem) => {
  // 跳转到定时消息日志页面，携带cronMessageId参数
  router.push(`/sendlogs?taskid=${cronMessage.task_id}`)
}

// 切换状态
const toggleStatus = async (cronMessage: CronMessageItem) => {
  const newStatus = !cronMessage.enable ? 1 : 0
  cronMessage.enable = newStatus
  const rsp = await request.post('/cronmessages/edit', cronMessage)
  if (rsp.data.code === 200) {
    cronMessage.enable = newStatus
    toast.success(rsp.data.msg)
  }
}

const changePage = async (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    state.currPage = page
    await queryListDataWithStatus()
  }
}

//触发过滤筛选
const filterFunc = async () => {
  await queryListData(state.currPage, state.pageSize, state.search, state.optionValue);
}

// 查询数据（包含状态过滤）
const queryListDataWithStatus = async () => {
  const statusParam = selectedStatus.value === 'all' ? '' : selectedStatus.value;
  await queryListData(state.currPage, state.pageSize, state.search, '', '', statusParam);
}

const queryListData = async (page: number, size: number, name = '', taskType = '', query = '', status = '') => {
  let params: any = { page: page, size: size, name: name, type: taskType, query: query };
  if (status !== '') {
    params.status = status;
  }
  const rsp = await request.get('/cronmessages/list', { params: params });
  state.tableData = await rsp.data.data.lists;
  state.total = await rsp.data.data.total;
}

// 删除定时消息
const handleDelete = async (id: string) => {
  const rsp = await request.post('/cronmessages/delete', { id: id });
  if (rsp.status == 200 && await rsp.data.code == 200) {
    toast.success(rsp.data.msg);
    setTimeout(() => {
      window.location.reload();
    }, 1000);
  }
}


onMounted(async () => {
  // 初始化查询
  state.search = route.query.name?.toString() || '';
  await queryListData(
    1,
    state.pageSize,
    route.query.name?.toString() || '',
    route.query.task_type?.toString() || ''
  );
});

</script>

<template>
  <div class="p-4 w-full max-w-6xl mx-auto space-y-2">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 sm:gap-4">
      <div class="flex flex-col sm:flex-row gap-2 sm:gap-4">
        <div class="flex-1 sm:flex-initial">
          <Input v-model="state.search" placeholder="搜索定时任务名称..." class="w-full sm:w-64" @keyup.enter="filterFunc"
            @blur="filterFunc" />
        </div>
      </div>

      <div class="flex-shrink-0">
        <Dialog v-model:open="isAddCronMessageDialogOpen">
          <DialogTrigger as-child>
            <Button variant="default" class="w-full sm:w-auto">
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              新增定时消息
            </Button>
          </DialogTrigger>

          <DialogContent class="w-[500px] max-w-[90vw]">
            <DialogHeader>
              <DialogTitle>新增定时消息</DialogTitle>
            </DialogHeader>

            <div class="px-4 pb-4">
              <AddCronMessages v-model:open="isAddCronMessageDialogOpen" @save="handleSaveCronMessage"
                @cancel="() => isAddCronMessageDialogOpen = false" />
            </div>
          </DialogContent>
        </Dialog>
      </div>
    </div>

    <!-- 表格 -->
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead class="w-20">ID</TableHead>
          <TableHead>名称</TableHead>
          <TableHead>内容</TableHead>
          <TableHead>Cron表达式</TableHead>
          <TableHead>关联任务</TableHead>
          <TableHead>下次执行时间</TableHead>
          <TableHead>创建时间</TableHead>
          <TableHead class="text-center">操作</TableHead>
        </TableRow>
      </TableHeader>

      <TableBody>
        <!-- 空数据展示 -->
        <TableRow v-if="state.tableData.length === 0">
          <TableCell colspan="8" class="text-center py-12">
            <EmptyTableState title="暂无定时消息" description="还没有配置任何定时消息，请先添加定时消息" />
          </TableCell>
        </TableRow>

        <!-- 数据行 -->
        <TableRow v-for="cronMessage in state.tableData" :key="cronMessage.id">
          <TableCell>{{ cronMessage.id }}</TableCell>
          <TableCell class="max-w-32 truncate" :title="cronMessage.title">{{ cronMessage.title }}</TableCell>
          <TableCell class="max-w-32 truncate" :title="cronMessage.content">{{ cronMessage.content }}</TableCell>
          <TableCell>
            <code class="px-2 py-1 rounded text-sm font-mono bg-muted text-foreground border border-border">
              {{ cronMessage.cron }}
            </code>
          </TableCell>
          <TableCell class="max-w-24 truncate" :title="cronMessage.task_id">{{ cronMessage.task_id }}</TableCell>
          <TableCell class="max-w-32 truncate" :title="cronMessage.next_time || '-'">{{ cronMessage.next_time || '-' }}
          </TableCell>
          <TableCell class="max-w-32 truncate" :title="cronMessage.created_on">{{ cronMessage.created_on }}</TableCell>
          <!-- <TableCell class="text-center">
            <Badge :class="cronMessage.status === 1 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-600'">
              {{ getStatusText(cronMessage.status) }}
            </Badge>
          </TableCell> -->
          <TableCell class="text-center space-x-2">
            <Button size="sm" variant="outline" @click="handleViewLogs(cronMessage)">日志</Button>
            <Button size="sm" variant="outline" @click="openEditCronMessageDialog(cronMessage)">编辑</Button>
            <Button size="sm" variant="outline" class="text-red-500 border-red-300 hover:bg-red-50 
              hover:border-red-400 hover:text-red-600 hover:shadow-md
               transition-all duration-200" @click="handleDelete(cronMessage.id)">删除</Button>
            <Switch :model-value="cronMessage.enable === 1" @update:model-value="toggleStatus(cronMessage)" />

          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <!-- 分页 -->
    <Pagination :total="state.total" :current-page="state.currPage" :page-size="state.pageSize"
      @page-change="changePage" />

    <!-- 编辑定时消息Dialog -->
    <Dialog v-model:open="isEditCronMessageDialogOpen">
      <DialogContent class="w-[500px] max-w-[90vw] max-h-[90vh] overflow-hidden flex flex-col">
        <DialogHeader class="flex-shrink-0">
          <DialogTitle>编辑定时消息</DialogTitle>
        </DialogHeader>

        <div class="px-4 pb-4 flex-1 overflow-y-auto">
          <EditCronMessages v-model:open="isEditCronMessageDialogOpen" :cron-message="editCronMessageData"
            @save="handleEditCronMessage" />
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>