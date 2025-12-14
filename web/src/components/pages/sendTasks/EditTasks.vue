<script setup lang="ts">
import { defineProps, withDefaults } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import InstanceConfig from '@/components/ui/InstanceConfig.vue'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'

// 组件props
interface Props {
  open?: boolean
  editData?: any // 编辑时传入的数据
}
const props = withDefaults(defineProps<Props>(), {
  open: false,
  editData: null
})

// 定义emits
const emit = defineEmits(['update:open', 'save'])

const handleEditTask = async () => {
  let postData = { id: props.editData.id, name: props.editData.name }
  const rsp = await request.post('/sendtasks/edit', postData)
  if (await rsp.data.code == 200) {
    toast.success(rsp.data.msg)
  }
}
</script>

<template>
  <div class="w-full">
    <!-- 任务信息展示区域 -->
    <div v-if="props.editData" class="flex flex-col sm:flex-row sm:items-center gap-2 border-b p-4">
      <Label class="w-16 sm:w-16">任务名称</Label>
      <Input v-model="props.editData.name" placeholder="请输入任务名称" class="w-full sm:w-64" />
      <Button size="sm" variant="outline" class="w-full sm:w-auto sm:ml-auto" @click="handleEditTask">修改</Button>
    </div>

    <!-- 实例配置组件 -->
    <div class="p-4">
      <InstanceConfig 
        type="task" 
        :data="editData"
      />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'EditTasks'
})
</script>