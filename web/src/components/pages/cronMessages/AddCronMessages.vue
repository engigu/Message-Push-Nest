<script setup lang="ts">
import { ref, reactive } from 'vue'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
import { generateBizUniqueID } from '@/util/uuid'
import CronMessageForm from './CronMessageForm.vue'

interface Props {
  open: boolean
}

interface Emits {
  (e: 'save', data: any): void
  (e: 'cancel'): void
  (e: 'update:open', value: boolean): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()

defineOptions({
  name: 'AddCronMessages'
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
  loading.value = true
  try {
    let postData = {
      "name": formData.name,
      "id": generateBizUniqueID("CM"),
      "title": formData.title,
      "content": formData.content,
      "cron": formData.cron_expression,
      "url": formData.url,
      "task_id": formData.task_id
    }

    const rsp = await request.post('/cronmessages/addone', postData)
    if (rsp.data.code === 200) {
      toast.success(rsp.data.msg)
      setTimeout(() => {
        window.location.reload()
      }, 1000)
    } else {
      toast.error(rsp.data.msg)
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

// 立即发送（新增模式也支持，可以在创建前测试发送效果）
const handleSendNow = async () => {
  // 验证必填字段
  if (!formData.task_id) {
    toast.error('请先选择关联的发信任务')
    return
  }
  if (!formData.title) {
    toast.error('请先填写消息标题')
    return
  }
  if (!formData.content) {
    toast.error('请先填写消息内容')
    return
  }

  loading.value = true
  try {
    const postData = {
      task_id: formData.task_id,
      title: formData.title,
      content: formData.content,
      url: formData.url
    }

    const rsp = await request.post('/cronmessages/sendnow', postData)
    if (rsp.data.code === 200) {
      toast.success(rsp.data.msg)
    } else {
      toast.error(rsp.data.msg)
    }
  } catch (error) {
    toast.error('发送失败，请稍后重试')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <CronMessageForm
    :model-value="formData"
    @update:model-value="(val) => {
      console.log('Received update:model-value:', val);
      Object.assign(formData, val);
    }"
    mode="add"
    :loading="loading"
    @submit="handleSubmit"
    @cancel="handleCancel"
    @send-now="handleSendNow"
  />
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'AddCronMessages'
})
</script>