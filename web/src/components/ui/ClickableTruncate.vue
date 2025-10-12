<script setup lang="ts">
import { ref } from 'vue'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'

interface Props {
  text: string
  previewTitle?: string
  wrapperClass?: string
}

defineProps<Props>()
const open = ref(false)

const handleClick = () => {
  open.value = true
}
</script>

<template>
  <span
    class="inline-block truncate align-middle cursor-pointer"
    :class="wrapperClass"
    :title="text"
    @click="handleClick"
  >
    {{ text || '-' }}
  </span>

  <Dialog v-model:open="open">
    <DialogContent class="w-[90vw] sm:max-w-xl lg:max-w-2xl">
      <DialogHeader>
        <DialogTitle class="text-sm font-medium text-foreground">{{ previewTitle || '内容' }}</DialogTitle>
      </DialogHeader>
      <div class="mt-1">
        <div class="rounded-lg p-4 bg-muted/40 dark:bg-white/5 ring-1 ring-border/50 shadow-sm max-h-[65vh] overflow-auto overflow-x-hidden">
          <pre class="whitespace-pre-wrap break-words break-all text-sm leading-relaxed text-foreground">{{ text }}</pre>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>


