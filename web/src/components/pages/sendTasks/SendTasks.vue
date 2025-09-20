<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import AddTasks from './AddTasks.vue'
import EditTasks from './EditTasks.vue'
import ApiCodeViewer from './ApiCodeViewer.vue'
import { toast } from 'vue-sonner'

import { useRoute, useRouter } from 'vue-router';
import { request } from '@/api/api';
import { CONSTANT } from '@/constant';
// @ts-ignore
import { getPageSize } from '@/util/pageUtils';


interface WayItem {
  id: string
  name: string
  type: string
  config: string
  created_on: string
  modified_on: string
  status: number
}

const route = useRoute();
const router = useRouter();
let state = reactive({
  tableData: [] as WayItem[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize(),
  search: '',
  optionValue: '',
})

// 状态过滤
const selectedStatus = ref('all')

// 任务类型过滤
const selectedChannelType = ref('all')

// Sheet 相关状态
const isSheetOpen = ref(false)
const selectedConfig = ref('')
const selectedChannelName = ref('')

// 新增任务 Sheet 状态
const isAddChannelDrawerOpen = ref(false)

// 编辑任务 Sheet 状态
const isEditChannelDrawerOpen = ref(false)
const editChannelData = ref<WayItem | null>(null)

// API代码查看器状态
const isApiCodeViewerOpen = ref(false)
const selectedTaskData = ref<Record<string, any> | undefined>(undefined)

// 处理保存新任务
const handleSaveChannel = (_data: any) => {
  // 这里可以添加实际的保存逻辑
  // 保存成功后刷新列表
  queryListDataWithStatus()
}

// 总页数
const totalPages = computed(() => Math.ceil(state.total / state.pageSize))

// const getWayTypeText = (type: string) => {
//   const wayItem = CONSTANT.WAYS_DATA.find(item => item.type === type)
//   return wayItem ? wayItem.label : type
// }

// 打开编辑任务Drawer
const openEditChannelDrawer = (channel: WayItem) => {
  editChannelData.value = channel
  isEditChannelDrawerOpen.value = true
}

// 处理编辑任务保存
const handleEditChannel = (_data: any) => {
  // 保存成功后刷新列表
  queryListDataWithStatus()
}

// 处理查看日志
const handleViewLogs = (channel: WayItem) => {
  // 跳转到发信日志页面，携带taskid参数
  router.push(`/sendlogs?taskid=${channel.id}`)
}

// 处理查看API接口
const handleViewApi = (channel: WayItem) => {
  selectedTaskData.value = channel
  isApiCodeViewerOpen.value = true
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

// 按任务类型过滤
// const filterByChannelType = async (value: any) => {
//   if (value) {
//     selectedChannelType.value = String(value);
//     state.currPage = 1; // 重置到第一页
//     await queryListDataWithStatus();
//   }
// }

// 查询数据（包含状态过滤）
const queryListDataWithStatus = async () => {
  const statusParam = selectedStatus.value === 'all' ? '' : selectedStatus.value;
  const channelTypeParam = selectedChannelType.value === 'all' ? '' : selectedChannelType.value;
  await queryListData(state.currPage, state.pageSize, state.search, channelTypeParam, '', statusParam);
}

const queryListData = async (page: number, size: number, name = '', channelType = '', query = '', status = '') => {
  let params: any = { page: page, size: size, name: name, type: channelType, query: query };
  if (status !== '') {
    params.status = status;
  }
  const rsp = await request.get('/sendtasks/list', { params: params });
  state.tableData = await rsp.data.data.lists;
  state.total = await rsp.data.data.total;
}
// 删除任务
const handleDelete = async (id: string) => {
  const rsp = await request.post('/sendtasks/delete', { id: id });
  if (rsp.status == 200 && await rsp.data.code == 200) {
    // state.tableData.splice(index, 1);
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
    route.query.channel_type?.toString() || ''
  );
});

</script>

<template>
  <div class="p-4 w-full max-w-6xl mx-auto space-y-2">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 sm:gap-4">
      <div class="flex flex-col sm:flex-row gap-2 sm:gap-4">
        <div class="flex-1 sm:flex-initial">
          <Input v-model="state.search" placeholder="搜索发信方式名称..." class="w-full sm:w-64" @keyup.enter="filterFunc"
            @blur="filterFunc" />
        </div>
      </div>

      <div class="flex-shrink-0">
        <Dialog v-model:open="isAddChannelDrawerOpen">
          <DialogTrigger as-child>
            <Button variant="default" class="w-full sm:w-auto bg-gray-800 hover:bg-gray-900">
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              新增任务
            </Button>
          </DialogTrigger>

          <DialogContent class="w-[500px] max-w-[90vw]">
            <DialogHeader>
              <DialogTitle>新增发信任务</DialogTitle>
            </DialogHeader>

            <div class="px-4 pb-4">
              <AddTasks v-model:open="isAddChannelDrawerOpen" @save="handleSaveChannel"
                @cancel="() => isAddChannelDrawerOpen = false" />
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
          <TableHead>发信任务名称</TableHead>
          <TableHead>创建时间</TableHead>
          <TableHead>更新时间</TableHead>
          <TableHead class="text-center">操作/状态</TableHead>
        </TableRow>
      </TableHeader>

      <TableBody>
        <!-- 空数据展示 -->
        <TableRow v-if="state.tableData.length === 0">
          <TableCell colspan="6" class="text-center py-12">
            <EmptyTableState title="暂无发信方式" description="还没有配置任何发信方式，请先添加发信方式" />
          </TableCell>
        </TableRow>

        <!-- 数据行 -->
        <TableRow v-for="channel in state.tableData" :key="channel.id">
          <TableCell>{{ channel.id }}</TableCell>
          <TableCell>{{ channel.name }}</TableCell>
          <TableCell>{{ channel.created_on }}</TableCell>
          <TableCell>{{ channel.modified_on }}</TableCell>
          <TableCell class="text-center space-x-2" v-if="channel.id !== CONSTANT.LOG_TASK_ID">
            <Button size="sm" variant="outline" @click="handleViewApi(channel)">接口</Button>
            <Button size="sm" variant="outline" @click="handleViewLogs(channel)">日志</Button>
            <Button size="sm" variant="outline" @click="openEditChannelDrawer(channel)">编辑</Button>
            <Button size="sm" variant="outline" class="text-red-500 border-red-300 hover:bg-red-50 
              hover:border-red-400 hover:text-red-600 hover:shadow-md
               transition-all duration-200" @click="handleDelete(channel.id)">删除</Button>
          </TableCell>
          <TableCell class="text-center space-x-2" v-else>
            <Button size="sm" variant="outline" @click="handleViewLogs(channel)">日志</Button>
            <label>系统保留任务</label>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <!-- 分页 -->
    <Pagination :total="state.total" :current-page="state.currPage" :page-size="state.pageSize"
      @page-change="changePage" />

    <!-- 编辑任务Dialog -->
    <Dialog v-model:open="isEditChannelDrawerOpen">
      <DialogContent class="w-[500px] max-w-[90vw] max-h-[90vh] overflow-hidden flex flex-col">
        <DialogHeader class="flex-shrink-0">
          <DialogTitle>编辑发信任务</DialogTitle>
        </DialogHeader>

        <div class="px-4 pb-4 flex-1 overflow-y-auto">
          <EditTasks v-model:open="isEditChannelDrawerOpen" :edit-data="editChannelData" @save="handleEditChannel" />
        </div>
      </DialogContent>
    </Dialog>

    <!-- 配置详情Sheet -->
    <Sheet v-model:open="isSheetOpen">
      <SheetContent class="w-[600px] sm:w-[900px] lg:w-[1000px]">
        <SheetHeader>
          <SheetTitle>{{ selectedChannelName }} - 发信方式配置详情</SheetTitle>
        </SheetHeader>
        <div class="mt-6">
          <div class="bg-gray-50 p-4 rounded-lg">
            <pre class="whitespace-pre-wrap text-sm font-mono">{{ selectedConfig }}</pre>
          </div>
        </div>
      </SheetContent>
    </Sheet>

    <!-- API代码查看器 -->
    <ApiCodeViewer v-model:open="isApiCodeViewerOpen" :task-data="selectedTaskData" />
  </div>
</template>