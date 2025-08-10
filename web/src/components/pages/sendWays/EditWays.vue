<script setup lang="ts">
import { defineEmits, defineProps, withDefaults } from 'vue'
import WaysForm from './WaysForm.vue'

// 组件props
interface Props {
  open?: boolean
  editData?: any // 编辑时传入的数据
}
withDefaults(defineProps<Props>(), {
  open: false,
  editData: null
})

// 组件emits
const emit = defineEmits<{
  'update:open': [value: boolean]
  'save': [data: any]
}>()

// 处理保存事件
const handleSave = (data: any) => {
  emit('save', data)
}

// 处理关闭事件
const handleUpdateOpen = (value: boolean) => {
  emit('update:open', value)
}
</script>

<template>
  <WaysForm 
    :open="open" 
    :edit-data="editData"
    mode="edit" 
    @update:open="handleUpdateOpen" 
    @save="handleSave" 
  />
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'EditWays'
})
</script>