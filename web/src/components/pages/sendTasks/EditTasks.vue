<script setup lang="ts">
import { ref, computed, defineEmits, defineProps, withDefaults, onMounted } from 'vue'
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
  open?: boolean
  editData?: any // 编辑时传入的数据
}
const props = withDefaults(defineProps<Props>(), {
  open: false,
  editData: null
})

// 组件emits
const emit = defineEmits<{
  'update:open': [value: boolean]
  'save': [data: any]
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

// 当前选中任务的配置
const currentChannelConfig = computed(() => {
  // 先根据label找到type
  let type = displayOptions.value.find(item => item.name === channelName.value)?.type
  // 再根据type找到配置
  let rs = waysConfigMap.find(item => item.type === type) || null;
  
  return rs;
})

// 表单数据
const formData = ref<Record<string, any>>({})

// 监听任务模式变化
const handlechannelNameChange = () => {
  // 数据加载后，text/html单选设置默认选中（这里选第一个）
  if (currentChannelConfig.value?.taskInsRadios.length > 0) {
    formData.value.taskInsRadio = currentChannelConfig.value?.taskInsRadios[0].subLabel
  }
}

// 关闭drawer
const handleClose = () => {
  emit('update:open', false)
}

// 添加单条实例配置
const handleAddSubmit = async () => {
  // 组建表单数据
  let postData = {
    "id": generateBizUniqueID('I'),
    "enable": 1,
    "task_id": props.editData.id,
    "way_id": displayOptions.value[0]?.id,
    "way_type": displayOptions.value[0]?.type,
    "way_name": displayOptions.value[0]?.name,
    "content_type": formData.value.taskInsRadio,
    "config": "{}"
  };
  const { taskInsRadio, ...configValue } = formData.value
  if (configValue) {
    postData.config = JSON.stringify(configValue)
  }

  const rsp = await request.post('/sendtasks/ins/addone', postData);
  if (await rsp.data.code == 200) {
    // 动态添加表格第一行
    const newItem = {
      ...postData,
      created_by: '',
      modified_by: '',
      created_on: new Date().toISOString(),
      modified_on: new Date().toISOString(),
      extra: ''
    };
    insTableData.value.unshift(newItem);
    toast.success(rsp.data.msg);
  }
}

const insTableData = ref<Array<{
  id: string;
  created_by: string;
  modified_by: string;
  created_on: string;
  modified_on: string;
  task_id: string;
  way_id: string;
  way_type: string;
  content_type: string;
  config: string;
  extra: string;
  enable: number;
  way_name: string;
}>>([]);

const queryInsListData = async () => {
  let params = { id: props.editData.id };
  const rsp = await request.get('/sendtasks/ins/gettask', { params: params });
  insTableData.value = await rsp.data.data.ins_data;
}



// 防抖定时器
let searchTimer: number | null = null

// 搜索处理函数（带防抖）
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
      const results = await querySearchWayAsync(query)
      displayOptions.value = results.map((item: any) => ({
        id: item.id,
        name: item.name,
        type: item.type
      }))
    } catch (error) {
      console.error('搜索失败:', error)
      displayOptions.value = []
    } finally {
      isSearching.value = false
    }
  }, 500)
}



const querySearchWayAsync = async (query: string) => {
  const params = { name: query }
  const rsp = await request.get('/sendways/list', { params })
  return rsp.data.data.lists
}

const getWayTypeText = (type: string) => {
  const wayData = CONSTANT.WAYS_DATA.find(item => item.type === type)
  return wayData ? wayData.label : type
}

// 切换启用状态
const toggleEnable = async (item: any) => {
  const newStatus = item.enable === 1 ? 0 : 1
  let postData = { ins_id: item.id, status: newStatus };
  const rsp = await request.post('/sendtasks/ins/update_enable', postData);
  if (await rsp.data.code == 200) {
    toast.success(rsp.data.msg);
    // 更新本地状态
    item.enable = newStatus
  }
}

const handleDelete = async (item: any) => {
  const rsp = await request.post('/sendtasks/ins/delete', { id: item.id });
  if (rsp.status == 200) {
    toast.success(rsp.data.msg);
    insTableData.value = insTableData.value.filter((ins) => ins.id !== item.id)
  }
}

const handleEditTask = async () => {
  let postData = { id: props.editData.id, name: props.editData.name };
  const rsp = await request.post('/sendtasks/edit', postData);
  if (await rsp.data.code == 200) {
    toast.success(rsp.data.msg);
  }
}

// 格式化额外信息列的值
const formatInsConfigDisplay = (row: any) => {
  if (!row.config) {
    return ""
  }
  if (["Email", "WeChatOFAccount"].includes(row.way_type)) {
    let config = JSON.parse(row.config)
    let info = `${config.to_account}`
    return info
  } else {
    return ""
  }
}

// 组件挂载时加载实例配置列表
onMounted(() => {
  queryInsListData();
})
</script>

<template>
  <div class="w-full">
    <!-- 任务信息展示区域 -->
    <div v-if="props.editData" class="flex flex-col sm:flex-row sm:items-center gap-2 border-b p-4">
      <Label class="w-16 sm:w-16">任务名称</Label>
      <Input v-model="props.editData.name" placeholder="请输入任务名称" class="w-full sm:w-64" />
      <Button class="w-full sm:w-auto sm:ml-auto" @click="handleEditTask">修改</Button>
    </div>



    <div class="mt-4">
      <div class="flex gap-4">

        <div class="flex-1">
          <Combobox v-model="channelName" @update:model-value="handlechannelNameChange">
            <ComboboxAnchor class="w-full">
              <ComboboxInput v-model="inputDisplayValue" @input="handleSearch(inputDisplayValue)"
                class="flex h-10 w-full " placeholder="搜索或选择渠道类型进行实例的添加..." />
            </ComboboxAnchor>
            <ComboboxList class="w-[var(--reka-combobox-trigger-width)]">
              <ComboboxViewport>
                <div v-if="isSearching" class="px-2 py-1.5 text-sm text-muted-foreground">
                  搜索中...
                </div>
                <div v-else-if="searchQuery && displayOptions.length === 0"
                  class="px-2 py-1.5 text-sm text-muted-foreground">
                  未找到匹配项
                </div>
                <ComboboxItem v-for="option in displayOptions" :key="option.id" :value="option.name"
                  class="relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50">
                  <CheckIcon class="mr-2 h-4 w-4" :class="channelName === option.name ? 'opacity-100' : 'opacity-0'" />
                  {{ option.name }}
                </ComboboxItem>
              </ComboboxViewport>
            </ComboboxList>
          </Combobox>
        </div>
      </div>
      <!-- 动态任务配置区域 -->
      <div v-if="currentChannelConfig" class="mt-4">
        <!-- 任务指令输入字段 -->
        <div v-if="currentChannelConfig.taskInsInputs && currentChannelConfig.taskInsInputs.length > 0" class="mb-2">
          <Label class="text-sm font-medium text-gray-700 mb-1">实例配置</Label>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div v-for="input in currentChannelConfig.taskInsInputs" :key="input.col" class="space-y-2">
              <label class="text-xs font-medium text-gray-600">{{ input.label }}</label>
              <Input v-model="formData[input.col]" :placeholder="input.desc || `请输入${input.desc}`"
                :type="input.type || 'text'" class="text-sm" />
            </div>
          </div>
        </div>

        <!-- 任务指令单选项 -->
        <div v-if="currentChannelConfig.taskInsRadios && currentChannelConfig.taskInsRadios.length > 0">
          <Label class="text-sm font-medium text-gray-700 mb-1">消息格式</Label>
          <RadioGroup v-model="formData.taskInsRadio" class="flex flex-wrap gap-4">
            <div v-for="radio in currentChannelConfig.taskInsRadios" :key="radio.value"
              class="flex items-center space-x-2">
              <RadioGroupItem :value="radio.subLabel" :id="radio.subLabel" />
              <Label :for="radio.value"
                class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                {{ radio.subLabel }}
              </Label>
            </div>
          </RadioGroup>
        </div>
      </div>
    </div>
  </div>

  <div class="flex justify-end gap-2  border-b pb-4 mt-2">
    <Button variant="outline" @click="handleClose">取消</Button>
    <Button @click="handleAddSubmit">添加实例</Button>
  </div>

  <!-- 关联的实例表 -->
  <div class="mt-4">
    <h3 class="text-sm font-medium text-gray-900 mb-3">已经关联的实例</h3>
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>渠道名</TableHead>
          <TableHead>渠道/内容</TableHead>
          <TableHead class="text-center">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="item in insTableData" :key="item.id">
          <TableCell>{{ item.way_name }}</TableCell>
          <TableCell>
            <Badge variant="outline">
              {{ getWayTypeText(item.way_type) }}
            </Badge>
            <Badge variant="outline">
              {{ item.content_type }}
            </Badge>
            <Badge variant="outline" v-if="item.config !== '{}'">
              {{ formatInsConfigDisplay(item) }}
            </Badge>
          </TableCell>
          <TableCell class="text-center">
            <div class="flex items-center justify-center gap-2">
              <Switch :model-value="item.enable === 1" @update:model-value="() => toggleEnable(item)" />
              <Button size="sm" variant="outline" class="text-red-500 border-red-300 hover:bg-red-50 
              hover:border-red-400 hover:text-red-600 hover:shadow-md
               transition-all duration-200" @click="() => handleDelete(item)">删除</Button>
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
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'EditTasks'
})
</script>