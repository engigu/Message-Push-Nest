<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { Badge } from "@/components/ui/badge"
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
  // 实例类型：'template' 或 'task'
  type: 'template' | 'task'
  // 关联的数据（模板数据或任务数据）
  data: any
  // 是否在对话框中显示（用于模板）
  inDialog?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  inDialog: false
})

// API 配置映射
const apiConfig = computed(() => {
  if (props.type === 'template') {
    return {
      addIns: '/templates/ins/addone',
      getIns: '/templates/ins/get',
      deleteIns: '/sendtasks/ins/delete',
      updateEnable: '/sendtasks/ins/update_enable',
      idField: 'template_id',
      nameField: 'name'
    }
  } else {
    return {
      addIns: '/sendtasks/ins/addone',
      getIns: '/sendtasks/ins/gettask',
      deleteIns: '/sendtasks/ins/delete',
      updateEnable: '/sendtasks/ins/update_enable',
      idField: 'task_id',
      nameField: 'name'
    }
  }
})

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
  let rs = waysConfigMap.find(item => item.type === type) || null
  
  return rs
})

// 表单数据
const formData = ref<Record<string, any>>({
  allowMultiRecip: false  // 默认false为固定模式，true为动态模式
})

// 是否显示接收者输入框
const shouldShowRecipientInput = computed(() => {
  // 支持动态接收者 且 未勾选（固定模式）时显示输入框
  return currentChannelConfig.value?.dynamicRecipient?.support && !formData.value.allowMultiRecip
})

// 监听渠道变化
const handlechannelNameChange = () => {
  // 数据加载后，text/html单选设置默认选中（这里选第一个）
  if (currentChannelConfig.value?.taskInsRadios.length > 0) {
    formData.value.templ_type = currentChannelConfig.value?.taskInsRadios[0].subLabel
  }
  // 重置动态接收者设置
  formData.value.allowMultiRecip = false
}

// 添加单条实例配置
const handleAddSubmit = async () => {
  // 验证是否选择了渠道
  if (!channelName.value) {
    toast.error('请选择发送渠道')
    return
  }

  // 检查动态接收和固定接收不能混合使用
  if (insTableData.value.length > 0) {
    const hasDynamicInstance = insTableData.value.some(ins => {
      try {
        const config = JSON.parse(ins.config)
        return config.allowMultiRecip === true
      } catch {
        return false
      }
    })
    
    const entityName = props.type === 'template' ? '模板' : '任务'
    
    // 如果要添加动态接收实例，但已有其他实例
    if (formData.value.allowMultiRecip === true) {
      if (hasDynamicInstance) {
        toast.error(`该${entityName}已存在动态接收实例，一个${entityName}只能配置一个动态接收实例`)
        return
      }
      if (insTableData.value.length > 0) {
        toast.error(`动态接收实例不能与固定接收实例混合使用，请先删除所有固定实例`)
        return
      }
    }
    
    // 如果要添加固定接收实例，但已有动态接收实例
    if (formData.value.allowMultiRecip !== true && hasDynamicInstance) {
      toast.error(`该${entityName}已配置动态接收实例，不能再添加固定接收实例`)
      return
    }
  }

  // 验证内容类型
  const contentType = formData.value.templ_type
  if (!contentType) {
    toast.error('请选择消息格式')
    return
  }

  // 仅模板需要验证对应格式的内容是否为空
  if (props.type === 'template') {
    const templateFieldMap: Record<string, string> = {
      'text': 'text_template',
      'html': 'html_template',
      'markdown': 'markdown_template'
    }
    
    const fieldName = templateFieldMap[contentType.toLowerCase()]
    if (fieldName) {
      const templateContent = props.data?.[fieldName] || ''
      // 检查是否为空（去除所有空白字符后检查）
      if (!templateContent.trim()) {
        toast.error(`模板的 ${contentType} 格式内容为空，无法添加此类型的实例`)
        return
      }
    }
  }

  // 组建表单数据
  let postData: Record<string, any> = {
    "id": generateBizUniqueID('IN'),
    "enable": 1,
    [apiConfig.value.idField]: props.data.id,
    "way_id": displayOptions.value[0]?.id,
    "way_type": displayOptions.value[0]?.type,
    "way_name": displayOptions.value[0]?.name,
    "content_type": formData.value.templ_type,
    "config": JSON.stringify(formData.value),
  }

  try {
    const response = await request.post(apiConfig.value.addIns, postData)
    if (response.status === 200 && response.data.code === 200) {
      toast.success(response.data.msg)
      // 重新加载实例列表
      await queryInsListData()
      // 清空表单
      channelName.value = ''
      formData.value = { allowMultiRecip: false }
    } else {
      toast.error(response.data.msg || '添加实例失败')
    }
  } catch (error: any) {
    toast.error(error.response?.data?.msg || '添加实例失败')
  }
}

// 实例表格数据
const insTableData = ref<any[]>([])

// 格式化额外信息列的值
const formatInsConfigDisplay = (row: any) => {
  if (!row.config) {
    return "-"
  }
  let config = JSON.parse(row.config)
  
  // 检查是否为动态接收者模式
  if (config.allowMultiRecip === true) {
    return "动态接收"
  }
  
  // 固定模式，根据 constant.js 配置动态获取接收者字段
  const channelConfig = CONSTANT.WAYS_DATA.find((item: any) => item.type === row.way_type)
  if (channelConfig?.dynamicRecipient?.support) {
    const recipientField = channelConfig.dynamicRecipient.field
    return config[recipientField] || ""
  }

   if (channelConfig?.taskInsInputs && Array.isArray(channelConfig.taskInsInputs) && channelConfig.taskInsInputs.length === 0) {
    return "无需配置"
  }
  return ""
}

// 查询实例列表数据
const queryInsListData = async () => {
  if (!props.data?.id) return
  
  try {
    const response = await request.get(apiConfig.value.getIns, {
      params: { id: props.data.id }
    })
    if (response.status === 200 && response.data.code === 200) {
      // 模板返回 ins_list，任务返回 ins_data
      const insList = response.data.data.ins_list || response.data.data.ins_data || []
      insTableData.value = insList
    }
  } catch (error) {
    console.error('获取实例列表失败', error)
  }
}

// 删除实例
const handleDeleteIns = async (insId: string) => {
  try {
    const response = await request.post(apiConfig.value.deleteIns, { id: insId })
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
    const response = await request.post(apiConfig.value.updateEnable, {
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

// 监听数据变化，自动加载实例列表
watch(() => props.data?.id, (newVal) => {
  if (newVal) {
    queryInsListData()
  }
}, { immediate: true })

// 暴露方法供父组件调用
defineExpose({
  queryInsListData
})
</script>

<template>
  <div class="space-y-4" :class="{ 'px-4 pb-4': inDialog }">
    <!-- 信息展示区域 -->
    <div v-if="data" class="p-3 bg-muted rounded-lg space-y-1">
      <div class="flex items-baseline gap-2">
        <span class="text-base font-semibold">{{ data[apiConfig.nameField] }}</span>
        <Badge variant="outline" class="text-xs">{{ data.id }}</Badge>
      </div>
      <div class="text-xs text-muted-foreground">
        为此{{ type === 'template' ? '模板' : '任务' }}配置发送实例
      </div>
    </div>

    <!-- 添加实例表单 -->
    <div class="space-y-4">
      <div class="flex items-end gap-2">
        <div class="flex-1 space-y-2">
          <Label class="text-sm font-medium">选择发送渠道</Label>
          <Combobox v-model="channelName" @update:model-value="handlechannelNameChange">
            <ComboboxAnchor class="w-full">
              <ComboboxInput 
                v-model="inputDisplayValue" 
                @input="handleSearch(inputDisplayValue)"
                class="flex h-10 w-full" 
                placeholder="搜索或选择渠道类型进行实例的添加..." 
              />
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
        <Button size="sm" variant="outline" @click="handleAddSubmit">添加实例</Button>
      </div>
    </div>

    <!-- 渠道配置表单 -->
    <div v-if="currentChannelConfig" class="mt-4">
      <!-- 动态接收者勾选框 -->
      <div v-if="currentChannelConfig.dynamicRecipient?.support" class="mb-4 p-3 border rounded-lg bg-gray-50 dark:bg-gray-800/50">
        <div class="flex items-center space-x-2">
          <Switch 
            :model-value="formData.allowMultiRecip" 
            @update:model-value="(val: boolean) => formData.allowMultiRecip = val"
            :id="`allow-multi-${channelName}`" 
          />
          <Label :for="`allow-multi-${channelName}`" class="text-sm font-medium cursor-pointer">
            动态接收者模式
          </Label>
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 ml-8">
          {{ formData.allowMultiRecip ? '支持动态接收者，发送时通过API指定接收者列表（群发模式）' : '固定接收者模式，需要在下方配置固定接收者' }}
        </p>
        <p v-if="formData.allowMultiRecip" class="text-xs text-orange-500 dark:text-orange-400 mt-1 ml-8 font-medium">
          ⚠️ 注意：一个{{ type === 'template' ? '模板' : '任务' }}只能配置一个动态接收实例，且不能与固定接收实例混合使用
        </p>
      </div>

      <!-- 接收者输入字段 -->
      <div v-if="shouldShowRecipientInput" class="mb-2">
        <Label class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">实例配置</Label>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="space-y-2">
            <label class="text-xs font-medium text-gray-600 dark:text-gray-400">
              {{ currentChannelConfig.dynamicRecipient.label }}
            </label>
            <Input 
              v-model="formData[currentChannelConfig.dynamicRecipient.field]" 
              :placeholder="`请输入${currentChannelConfig.dynamicRecipient.desc}`"
              type="text" 
              class="text-sm" 
            />
          </div>
        </div>
      </div>
      
      <!-- 实例配置输入字段（排除动态接收者字段） -->
      <div v-if="currentChannelConfig.taskInsInputs && currentChannelConfig.taskInsInputs.length > 0" class="mb-2">
        <Label class="text-sm font-medium mb-1">实例配置</Label>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div 
            v-for="input in currentChannelConfig.taskInsInputs.filter((inp: any) => inp.col !== currentChannelConfig?.dynamicRecipient?.field)" 
            :key="input.col" 
            class="space-y-2"
          >
            <label class="text-xs font-medium text-muted-foreground">{{ input.label || input.desc }}</label>
            <Input 
              v-model="formData[input.col]" 
              :placeholder="input.desc || `请输入${input.label}`"
              :type="input.type || 'text'" 
              class="w-full" 
            />
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

    <!-- 关联的实例表 -->
    <div class="mt-4">
      <h3 class="text-sm font-medium mb-3">已经关联的实例</h3>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>渠道名称</TableHead>
            <TableHead>内容类型</TableHead>
            <TableHead>接收者</TableHead>
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
            <TableCell>
              <Badge v-if="formatInsConfigDisplay(ins)" variant="secondary">{{ formatInsConfigDisplay(ins) }}</Badge>
              <span v-else class="text-sm text-muted-foreground">-</span>
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
            <TableCell :colspan="4" class="h-24">
              <EmptyTableState title="暂无实例" description="还没有配置任何实例，请先添加" />
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </div>
</template>
