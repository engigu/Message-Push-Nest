<script setup lang="ts">
import { withDefaults } from 'vue'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import InstanceConfig from '@/components/ui/InstanceConfig.vue'

// 组件props
interface Props {
  open?: boolean
  templateData?: any // 模板数据
}
withDefaults(defineProps<Props>(), {
  open: false,
  templateData: null
})

// 组件emits
defineEmits<{
  'update:open': [value: boolean]
}>()
</script>

<template>
  <Dialog :open="open" @update:open="(value) => $emit('update:open', value)">
    <DialogContent class="w-[500px] max-w-[90vw] max-h-[90vh] overflow-hidden flex flex-col">
      <DialogHeader class="flex-shrink-0">
        <DialogTitle>配置发送实例</DialogTitle>
      </DialogHeader>

      <div class="flex-1 overflow-y-auto">
        <InstanceConfig 
          type="template" 
          :data="templateData" 
          :in-dialog="true"
        />
      </div>
    </DialogContent>
  </Dialog>
</template>
