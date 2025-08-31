<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { request } from '@/api/api'

interface CronMessageFormData {
  id?: number
  name: string
  cron_expression: string
  title: string
  content: string
  task_id: string
  url: string
}

interface Props {
  modelValue: CronMessageFormData
  mode: 'add' | 'edit'
  loading?: boolean
}

interface Emits {
  (e: 'update:modelValue', value: CronMessageFormData): void
  (e: 'submit'): void
  (e: 'cancel'): void
}

const props = withDefaults(defineProps<Props>(), {
  loading: false
})
const emit = defineEmits<Emits>()

defineOptions({
  name: 'CronMessageForm'
})

// 本地表单数据
const localFormData = reactive<CronMessageFormData>({ ...props.modelValue })

// 监听外部数据变化
watch(() => props.modelValue, (newValue) => {
  Object.assign(localFormData, newValue)
}, { deep: true })

// 监听本地数据变化，同步到外部
watch(localFormData, (newValue) => {
  emit('update:modelValue', newValue)
}, { deep: true })

// 可用的发信任务列表
const availableTasks = ref<Array<{ id: number, name: string }>>([])

// 常用的 Cron 表达式模板
const cronTemplates = [
  { label: '每分钟', value: '* * * * *' },
  { label: '每5分钟', value: '*/5 * * * *' },
  { label: '每小时', value: '0 * * * *' },
  { label: '每天凌晨2点', value: '0 2 * * *' },
  { label: '每周一凌晨2点', value: '0 2 * * 1' },
  { label: '每月1号凌晨2点', value: '0 2 1 * *' }
]

// 加载可用的发信任务
const loadAvailableTasks = async () => {
  try {
    const rsp = await request.get('/sendtasks/list', { params: { page: 1, size: 100 } })
    availableTasks.value = rsp.data.data.lists.map((task: any) => ({
      id: task.id,
      name: task.name
    }))
  } catch (error) {
    console.error('加载发信任务失败:', error)
  }
}

// 应用 Cron 模板
const applyCronTemplate = (template: string) => {
  localFormData.cron_expression = template
}

// 提交表单
const handleSubmit = () => {
  emit('submit')
}

// 取消操作
const handleCancel = () => {
  emit('cancel')
}

// 组件挂载时加载数据
loadAvailableTasks()
</script>

<template>
  <div class="space-y-3">
    <div class="space-y-1">
      <Label for="name" class="text-sm">定时消息名称</Label>
      <Input id="name" v-model="localFormData.name" placeholder="请输入定时消息名称" class="h-8" />
    </div>

    <div class="space-y-1">
      <Label for="task_id" class="text-sm">关联发信任务</Label>
      <Select :model-value="localFormData.task_id" @update:model-value="(val) => localFormData.task_id = String(val || '')">
        <SelectTrigger class="h-8">
          <SelectValue placeholder="选择要关联的发信任务" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectItem v-for="task in availableTasks" :key="task.id" :value="String(task.id)">
              {{ task.name }}
            </SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>

    <div class="space-y-1">
      <Label for="cron_expression" class="text-sm">Cron表达式</Label>
      <Input id="cron_expression" v-model="localFormData.cron_expression" placeholder="请输入Cron表达式，如: 0 2 * * *" class="h-8" />
      <div class="text-xs text-gray-500">
        <p class="mb-1">常用模板：</p>
        <div class="flex flex-wrap gap-1">
          <Button v-for="template in cronTemplates" :key="template.value" size="sm" variant="outline"
            @click="applyCronTemplate(template.value)" class="h-6 px-2 text-xs">
            {{ template.label }}
          </Button>
        </div>
      </div>
    </div>

    <div class="space-y-1">
      <Label for="title" class="text-sm">标题</Label>
      <Input id="title" v-model="localFormData.title" placeholder="请输入定时消息标题" class="h-8" />
    </div>

    <div class="space-y-1">
      <Label for="content" class="text-sm">内容</Label>
      <Textarea id="content" v-model="localFormData.content" placeholder="请输入定时消息内容" rows="2" class="min-h-[60px] resize-none" />
    </div>

    <div class="space-y-1">
      <Label for="url" class="text-sm">url（可选）</Label>
      <Input id="url" v-model="localFormData.url" placeholder="请输入定时消息描述" class="h-8" />
    </div>

    <div class="flex justify-end space-x-2 pt-2">
      <Button variant="outline" @click="handleCancel" size="sm" :disabled="loading">
        取消
      </Button>
      <Button @click="handleSubmit" size="sm" :disabled="loading">
        {{ mode === 'add' ? '创建定时消息' : '更新定时消息' }}
      </Button>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'CronMessageForm'
})
</script>