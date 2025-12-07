<script setup lang="ts">
import { ref, defineEmits } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { generateBizUniqueID } from '@/util/uuid'
import { request } from '@/api/api'
import { toast } from 'vue-sonner'

// 组件emits
const emit = defineEmits<{
  'save': [data: any]
  'cancel': []
}>()

// 状态管理
const inputValue = ref('')


// 处理取消
const handleCancel = () => {
  inputValue.value = ''
  emit('cancel')
}

// 添加一条任务
const handleSubmit = async () => {
  const taskId = generateBizUniqueID('TK');
  const postData: Record<string, any> = {
    id: taskId,
    name: inputValue.value.trim(),
    ins_data: []
  }
  const rsp = await request.post('/sendtasks/ins/addmany', postData);
  if (await rsp.data.code == 200) {
    toast.success(rsp.data.msg);
    setTimeout(() => {
      window.location.reload();
    }, 1000);

  }
}

</script>

<template>
  <div class="space-y-6 p-6">
    <!-- 输入框区域 -->
    <div class="space-y-2">
      <Label for="task-name">任务名称</Label>
      <Input id="task-name" v-model="inputValue" placeholder="请输入任务名称" @keyup.enter="handleSubmit" class="w-full" />
    </div>

    <!-- 按钮区域 -->
    <div class="flex justify-end gap-2">
      <Button variant="outline" @click="handleCancel">
        取消
      </Button>
      <Button variant="default" class="!bg-gray-800 !hover:bg-gray-900 text-white" @click="handleSubmit"
        :disabled="!inputValue.trim()">
        保存
      </Button>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'AddTasks'
})
</script>