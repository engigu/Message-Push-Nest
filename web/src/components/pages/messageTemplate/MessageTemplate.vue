<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import Pagination from '@/components/ui/Pagination.vue'
import ClickableTruncate from '@/components/ui/ClickableTruncate.vue'
import TemplateApiViewer from './TemplateApiViewer.vue'
import TemplateInstanceConfig from './TemplateInstanceConfig.vue'
import TemplateEditor from './TemplateEditor.vue'
import { request } from '@/api/api'
import { getPageSize } from '@/util/pageUtils'
import { toast } from 'vue-sonner'
import { useRouter } from 'vue-router'

interface MessageTemplate {
  id: string  // 模板ID是字符串类型（UUID）
  name: string
  description: string
  text_template: string
  html_template: string
  markdown_template: string
  placeholders: string
  at_mobiles?: string
  at_user_ids?: string
  is_at_all?: boolean
  status: string
  created_on: string
  modified_on: string
}

const router = useRouter()

let state = reactive({
  tableData: [] as MessageTemplate[],
  total: 0,
  currPage: 1,
  pageSize: getPageSize() as number,
  search: '',
  status: 'all'
})

// API代码查看器状态
const isApiViewerOpen = ref(false)
const selectedTemplateForApi = ref<MessageTemplate | null>(null)

// 配置实例状态
const isInstanceConfigOpen = ref(false)
const selectedTemplateForInstance = ref<MessageTemplate | null>(null)

// 模板编辑器状态
const isEditorOpen = ref(false)
const isEditing = ref(false)
const selectedTemplateForEdit = ref<MessageTemplate | null>(null)

const totalPages = computed(() => Math.ceil(state.total / state.pageSize))

const queryListData = async (page: number, size: number, text = '', status = '') => {
  const params: any = { page, size, text, status }
  const rsp = await request.get('/templates/list', { params })
  state.tableData = rsp.data.data.lists || []
  state.total = rsp.data.data.total || 0
}

const changePage = async (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    state.currPage = page
    const statusParam = state.status === 'all' ? '' : state.status
    await queryListData(state.currPage, state.pageSize, state.search, statusParam)
  }
}

const filterFunc = async () => {
  state.currPage = 1
  const statusParam = state.status === 'all' ? '' : state.status
  await queryListData(state.currPage, state.pageSize, state.search, statusParam)
}

const openAddDialog = () => {
  isEditing.value = false
  selectedTemplateForEdit.value = null
  isEditorOpen.value = true
}

const openEditDialog = (template: MessageTemplate) => {
  isEditing.value = true
  selectedTemplateForEdit.value = template
  isEditorOpen.value = true
}

const handleEditorSaved = async () => {
  // 刷新列表
  const statusParam = state.status === 'all' ? '' : state.status
  await queryListData(state.currPage, state.pageSize, state.search, statusParam)
}

const deleteTemplate = async (id: string) => {
  const rsp = await request.post('/templates/delete', { id })
  if (rsp.status === 200 && rsp.data.code === 200) {
    toast.success(rsp.data.msg)
    // 刷新列表，处理status参数
    const statusParam = state.status === 'all' ? '' : state.status
    await queryListData(state.currPage, state.pageSize, state.search, statusParam)
  }
}

// 打开API查看器
const handleViewApi = (template: MessageTemplate) => {
  selectedTemplateForApi.value = template
  isApiViewerOpen.value = true
}

// 打开配置实例
const handleConfigInstance = (template: MessageTemplate) => {
  selectedTemplateForInstance.value = template
  isInstanceConfigOpen.value = true
}

// 查看日志
const handleViewLogs = (template: MessageTemplate) => {
  // 跳转到发信日志页面，携带 taskid 参数（传递模板 id）
  router.push(`/sendlogs?taskid=${template.id}`)
}

onMounted(async () => {
  await queryListData(1, state.pageSize)
})
</script>

<template>
  <div class="p-4 w-full max-w-6xl mx-auto space-y-2">
    <!-- 搜索和筛选 -->
    <div class="flex flex-row sm:flex-row sm:items-center gap-2 -mx-2 px-2 sm:mx-0 sm:px-0">
      <div class="flex-[3] sm:flex-initial min-w-0">
        <Input
          v-model="state.search"
          placeholder="搜索模板名称或描述..."
          class="w-full sm:w-64"
          @keyup.enter="filterFunc"
          @blur="filterFunc"
        />
      </div>
      
      <div class="flex-[2] sm:flex-initial min-w-0">
        <Select v-model="state.status" class="w-full" @update:model-value="filterFunc">
          <SelectTrigger class="w-full">
            <SelectValue placeholder="选择状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">全部</SelectItem>
              <SelectItem value="enabled">启用</SelectItem>
              <SelectItem value="disabled">禁用</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>
      
      <div class="flex-1"></div>
      
      <Button @click="openAddDialog">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        新建模板
      </Button>
    </div>

    <!-- 表格 -->
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead class="w-24">ID</TableHead>
          <TableHead>模板名称</TableHead>
          <TableHead>描述</TableHead>
          <TableHead>支持格式</TableHead>
          <TableHead>状态</TableHead>
          <TableHead class="whitespace-nowrap w-[160px]">创建时间</TableHead>
          <TableHead class="text-center">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <!-- 空数据展示 -->
        <TableRow v-if="state.tableData.length === 0">
          <TableCell colspan="7" class="text-center py-12">
            <EmptyTableState 
              title="暂无消息模板" 
              description="还没有创建任何消息模板，点击右上角按钮创建新模板" 
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
        <TableRow v-for="item in state.tableData" :key="item.id">
          <TableCell>{{ item.id }}</TableCell>
          <TableCell>
            <ClickableTruncate :text="item.name" wrapper-class="max-w-[80px] sm:max-w-[100px]" preview-title="模板名称" />
          </TableCell>
          <TableCell>
            <ClickableTruncate :text="item.description || '-'" wrapper-class="max-w-[80px] sm:max-w-[130px]" preview-title="模板描述" />
          </TableCell>
          <TableCell>
            <div class="flex gap-1">
              <Badge v-if="item.text_template" variant="secondary">Text</Badge>
              <Badge v-if="item.html_template" variant="secondary">HTML</Badge>
              <Badge v-if="item.markdown_template" variant="secondary">Markdown</Badge>
            </div>
          </TableCell>
          <TableCell>
            <Badge :variant="item.status === 'enabled' ? 'default' : 'secondary'">
              {{ item.status === 'enabled' ? '启用' : '禁用' }}
            </Badge>
          </TableCell>
          <TableCell class="whitespace-nowrap w-[160px]">{{ item.created_on }}</TableCell>
          <TableCell class="text-center space-x-2">
            <Button size="sm" variant="outline" @click="handleViewLogs(item)">日志</Button>
            <Button size="sm" variant="outline" @click="handleViewApi(item)">接口</Button>
            <Button size="sm" variant="outline" @click="openEditDialog(item)">编辑</Button>
            <Button size="sm" variant="outline" @click="handleConfigInstance(item)">实例</Button>
            <Button size="sm" variant="destructive" @click="deleteTemplate(item.id)">删除</Button>
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

    <!-- 模板编辑器 -->
    <TemplateEditor
      :open="isEditorOpen"
      :is-editing="isEditing"
      :template-data="selectedTemplateForEdit"
      @update:open="isEditorOpen = $event"
      @saved="handleEditorSaved"
    />

    <!-- API代码查看器 -->
    <TemplateApiViewer 
      :open="isApiViewerOpen" 
      :template-data="selectedTemplateForApi || undefined"
      @update:open="isApiViewerOpen = $event"
    />

    <!-- 配置实例 -->
    <TemplateInstanceConfig 
      :open="isInstanceConfigOpen" 
      :template-data="selectedTemplateForInstance"
      @update:open="isInstanceConfigOpen = $event"
    />
  </div>
</template>
