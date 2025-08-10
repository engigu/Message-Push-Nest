<script lang="ts">
import { ref, computed, defineComponent } from 'vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
// @ts-ignore
import { ApiStrGenerate } from '@/util/viewApi'

export default defineComponent({
  name: 'ApiCodeViewer',
  components: {
    Button,
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
    Badge
  },
  props: {
    open: {
      type: Boolean,
      default: false
    },
    taskData: {
      type: Object,
      default: null
    }
  },
  emits: ['update:open'],
  setup(props, { emit }) {
    // 处理关闭事件
    const handleUpdateOpen = (value: boolean) => {
      emit('update:open', value)
    }

    // 当前选中的标签
    const activeTab = ref('curl')

    // 代码语言选项
    const codeLanguages = [
      { value: 'curl', label: 'cURL', icon: '🌐' },
      { value: 'javascript', label: 'JS', icon: '🟨' },
      { value: 'python', label: 'Python', icon: '🐍' },
      { value: 'php', label: 'PHP', icon: '🐘' },
      { value: 'golang', label: 'Go', icon: '🐹' },
      { value: 'java', label: 'Java', icon: '☕' },
      { value: 'rust', label: 'Rust', icon: '🦀' }
    ]

    // 生成API代码示例
    const generateApiCode = (language: string) => {
      const taskId = props.taskData?.id || 'TASK_ID'
      const options = { html: false, markdown: false, url: false }

      switch (language) {
        case 'curl':
          return ApiStrGenerate.getCurlString(taskId, options)
        case 'javascript':
          return ApiStrGenerate.getNodeString(taskId, options)
        case 'python':
          return ApiStrGenerate.getPythonString(taskId, options)
        case 'php':
          return ApiStrGenerate.getPHPString(taskId, options)
        case 'golang':
          return ApiStrGenerate.getGolangString(taskId, options)
        case 'java':
          return ApiStrGenerate.getJaveString(taskId, options)
        case 'rust':
          return ApiStrGenerate.getRustString(taskId, options)
        default:
          return '// 请选择一种编程语言查看示例代码'
      }
    }

    // 复制代码到剪贴板
    const copyToClipboard = async (text: string) => {
      try {
        await navigator.clipboard.writeText(text)
        // 这里可以添加成功提示
      } catch (err) {
        console.error('复制失败:', err)
      }
    }

    return {
      handleUpdateOpen,
      activeTab,
      codeLanguages,
      generateApiCode,
      copyToClipboard
    }
  }
})
</script>

<template>
  <Dialog :open="open" @update:open="handleUpdateOpen">
    <DialogContent class="w-[800px] sm:w-[900px] lg:w-[1000px] max-w-[90vw] max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
          <span>API接口</span>
          <Badge v-if="taskData" variant="outline">{{ taskData.name }}</Badge>
        </DialogTitle>
      </DialogHeader>

      <div class="mt-6 space-y-4">
        <!-- API 信息概览 -->

        <div class="space-y-3">
          <div class="border rounded-lg p-4">
            <div class="flex items-center gap-2 mb-2">
              <Badge variant="default">POST</Badge>
              <code class="text-sm">/sendtasks/send</code>
            </div>
            <p class="text-sm text-gray-600">发送消息，创建新的消息</p>
          </div>
        </div>

        <!-- 代码示例 -->
        <div class="space-y-4">
          <h3 class="font-semibold">代码示例</h3>

          <Tabs v-model="activeTab" class="w-full">
            <TabsList class="grid w-full grid-cols-7 gap-1">
              <TabsTrigger v-for="lang in codeLanguages" :key="lang.value" :value="lang.value"
                class="flex items-center gap-1 px-2 py-1 text-xs">
                <span>{{ lang.icon }}</span>
                <span class="hidden sm:inline">{{ lang.label }}</span>
                <span class="sm:hidden">{{ lang.label.slice(0, 3) }}</span>
              </TabsTrigger>
            </TabsList>

            <TabsContent v-for="lang in codeLanguages" :key="lang.value" :value="lang.value" class="mt-4">
              <div class="relative">
                <Button size="sm" variant="outline" class="absolute top-2 right-2 z-10"
                  @click="copyToClipboard(generateApiCode(lang.value))">
                  复制代码
                </Button>
                <pre
                  class="bg-gray-900 text-gray-100 p-3 rounded-lg overflow-x-auto text-xs leading-tight max-w-full whitespace-pre-wrap break-words"><code class="text-xs font-mono">{{ generateApiCode(lang.value) }}</code></pre>
              </div>
            </TabsContent>
          </Tabs>
        </div>


      </div>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
/* 代码块样式优化 */
pre {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

code {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}
</style>