<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import EmptyTableState from '@/components/ui/EmptyTableState.vue'
import {
  Combobox,
  ComboboxAnchor,
  ComboboxInput,
  ComboboxList,
  ComboboxItem,
  ComboboxViewport
} from '@/components/ui/combobox'
import { CheckIcon } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Label } from '@/components/ui/label'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Switch } from '@/components/ui/switch'
import { CONSTANT } from '@/constant'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
import { generateBizUniqueID } from '@/util/uuid'

// 组件props
interface Props {
  open?: boolean
  templateData?: any // 模板数据
}
const props = withDefaults(defineProps<Props>(), {
  open: false,
  templateData: null
})

// 组件emits
defineEmits<{
  'update:open': [value: boolean]
}>()

// 前端的页面添加配置
const waysConfigMap = CONSTANT.WAYS_DATA

// 搜索相关状态
const searchQuery = ref('')
const isSearching = ref(false)
const channelName = ref('')

// 当前显示的选项（搜索结果或所有选项）
const displayOptions = ref<Array<{ id: string, name: string, type: string }>>([])

// 输入框显示值（只在用户搜索时显示搜索内容）
const inputDisplayValue = computed({
  get: () => searchQuery.value,
  set: (value: string) => {
    searchQuery.value = value
  }
})

// 当前选中渠道的配置
const currentChannelConfig = computed(() => {
  // 先根据label找到type
  let type = displayOptions.value.find(item => item.name === channelName.value)?.type
  // 再根据type找到配置
  let rs = waysConfigMap.find(item => item.type === type) || null;
  
  return rs;
})

// 表单数据
const formData = ref<Record<string, any>>({})

// 监听渠道变化
const handlechannelNameChange = () => {
  // 数据加载后，text/html单选设置默认选中（这里选第一个）
  if (currentChannelConfig.value?.taskInsRadios.length > 0) {
    formData.value.templ_type = currentChannelConfig.value?.taskInsRadios[0].subLabel
  }
}

// 添加单条实例配置
const handleAddSubmit = async () => {
  // 验证是否选择了渠道
  if (!channelName.value) {
    toast.error('请选择发送渠道')
    return
  }

  // 验证内容类型
  const contentType = formData.value.templ_type
  if (!contentType) {
    toast.error('请选择消息格式')
    return
  }

  // 验证模板对应格式的内容是否为空
  const templateFieldMap: Record<string, string> = {
    'text': 'text_template',
    'html': 'html_template',
    'markdown': 'markdown_template'
  }
  
  const fieldName = templateFieldMap[contentType.toLowerCase()]
  if (fieldName) {
    const templateContent = props.templateData?.[fieldName] || ''
    // 检查是否为空（去除所有空白字符后检查）
    if (!templateContent.trim()) {
      toast.error(`模板的 ${contentType} 格式内容为空，无法添加此类型的实例`)
      return
    }
  }

  // 组建表单数据
  let postData = {
    "id": generateBizUniqueID('IN'),
    "enable": 1,
    "template_id": props.templateData.id,
    "way_id": displayOptions.value[0]?.id,
    "way_type": displayOptions.value[0]?.type,
    "way_name": displayOptions.value[0]?.name,
    "content_type": formData.value.templ_type,
    "config": JSON.stringify(formData.value),
  }

  try {
    const response = await request.post('/templates/ins/addone', postData)
    if (response.status === 200 && response.data.code === 200) {
      toast.success(response.data.msg)
      // 重新加载实例列表
      await queryInsListData()
      // 清空表单
      channelName.value = ''
      formData.value = {}
    } else {
      toast.error(response.data.msg || '添加实例失败')
    }
  } catch (error: any) {
    toast.error(error.response?.data?.msg || '添加实例失败')
  }
}

// 实例表格数据
const insTableData = ref<any[]>([])

// 查询实例列表数据
const queryInsListData = async () => {
  if (!props.templateData?.id) return
  
  try {
    const response = await request.get('/templates/ins/get', {
      params: { id: props.templateData.id }
    })
    if (response.status === 200 && response.data.code === 200) {
      const insList = response.data.data.ins_list || []
      insTableData.value = insList
    }
  } catch (error) {
    console.error('获取实例列表失败', error)
  }
}

// 删除实例
const handleDeleteIns = async (insId: string) => {
  try {
    const response = await request.post('/sendtasks/ins/delete', { id: insId })
    if (response.status === 200 && response.data.code === 200) {
      toast.success(response.data.msg)
      await queryInsListData()
    } else {
      toast.error(response.data.msg || '删除失败')
    }
  } catch (error: any) {
    toast.error(error.response?.data?.msg || '删除失败')
  }
}

// 切换实例启用状态
const handleToggleEnable = async (insId: string, currentStatus: number | string) => {
  const isEnabled = Number(currentStatus) === 1
  const newStatus = isEnabled ? 0 : 1
  
  // 立即更新本地状态，提供即时反馈
  const insIndex = insTableData.value.findIndex(ins => ins.id === insId)
  if (insIndex !== -1) {
    insTableData.value[insIndex].enable = newStatus
  }
  
  try {
    const response = await request.post('/sendtasks/ins/update_enable', {
      ins_id: insId,
      status: newStatus
    })
    
    if (response.status === 200 && response.data.code === 200) {
      toast.success(response.data.msg)
      // 重新加载确保数据同步
      await queryInsListData()
    } else {
      toast.error(response.data.msg || '更新失败')
      // 失败时恢复原状态
      if (insIndex !== -1) {
        insTableData.value[insIndex].enable = currentStatus
      }
    }
  } catch (error: any) {
    console.error('状态切换失败:', error)
    toast.error(error.response?.data?.msg || '更新失败')
    // 失败时恢复原状态
    if (insIndex !== -1) {
      insTableData.value[insIndex].enable = currentStatus
    }
  }
}

// 防抖定时器
let searchTimer: number | null = null

// 搜索渠道（带防抖）
const handleSearch = (query: string) => {
  // 清除之前的定时器
  if (searchTimer) {
    clearTimeout(searchTimer)
  }
  
  if (!query.trim()) {
    displayOptions.value = []
    return
  }

  // 设置防抖延迟（500ms）
  searchTimer = window.setTimeout(async () => {
    isSearching.value = true
    try {
      const response = await request.get('/sendways/list', {
        params: { name: query }
      })
      if (response.status === 200 && response.data.code === 200) {
        displayOptions.value = response.data.data.lists.map((item: any) => ({
          id: item.id,
          name: item.name,
          type: item.type
        }))
      }
    } catch (error) {
      console.error('搜索渠道失败', error)
      displayOptions.value = []
    } finally {
      isSearching.value = false
    }
  }, 500)
}

// 监听对话框打开状态，打开时加载实例列表
watch(() => props.open, (newVal) => {
  if (newVal && props.templateData?.id) {
    queryInsListData()
  }
}, { immediate: true })
</script>

<template>
  <Dialog :open="open" @update:open="(value) => $emit('update:open', value)">
    <DialogContent class="w-[500px] max-w-[90vw] max-h-[90vh] overflow-hidden flex flex-col">
      <DialogHeader class="flex-shrink-0">
        <DialogTitle>配置发送实例</DialogTitle>
      </DialogHeader>

      <div class="px-4 pb-4 flex-1 overflow-y-auto">
        <div class="space-y-4">

        <!-- 模板信息 -->
        <div class="mb-6 p-4 bg-muted rounded-lg">
          <div class="text-sm text-muted-foreground">模板名称</div>
          <div class="text-lg font-medium">{{ templateData?.name }}</div>
          <div class="text-xs text-muted-foreground mt-1">ID: {{ templateData?.id }}</div>
        </div>

        <!-- 添加实例表单 -->
        <div class="space-y-4">
          <Label class="text-sm font-medium">选择发送渠道</Label>
          <Combobox v-model="channelName" @update:model-value="handlechannelNameChange">
            <ComboboxAnchor class="w-full">
              <ComboboxInput v-model="inputDisplayValue" @input="handleSearch(inputDisplayValue)"
                class="flex h-10 w-full" placeholder="搜索或选择渠道类型进行实例的添加..." />
            </ComboboxAnchor>
            <ComboboxList class="w-[var(--reka-combobox-trigger-width)]">
              <ComboboxViewport>
                <ComboboxItem v-for="option in displayOptions" :key="option.id" :value="option.name">
                  <div class="flex items-center justify-between w-full">
                    <span>{{ option.name }}</span>
                    <CheckIcon v-if="channelName === option.name" class="h-4 w-4" />
                  </div>
                </ComboboxItem>
                <div v-if="isSearching" class="p-2 text-sm text-muted-foreground">搜索中...</div>
                <div v-if="!isSearching && displayOptions.length === 0 && searchQuery" class="p-2 text-sm text-muted-foreground">
                  未找到匹配的渠道
                </div>
              </ComboboxViewport>
            </ComboboxList>
          </Combobox>
        </div>

        <!-- 渠道配置表单 -->
        <div v-if="currentChannelConfig" class="mt-4">
          <!-- 实例配置输入字段 -->
          <div v-if="currentChannelConfig.taskInsInputs && currentChannelConfig.taskInsInputs.length > 0" class="mb-2">
            <Label class="text-sm font-medium mb-1">实例配置</Label>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div v-for="input in currentChannelConfig.taskInsInputs" :key="input.col" class="space-y-2">
                <label class="text-xs font-medium text-muted-foreground">{{ input.label || input.desc }}</label>
                <Input v-model="formData[input.col]" :placeholder="input.desc || `请输入${input.label}`"
                  :type="input.type || 'text'" class="w-full" />
              </div>
            </div>
          </div>

          <!-- 单选框 -->
          <div v-if="currentChannelConfig.taskInsRadios && currentChannelConfig.taskInsRadios.length > 0" class="mt-4">
            <Label class="text-sm font-medium mb-2">消息格式</Label>
            <RadioGroup v-model="formData.templ_type" class="flex gap-4">
              <div v-for="radio in currentChannelConfig.taskInsRadios" :key="radio.subLabel" class="flex items-center space-x-2">
                <RadioGroupItem :value="radio.subLabel" :id="radio.subLabel" />
                <Label :for="radio.subLabel" class="text-sm cursor-pointer">{{ radio.subLabel }}</Label>
              </div>
            </RadioGroup>
          </div>
        </div>

        <div class="flex justify-end gap-2 border-b pb-4 mt-2">
          <Button @click="handleAddSubmit">添加实例</Button>
        </div>

        <!-- 关联的实例表 -->
        <div class="mt-4">
          <h3 class="text-sm font-medium mb-3">已经关联的实例</h3>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>渠道名称</TableHead>
                <TableHead>内容类型</TableHead>
                <TableHead class="text-center">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="ins in insTableData" :key="ins.id">
                <TableCell>
                  <div class="font-medium">{{ ins.way_name || '未命名' }}</div>
                  <div class="text-xs text-muted-foreground">{{ ins.way_type }}</div>
                </TableCell>
                <TableCell>
                  <Badge variant="secondary">{{ ins.content_type }}</Badge>
                </TableCell>
                <TableCell class="text-center">
                  <div class="flex items-center justify-center gap-2">
                    <Switch 
                      :model-value="ins.enable === 1" 
                      @update:model-value="() => handleToggleEnable(ins.id, ins.enable)" 
                    />
                    <Button 
                      size="sm" 
                      variant="outline" 
                      class="text-red-500 border-red-300 hover:bg-red-50 hover:border-red-400 hover:text-red-600 hover:shadow-md transition-all duration-200" 
                      @click="handleDeleteIns(ins.id)"
                    >
                      删除
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
              <TableRow v-if="!insTableData || insTableData.length === 0">
                <TableCell :colspan="3" class="h-24">
                  <EmptyTableState title="暂无实例" description="还没有配置任何实例，请先添加" />
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
