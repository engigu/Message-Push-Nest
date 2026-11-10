<script setup lang="ts">
import { ref, computed, watch, nextTick, onBeforeUnmount } from 'vue'
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
import { createChannelSearchController, getSelectedWayPayload, isSelectedWayValid } from '@/lib/channelSearchController.mjs'
import AliyunSMSTestDialog from '@/components/ui/AliyunSMSTestDialog.vue'
import { Checkbox } from '@/components/ui/checkbox'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import {
  areTemplatePlaceholdersValid,
  buildInstanceFormForChannel,
  getAliyunConfigSummary,
  getAliyunPlaceholderDiff,
  normalizeAliyunConfig,
  parseTemplatePlaceholders,
  setAliyunParamOverride,
  toAliyunEditableConfig,
  toggleAliyunParam,
  validateAliyunConfig,
} from '@/lib/aliyunSmsConfig.mjs'

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
const selectedWayId = ref<string | null>(null)
const selectedChannel = ref<{ id: string, name: string, type: string } | null>(null)
const channelName = computed(() => selectedChannel.value?.name || '')
const isComposing = ref(false)
let isHandlingChannelSelection = false

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
  return waysConfigMap.find(item => item.type === selectedChannel.value?.type) || null
})

const templatePlaceholders = computed(() => parseTemplatePlaceholders(props.data?.placeholders))
const templatePlaceholdersValid = computed(() => areTemplatePlaceholdersValid(props.data?.placeholders))
const showAliyunParamConfig = computed(() => props.type === 'template' && selectedChannel.value?.type === 'AliyunSMS')
const selectedAliyunParams = computed(() => new Set(formData.value.selected_params || []))
const addAliyunConfigError = ref('')
const isAliyunTestOpen = ref(false)
const aliyunTestScenario = ref<'new' | 'edit' | 'saved'>('new')
const aliyunTestInstanceId = ref('')
const aliyunTestConfig = ref<Record<string, any>>({})
const aliyunEditOpen = ref(false)
const editingInstance = ref<any>(null)
const editingConfig = ref<Record<string, any>>({})
const editConfigError = ref('')
const isSavingEdit = ref(false)

const openAliyunTest = (scenario: 'new' | 'edit' | 'saved', config: Record<string, any>, instanceId = '') => {
  aliyunTestScenario.value = scenario
  aliyunTestInstanceId.value = instanceId
  aliyunTestConfig.value = { ...config }
  isAliyunTestOpen.value = true
}

const updateAliyunSelection = (key: string, checked: boolean) => {
  formData.value = toggleAliyunParam(formData.value, key, checked)
  addAliyunConfigError.value = ''
}

const updateAliyunOverride = (key: string, value: string, target: 'new' | 'edit' = 'new') => {
  if (target === 'edit') editingConfig.value = setAliyunParamOverride(editingConfig.value, key, value)
  else formData.value = setAliyunParamOverride(formData.value, key, value)
  if (target === 'edit') editConfigError.value = ''
  else addAliyunConfigError.value = ''
}

const editSelectedParams = computed(() => new Set(editingConfig.value.selected_params || []))
const editPlaceholderDiff = computed(() => getAliyunPlaceholderDiff(editingConfig.value, templatePlaceholders.value))

const editParamSelected = (key: string) => editSelectedParams.value.has(key)
const updateEditAliyunSelection = (key: string, checked: boolean) => {
  editingConfig.value = toggleAliyunParam(editingConfig.value, key, checked)
  editConfigError.value = ''
}

const openEditInstance = (instance: any) => {
  if (props.type !== 'template' || instance.way_type !== 'AliyunSMS') return
  let config: Record<string, any> = {}
  try { config = JSON.parse(instance.config || '{}') } catch { config = {} }
  editingInstance.value = instance
  editingConfig.value = normalizeAliyunConfig(config)
  editConfigError.value = ''
  aliyunEditOpen.value = true
}

const saveAliyunEdit = async () => {
  const validation = validateAliyunConfig(editingConfig.value, templatePlaceholders.value, templatePlaceholdersValid.value)
  if (validation) {
    editConfigError.value = validation
    return
  }
  isSavingEdit.value = true
  try {
    const response = await request.post('/templates/ins/aliyun/edit', {
      id: editingInstance.value.id,
      config: toAliyunEditableConfig(editingConfig.value),
    })
    if (response.status === 200 && response.data.code === 200) {
      toast.success(response.data.msg || '实例配置已更新')
      aliyunEditOpen.value = false
      await queryInsListData()
    } else {
      editConfigError.value = response.data.msg || '保存实例配置失败'
    }
  } catch (error: any) {
    editConfigError.value = error.response?.data?.msg || '保存实例配置失败，请检查权限和配置'
  } finally {
    isSavingEdit.value = false
  }
}

const openSavedAliyunTest = (instance: any) => {
  let config: Record<string, any> = {}
  try { config = JSON.parse(instance.config || '{}') } catch { config = {} }
  openAliyunTest('saved', config, instance.id)
}

const channelSearch = createChannelSearchController({
  delay: 250,
  search: async (query) => {
    const response = await request.get('/sendways/list', { params: { name: query } })
    if (response.status !== 200 || response.data.code !== 200) return []
    return response.data.data.lists.map((item: any) => ({
      id: item.id,
      name: item.name,
      type: item.type
    }))
  },
  onOptions: (options) => { displayOptions.value = options },
  onLoading: (loading) => { isSearching.value = loading }
})

const getChannelDisplayValue = (wayId: string | null) => {
  if (!wayId) return ''
  return selectedChannel.value?.id === wayId
    ? selectedChannel.value.name
    : displayOptions.value.find(option => option.id === wayId)?.name || ''
}

const handleSelectedWayChange = (value: unknown) => {
  const wayId = typeof value === 'string' ? value : null
  selectedWayId.value = wayId
  if (!wayId) {
    selectedChannel.value = null
    return
  }

  const option = displayOptions.value.find(item => item.id === wayId)
  if (!option) {
    selectedChannel.value = null
    return
  }

  const channelChanged = selectedChannel.value?.id !== option.id
  isHandlingChannelSelection = true
  selectedChannel.value = option
  searchQuery.value = option.name
  channelSearch.clear()

  if (channelChanged) {
    const channelConfig = currentChannelConfig.value
    const messageFormat = channelConfig?.taskInsRadios?.[0]?.subLabel || 'text'
    formData.value = props.type === 'template'
      ? buildInstanceFormForChannel(option.type, templatePlaceholders.value, messageFormat)
      : { allowMultiRecip: false, templ_type: messageFormat }
    addAliyunConfigError.value = ''
  }

  nextTick(() => { isHandlingChannelSelection = false })
}

const handleCompositionStart = () => {
  isComposing.value = true
  selectedChannel.value = null
  selectedWayId.value = null
  channelSearch.startComposition()
}

const handleCompositionEnd = () => {
  nextTick(() => {
    isComposing.value = false
    channelSearch.endComposition(searchQuery.value)
  })
}

const handleSearchQueryChange = (query: string) => {
  if (isHandlingChannelSelection) return

  if (selectedChannel.value && selectedChannel.value.name !== query) {
    selectedChannel.value = null
    selectedWayId.value = null
  }

  channelSearch.search(query)
}

// 表单数据
const formData = ref<Record<string, any>>({
  allowMultiRecip: false  // 默认false为固定模式，true为动态模式
})

// 是否显示接收者输入框
const shouldShowRecipientInput = computed(() => {
  // 支持动态接收者 且 未勾选（固定模式）时显示输入框
  return currentChannelConfig.value?.dynamicRecipient?.support && !formData.value.allowMultiRecip
})

// 添加单条实例配置
const handleAddSubmit = async () => {
  const selectedWay = getSelectedWayPayload(selectedChannel.value)
  if (!selectedWay || !isSelectedWayValid(selectedChannel.value, selectedWayId.value)) {
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

  if (showAliyunParamConfig.value) {
    const validation = validateAliyunConfig(formData.value, templatePlaceholders.value, templatePlaceholdersValid.value)
    if (validation) {
      addAliyunConfigError.value = validation
      return
    }
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
    ...selectedWay,
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
      selectedChannel.value = null
      selectedWayId.value = null
      searchQuery.value = ''
      channelSearch.clear()
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

  // 通用逻辑：对于没有 dynamicRecipient 但有其他输入配置的渠道，将其展示为 "标签: 值"
  if (channelConfig?.taskInsInputs && Array.isArray(channelConfig.taskInsInputs) && channelConfig.taskInsInputs.length > 0) {
    return channelConfig.taskInsInputs
      .map((inp: any) => {
        const val = config[inp.col]
        return val ? `${inp.label || inp.desc}: ${val}` : ''
      })
      .filter(Boolean)
      .join(' | ') || "-"
  }

   if (channelConfig?.taskInsInputs && Array.isArray(channelConfig.taskInsInputs) && channelConfig.taskInsInputs.length === 0) {
    return "无需配置"
  }
  return ""
}

const getInsAliyunSummary = (row: any) => row.way_type === 'AliyunSMS' ? getAliyunConfigSummary(row.config) : null

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

// 监听数据变化，自动加载实例列表
watch(() => props.data?.id, (newVal) => {
  if (newVal) {
    queryInsListData()
  }
}, { immediate: true })

// 监听搜索词变化，组合输入期间不查询拼音中间态
watch(searchQuery, (newVal) => {
  if (isComposing.value) return
  handleSearchQueryChange(newVal)
})

onBeforeUnmount(() => channelSearch.clear())

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
          <Combobox v-model="selectedWayId" :ignore-filter="true" :reset-search-term-on-select="false" @update:model-value="handleSelectedWayChange">
            <ComboboxAnchor class="w-full">
              <ComboboxInput 
                v-model="inputDisplayValue" 
                :display-value="getChannelDisplayValue"
                @compositionstart="handleCompositionStart"
                @compositionend="handleCompositionEnd"
                class="flex h-10 w-full" 
                placeholder="搜索或选择渠道类型进行实例的添加..." 
              />
            </ComboboxAnchor>
            <ComboboxList class="w-[var(--reka-combobox-trigger-width)]">
              <ComboboxViewport>
                <ComboboxItem v-for="option in displayOptions" :key="option.id" :value="option.id" :text-value="option.name">
                  <div class="flex items-center justify-between w-full">
                    <span>{{ option.name }}</span>
                    <CheckIcon v-if="selectedChannel?.id === option.id" class="h-4 w-4" />
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
        <div class="flex gap-2">
          <Button v-if="showAliyunParamConfig" size="sm" variant="secondary" :disabled="!isSelectedWayValid(selectedChannel, selectedWayId)" @click="openAliyunTest('new', formData)">发送测试</Button>
          <Button size="sm" variant="outline" :disabled="!isSelectedWayValid(selectedChannel, selectedWayId)" @click="handleAddSubmit">添加实例</Button>
        </div>
      </div>
    </div>

    <!-- 渠道配置表单 -->
    <div v-if="currentChannelConfig" class="mt-4 space-y-4">
      <!-- 动态接收者勾选框 -->
      <div v-if="currentChannelConfig.dynamicRecipient?.support" class="p-3 border rounded-lg bg-gray-50 dark:bg-gray-800/50">
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
      <div v-if="shouldShowRecipientInput" class="space-y-2">
        <Label class="text-sm font-medium text-gray-700 dark:text-gray-300 block mb-1">实例配置</Label>
        <div class="grid grid-cols-1 gap-4">
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
      <div v-if="currentChannelConfig.taskInsInputs && currentChannelConfig.taskInsInputs.length > 0" class="space-y-2">
        <Label class="text-sm font-medium block mb-1">实例配置</Label>
        <div class="grid grid-cols-1 gap-4">
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

      <section v-if="showAliyunParamConfig" class="space-y-3 rounded-lg border p-4">
        <div>
          <h3 class="text-sm font-semibold">阿里云短信模板参数</h3>
          <p class="mt-1 text-xs text-muted-foreground">业务变量来自当前消息模板；默认同名映射。取消勾选后，该变量不会发送。</p>
        </div>
        <div v-if="templatePlaceholders.length === 0" class="text-sm text-amber-700">
          当前模板没有可用的业务变量。请确认模板 Placeholders 已正确配置。
        </div>
        <div v-else class="space-y-3">
          <div v-for="placeholder in templatePlaceholders" :key="placeholder.key" class="grid grid-cols-[auto_minmax(0,1fr)_minmax(0,1fr)] items-center gap-3 rounded-md bg-muted/40 p-3">
            <Checkbox
              :model-value="selectedAliyunParams.has(placeholder.key)"
              :aria-label="`发送变量 ${placeholder.key}`"
              @update:model-value="(checked: boolean | 'indeterminate') => updateAliyunSelection(placeholder.key, checked === true)"
            />
            <div class="min-w-0">
              <div class="truncate text-sm font-medium">{{ placeholder.label }}</div>
              <div class="text-xs text-muted-foreground">业务变量：{{ placeholder.key }}</div>
            </div>
            <div class="space-y-1">
              <Label :for="`aliyun-map-${placeholder.key}`" class="text-xs">阿里云参数名</Label>
              <Input
                :id="`aliyun-map-${placeholder.key}`"
                :model-value="formData.param_name_overrides?.[placeholder.key] || placeholder.key"
                :disabled="!selectedAliyunParams.has(placeholder.key)"
                :placeholder="placeholder.key"
                @update:model-value="(value: string | number) => updateAliyunOverride(placeholder.key, String(value))"
              />
            </div>
          </div>
        </div>
        <div v-if="selectedAliyunParams.size" class="rounded-md bg-muted p-3 text-xs">
          <div class="font-medium">参数映射预览</div>
          <div v-for="key in formData.selected_params" :key="key" class="mt-1">
            {{ key }} → {{ formData.param_name_overrides?.[key] || key }}
          </div>
        </div>
        <p v-if="addAliyunConfigError" role="alert" class="text-sm text-destructive">{{ addAliyunConfigError }}</p>
      </section>

      <!-- 单选框 -->
      <div v-if="currentChannelConfig.taskInsRadios && currentChannelConfig.taskInsRadios.length > 0" class="space-y-2">
        <Label class="text-sm font-medium block mb-1">消息格式</Label>
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
              <div v-if="props.type === 'template' && ins.way_type === 'AliyunSMS'" class="space-y-1 text-left">
                <div><Badge variant="secondary">{{ getInsAliyunSummary(ins)?.templateCode || '未设置 TemplateCode' }}</Badge></div>
                <div class="flex flex-wrap gap-1 text-xs">
                  <Badge variant="outline">{{ getInsAliyunSummary(ins)?.mode === 'structured' ? 'structured' : 'legacy content' }}</Badge>
                  <Badge v-if="getInsAliyunSummary(ins)?.selectedCount !== null" variant="outline">已选 {{ getInsAliyunSummary(ins)?.selectedCount }} 项</Badge>
                  <Badge v-if="getInsAliyunSummary(ins)?.overrideCount !== null" variant="outline">覆盖映射 {{ getInsAliyunSummary(ins)?.overrideCount }} 项</Badge>
                </div>
                <div class="text-xs text-muted-foreground">{{ formatInsConfigDisplay(ins) || '-' }}</div>
              </div>
              <template v-else>
                <Badge v-if="formatInsConfigDisplay(ins)" variant="secondary">{{ formatInsConfigDisplay(ins) }}</Badge>
                <span v-else class="text-sm text-muted-foreground">-</span>
              </template>
            </TableCell>
            <TableCell class="text-center">
              <div class="flex items-center justify-center gap-2">
                <Switch 
                  :model-value="ins.enable === 1" 
                  @update:model-value="() => handleToggleEnable(ins.id, ins.enable)" 
                />
                <template v-if="props.type === 'template' && ins.way_type === 'AliyunSMS'">
                  <Button size="sm" variant="outline" @click="openEditInstance(ins)">编辑</Button>
                  <Button size="sm" variant="secondary" @click="openSavedAliyunTest(ins)">发送测试</Button>
                </template>
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

    <Dialog :open="aliyunEditOpen" @update:open="(open) => aliyunEditOpen = open">
      <DialogContent class="max-h-[90vh] max-w-2xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>编辑阿里云短信实例</DialogTitle>
          <DialogDescription>只更新实例 Config，实例 ID、模板和渠道关联保持不变。</DialogDescription>
        </DialogHeader>
        <div class="space-y-4">
          <div class="space-y-2">
            <Label for="edit-aliyun-template-code">短信 TemplateCode</Label>
            <Input id="edit-aliyun-template-code" v-model="editingConfig.template_code" placeholder="例如 SMS_123456789" />
          </div>
          <div class="space-y-2">
            <Label for="edit-aliyun-phone">固定接收手机号</Label>
            <Input id="edit-aliyun-phone" v-model="editingConfig.phone_number" placeholder="保留或填写实例接收手机号" />
          </div>
          <div v-if="editingConfig.template_param_mode !== 'structured'" class="rounded-md border border-amber-300 bg-amber-50 p-3 text-sm text-amber-900">
            此历史实例使用兼容模式（legacy content）。本次编辑不会切换到 structured；如需多参数配置，请新建阿里云短信实例。
          </div>
          <template v-else>
            <div v-if="editPlaceholderDiff.added.length" class="rounded-md border border-blue-300 bg-blue-50 p-3 text-sm text-blue-900">
              模板新增变量不会自动加入此实例。需要发送时请手动勾选。
            </div>
            <div v-if="editPlaceholderDiff.removed.length" class="rounded-md border border-amber-300 bg-amber-50 p-3 text-sm text-amber-900">
              以下已选变量已从模板删除或改名；请取消选择后再保存：{{ editPlaceholderDiff.removed.join('、') }}
            </div>
            <div class="space-y-3">
              <div v-for="placeholder in templatePlaceholders" :key="placeholder.key" class="grid grid-cols-[auto_minmax(0,1fr)_minmax(0,1fr)] items-center gap-3 rounded-md bg-muted/40 p-3">
                <Checkbox
                  :model-value="editParamSelected(placeholder.key)"
                  :aria-label="`编辑发送变量 ${placeholder.key}`"
                  @update:model-value="(checked: boolean | 'indeterminate') => updateEditAliyunSelection(placeholder.key, checked === true)"
                />
                <div class="min-w-0">
                  <div class="truncate text-sm font-medium">{{ placeholder.label }}</div>
                  <div class="text-xs text-muted-foreground">业务变量：{{ placeholder.key }}</div>
                  <div v-if="!editParamSelected(placeholder.key)" class="text-xs text-blue-700">模板新增或当前未选择</div>
                </div>
                <div class="space-y-1">
                  <Label :for="`edit-aliyun-map-${placeholder.key}`" class="text-xs">阿里云参数名</Label>
                  <Input
                    :id="`edit-aliyun-map-${placeholder.key}`"
                    :model-value="editingConfig.param_name_overrides?.[placeholder.key] || placeholder.key"
                    :disabled="!editParamSelected(placeholder.key)"
                    @update:model-value="(value: string | number) => updateAliyunOverride(placeholder.key, String(value), 'edit')"
                  />
                </div>
              </div>
            </div>
            <div v-if="editPlaceholderDiff.removed.length" class="space-y-2 rounded-md border p-3">
              <div class="text-sm font-medium">移除失效变量</div>
              <div v-for="key in editPlaceholderDiff.removed" :key="key" class="flex items-center gap-2 text-sm">
                <Checkbox :model-value="true" :aria-label="`移除失效变量 ${key}`" @update:model-value="() => updateEditAliyunSelection(key, false)" />
                <span>{{ key }}（取消勾选后保存将从此实例移除）</span>
              </div>
            </div>
            <div v-if="editingConfig.selected_params?.length" class="rounded-md bg-muted p-3 text-xs">
              <div class="font-medium">参数映射预览</div>
              <div v-for="key in editingConfig.selected_params" :key="key" class="mt-1">{{ key }} → {{ editingConfig.param_name_overrides?.[key] || key }}</div>
            </div>
          </template>
          <p v-if="editConfigError" role="alert" class="text-sm text-destructive">{{ editConfigError }}</p>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="openAliyunTest('edit', editingConfig, editingInstance?.id)">发送测试</Button>
          <Button variant="outline" @click="aliyunEditOpen = false">取消</Button>
          <Button :disabled="isSavingEdit" @click="saveAliyunEdit">{{ isSavingEdit ? '保存中...' : '保存配置' }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <AliyunSMSTestDialog
      :open="isAliyunTestOpen"
      :scenario="aliyunTestScenario"
      :instance-id="aliyunTestInstanceId"
      :template-id="props.data?.id || ''"
      :way-id="aliyunTestScenario === 'new' ? selectedWayId || '' : editingInstance?.way_id || ''"
      :config="aliyunTestConfig"
      :placeholders="templatePlaceholders"
      @update:open="(open: boolean) => isAliyunTestOpen = open"
    />
  </div>
</template>
