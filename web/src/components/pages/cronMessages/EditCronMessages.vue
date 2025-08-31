<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
import CronMessageForm from './CronMessageForm.vue'

interface CronMessageItem {
  id: string
  name: string
  title: string
  content: string
  cron: string
  url: string
  task_id: string
  enable: number
  status: boolean
}

interface Props {
  open: boolean
  cronMessage: CronMessageItem | null
}

interface Emits {
  (e: 'save', data: any): void
  (e: 'cancel'): void
  (e: 'update:open', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

defineOptions({
  name: 'EditCronMessages'
})

// 表单数据
const formData = reactive({
  name: '',
  cron_expression: '',
  title: '',
  content: '',
  task_id: '',
  url: ''
})


// 加载状态
const loading = ref(false)

// 提交表单
const handleSubmit = async () => {
  if (!props.cronMessage) {
    toast.error('未找到要编辑的定时消息')
    return
  }
  
  
  loading.value = true
  try {
    let postData = {
      "name": formData.name,
      "id": props.cronMessage.id,
      "title": formData.title,
      "content": formData.content,
      "cron": formData.cron_expression,
      "url": formData.url,
      "task_id": formData.task_id,
      "enable": props.cronMessage.enable,
    }

    const rsp = await request.post('/cronmessages/edit', postData)
    if (rsp.data.code === 200) {
      toast.success(rsp.data.msg)
      setTimeout(() => {
        window.location.reload()
      }, 1000)
    } else {
        toast.success(rsp.data.msg)
      }
  } finally {
    loading.value = false
  }
}

// 取消操作
const handleCancel = () => {
  emit('cancel')
  emit('update:open', false)
}

// 监听 cronMessage 变化，更新表单数据
watch(
  () => props.cronMessage,
  (newCronMessage) => {
    if (newCronMessage) {
      formData.name = newCronMessage.name
      formData.cron_expression = newCronMessage.cron
      formData.title = newCronMessage.title
      formData.content = newCronMessage.content
      formData.task_id = newCronMessage.task_id
      formData.url = newCronMessage.url
    }
  },
  { immediate: true }
)
</script>

<template>
  <CronMessageForm
    :model-value="formData"
    @update:model-value="(val) => Object.assign(formData, val)"
    mode="edit"
    :loading="loading"
    @submit="handleSubmit"
    @cancel="handleCancel"
  />
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'EditCronMessages'
})
</script>