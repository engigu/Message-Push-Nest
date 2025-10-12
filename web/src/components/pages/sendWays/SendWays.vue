<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Drawer, DrawerContent, DrawerHeader, DrawerTitle, DrawerTrigger } from '@/components/ui/drawer'
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import AddWays from './AddWays.vue'
import EditWays from './EditWays.vue'
import { toast } from 'vue-sonner'

import { useRoute } from 'vue-router';
import { request } from '@/api/api';
import { CONSTANT } from '@/constant';
// @ts-ignore
import { getPageSize } from '@/util/pageUtils';


interface WayItem {
  id: number
  name: string
  type: string
  config: string
  created_on: string
  modified_on: string
  status: number
}

const router = useRoute();

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

// 渠道类型过滤
const selectedChannelType = ref('all')

// 渠道类型选项 - 从CONSTANT.WAYS_DATA生成
const channelTypeOptions = computed(() => {
  const options = [{ value: 'all', label: '全部类型' }]
  CONSTANT.WAYS_DATA.forEach(item => {
    options.push({ value: item.type, label: item.label })
  })
  return options
})

// Sheet 相关状态
const isSheetOpen = ref(false)
const selectedConfig = ref('')
const selectedChannelName = ref('')

// 新增渠道 Sheet 状态
const isAddChannelDrawerOpen = ref(false)

// 编辑渠道 Sheet 状态
const isEditChannelDrawerOpen = ref(false)
const editChannelData = ref<WayItem | null>(null)

// 处理保存新渠道
const handleSaveChannel = (_data: any) => {
  // 这里可以添加实际的保存逻辑
  // 保存成功后刷新列表
  queryListDataWithStatus()
}

// 总页数
const totalPages = computed(() => Math.ceil(state.total / state.pageSize))

const getWayTypeText = (type: string) => {
  const wayData = CONSTANT.WAYS_DATA.find(item => item.type === type)
  return wayData ? wayData.label : type
}

// 打开编辑渠道Drawer
const openEditChannelDrawer = (channel: WayItem) => {
  editChannelData.value = channel
  isEditChannelDrawerOpen.value = true
}

// 处理编辑渠道保存
const handleEditChannel = (_data: any) => {
  // 保存成功后刷新列表
  queryListDataWithStatus()
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

// 按渠道类型过滤
const filterByChannelType = async (value: any) => {
  if (value) {
    selectedChannelType.value = String(value);
    state.currPage = 1; // 重置到第一页
    await queryListDataWithStatus();
  }
}

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
  const rsp = await request.get('/sendways/list', { params: params });
  state.tableData = await rsp.data.data.lists;
  state.total = await rsp.data.data.total;
}
// 删除渠道
const handleDelete = async (id: number) => {
  const rsp = await request.post('/sendways/delete', { id: id });
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
  state.search = router.query.name?.toString() || '';
  await queryListData(
    1,
    state.pageSize,
    router.query.name?.toString() || '',
    router.query.channel_type?.toString() || ''
  );
});
</script>

<template>
  <div class="p-4 w-full max-w-6xl mx-auto space-y-2">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 sm:gap-4">
      <div class="flex flex-row sm:flex-row gap-2 sm:gap-4 -mx-4 px-4 sm:mx-0 sm:px-0">
        <div class="flex-[3] sm:flex-initial min-w-0">
          <Input v-model="state.search" placeholder="搜索发信方式名称..." class="w-full sm:w-64" @keyup.enter="filterFunc"
            @blur="filterFunc" />
        </div>

        <div class="flex-[2] sm:flex-initial min-w-0">
          <Select v-model="selectedChannelType" class="w-full" @update:model-value="filterByChannelType">
            <SelectTrigger class="w-full">
              <SelectValue placeholder="选择渠道类型" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem v-for="option in channelTypeOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>
        
      </div>

      <div class="flex-shrink-0">
        <Drawer v-model:open="isAddChannelDrawerOpen">
          <DrawerTrigger as-child>
            <Button variant="default" class="w-full sm:w-auto">
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              新增渠道
            </Button>
          </DrawerTrigger>

          <DrawerContent class="w-[800px] max-w-[90vw] mx-auto h-[90vh] max-h-[90vh]">
            <DrawerHeader>
              <DrawerTitle>新增发信渠道</DrawerTitle>
            </DrawerHeader>

            <div class="px-4 pb-4 overflow-y-auto">
              <AddWays v-model:open="isAddChannelDrawerOpen" @save="handleSaveChannel" />
            </div>
          </DrawerContent>
        </Drawer>
      </div>
    </div>

    <!-- 表格 -->
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead class="w-20">ID</TableHead>
          <TableHead>发信方式名称</TableHead>
          <TableHead>发信方式类型</TableHead>
          <TableHead class="whitespace-nowrap w-[160px]">创建时间</TableHead>
          <TableHead class="whitespace-nowrap w-[160px]">更新时间</TableHead>
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
          <TableCell>
            <Badge variant="outline">{{ getWayTypeText(channel.type) }}</Badge>
          </TableCell>
          <TableCell class="whitespace-nowrap w-[160px]">{{ channel.created_on }}</TableCell>
          <TableCell class="whitespace-nowrap w-[160px]">{{ channel.modified_on }}</TableCell>
          <TableCell class="text-center space-x-2">
            <Button size="sm" variant="outline" @click="openEditChannelDrawer(channel)">编辑</Button>
            <!-- <Button size="sm" variant="outline" @click="openConfigSheet(channel)">查看</Button> -->
            <Button size="sm" variant="outline" class="text-red-500 border-red-300 hover:bg-red-50 
              hover:border-red-400 hover:text-red-600 hover:shadow-md
               transition-all duration-200" @click="handleDelete(channel.id)">删除</Button>
            <!-- <Badge :class="channel.status === 1 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-600'">
              {{ getStatusText(channel.status) }} -->
            <!-- </Badge> -->
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <!-- 分页 -->
    <Pagination :total="state.total" :current-page="state.currPage" :page-size="state.pageSize"
      @page-change="changePage" />

    <!-- 编辑渠道Drawer -->
    <Drawer v-model:open="isEditChannelDrawerOpen">
      <DrawerContent class="w-[800px] max-w-[90vw] mx-auto h-[90vh] max-h-[90vh]">
        <DrawerHeader>
          <DrawerTitle>编辑发信渠道</DrawerTitle>
        </DrawerHeader>

        <div class="px-4 pb-4 overflow-y-auto">
          <EditWays v-model:open="isEditChannelDrawerOpen" :edit-data="editChannelData" @save="handleEditChannel" />
        </div>
      </DrawerContent>
    </Drawer>

    <!-- 配置详情Sheet -->
    <Sheet v-model:open="isSheetOpen">
      <SheetContent class="w-[600px] sm:w-[900px] lg:w-[1000px]">
        <SheetHeader>
          <SheetTitle>{{ selectedChannelName }} - 发信方式配置详情</SheetTitle>
        </SheetHeader>
        <div class="mt-6">
          <div class="rounded-lg p-4 bg-muted/40 dark:bg-white/5 ring-1 ring-border/50 shadow-sm">
            <pre class="whitespace-pre-wrap text-sm font-mono leading-relaxed text-foreground">{{ selectedConfig }}</pre>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>