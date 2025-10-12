<script setup lang="ts">
import { ref, computed } from 'vue'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'

interface Props {
  text: string
  previewTitle?: string
  wrapperClass?: string
}

const props = defineProps<Props>()
const open = ref(false)

const handleClick = () => {
  open.value = true
}

// 生成带可点击URL的安全HTML，保持原文本顺序不变
const escapeHtml = (s: string) =>
  s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')

const linkedHtml = computed(() => {
  const value = (typeof props.text === 'string' ? props.text : '')
  const urlRegex = /(https?:\/\/[\w\-._~:/?#\[\]@!$&'()*+,;=%]+)/g
  let lastIndex = 0
  let html = ''
  let match: RegExpExecArray | null
  while ((match = urlRegex.exec(value)) !== null) {
    const start = match.index
    const url = match[0]
    if (start > lastIndex) {
      html += escapeHtml(value.slice(lastIndex, start))
    }
    // 处理结尾多余的标点，如 ) ] } . , ; : ! ? 等（不属于URL的一部分）
    let trimmed = url
    let suffix = ''
    while (trimmed.length && /[)\]\}\.,;:!?，。；：、！？"'”’]$/.test(trimmed)) {
      const ch = trimmed[trimmed.length - 1]
      if (ch === ')') {
        const opens = (trimmed.match(/\(/g) || []).length
        const closes = (trimmed.match(/\)/g) || []).length
        if (closes < opens) {
          break
        }
      }
      suffix = ch + suffix
      trimmed = trimmed.slice(0, -1)
    }
    const escapedTrimmed = escapeHtml(trimmed)
    html += `<a href="${escapedTrimmed}" target="_blank" rel="noopener noreferrer" class="underline break-all text-blue-600 dark:text-blue-400">${escapedTrimmed}</a>`
    if (suffix) {
      html += escapeHtml(suffix)
    }
    lastIndex = start + url.length
  }
  if (value.length > lastIndex) {
    html += escapeHtml(value.slice(lastIndex))
  }
  return html
})
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
          <pre class="whitespace-pre-wrap break-words break-all text-sm leading-relaxed text-foreground" v-html="linkedHtml"></pre>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>


