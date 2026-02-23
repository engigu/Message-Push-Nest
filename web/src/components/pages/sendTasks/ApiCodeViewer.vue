<script lang="ts">
import { ref, defineComponent, watch, toRef } from 'vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { ApiStrGenerate } from '@/util/viewApi'
import { useInstanceData } from '@/composables/useInstanceData'
import { useApiCodeViewer } from '@/composables/useApiCodeViewer'

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

    // 使用实例数据管理 composable
    const { hasDynamicRecipientInstance, enabledChannelNames } = useInstanceData(
      'task',
      toRef(props, 'taskData'),
      toRef(props, 'open')
    )

    // 使用 API 代码查看器 composable
    const { activeTab, codeLanguages, copyToClipboard } = useApiCodeViewer()

    // 可选参数选项
    const showHtml = ref(false)
    const showMarkdown = ref(false)
    const showUrl = ref(false)
    const showAtMobiles = ref(false)
    const showAtUserIds = ref(false)
    const showAtAll = ref(false)
    const showRecipients = ref(false)
    
    // 监听动态接收实例变化，自动勾选
    watch(hasDynamicRecipientInstance, (newVal) => {
      if (newVal) {
        showRecipients.value = true
      }
    })
    
    // 监听弹窗关闭，重置状态
    watch(() => props.open, (newVal) => {
      if (!newVal) {
        showRecipients.value = false
      }
    })

    // 生成API代码示例
    const generateApiCode = (language: string) => {
      const taskId = props.taskData?.id || 'TASK_ID'
      const options = { 
        html: showHtml.value, 
        markdown: showMarkdown.value, 
        url: showUrl.value,
        at_mobiles: showAtMobiles.value,
        at_user_ids: showAtUserIds.value,
        at_all: showAtAll.value,
        recipients: showRecipients.value
      }

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

    return {
      handleUpdateOpen,
      activeTab,
      hasDynamicRecipientInstance,
      enabledChannelNames,
      showHtml,
      showMarkdown,
      showUrl,
      showAtMobiles,
      showAtUserIds,
      showAtAll,
      showRecipients,
      codeLanguages,
      generateApiCode,
      copyToClipboard
    }
  }
})
</script>

<template>
  <Dialog :open="open" @update:open="handleUpdateOpen">
    <DialogContent class="w-[800px] sm:w-[900px] lg:w-[1000px] max-w-[90vw] max-h-[90vh] overflow-hidden flex flex-col">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
          <span>API接口</span>
          <Badge v-if="taskData" variant="outline">{{ taskData.name }}</Badge>
        </DialogTitle>
      </DialogHeader>

      <div class="space-y-2 flex-1 overflow-y-auto pr-2">
        <!-- API 信息概览 -->
        <div class="border rounded-lg p-4 space-y-2 bg-white dark:bg-slate-900">
          <div class="flex items-center gap-2">
            <Badge variant="default">POST</Badge>
            <code class="text-sm bg-gray-100 dark:bg-slate-800 px-2 py-1 rounded">/api/v1/message/send</code>
          </div>
          <p class="text-sm text-gray-600 dark:text-gray-400">发送消息到任务配置的渠道</p>
          
          <!-- 已启用的渠道列表 -->
          <div v-if="enabledChannelNames.length > 0" class="mt-3 pt-3 border-t">
            <p class="text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">已启用的发送渠道：</p>
            <div class="flex flex-wrap gap-2">
              <Badge 
                v-for="(name, index) in enabledChannelNames" 
                :key="index" 
                variant="secondary"
                class="text-xs"
              >
                {{ name }}
              </Badge>
            </div>
          </div>
          <div v-else class="mt-3 pt-3 border-t">
            <p class="text-xs text-amber-600 dark:text-amber-400">⚠️ 该任务暂无启用的发送渠道</p>
          </div>
        </div>

        <!-- 可选参数 -->
        <div class="border rounded-lg p-4 space-y-3">
          <h3 class="font-semibold text-sm">可选参数</h3>
          <div class="grid grid-cols-2 md:grid-cols-3 gap-3">
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="showHtml" class="rounded">
              <span class="text-sm">HTML</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="showMarkdown" class="rounded">
              <span class="text-sm">Markdown</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="showUrl" class="rounded">
              <span class="text-sm">URL</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="showAtMobiles" class="rounded">
              <span class="text-sm">@手机号</span>
              <Badge variant="secondary" class="text-xs">新</Badge>
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="showAtUserIds" class="rounded">
              <span class="text-sm">@用户ID</span>
              <Badge variant="secondary" class="text-xs">新</Badge>
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="showAtAll" class="rounded">
              <span class="text-sm">@所有人</span>
              <Badge variant="secondary" class="text-xs">新</Badge>
            </label>
            <label 
              v-if="hasDynamicRecipientInstance" 
              class="flex items-center gap-2 cursor-not-allowed opacity-75"
            >
              <input 
                type="checkbox" 
                v-model="showRecipients" 
                disabled
                class="rounded cursor-not-allowed"
              >
              <span class="text-sm">动态接收者</span>
              <Badge variant="secondary" class="text-xs">必填</Badge>
            </label>
          </div>
          <div class="space-y-1 text-xs text-gray-500 dark:text-gray-400">
            <p>💡 提示：@功能仅钉钉和企业微信支持</p>
            <p v-if="hasDynamicRecipientInstance" class="text-amber-600 dark:text-amber-400">📧 动态接收者：该任务配置了动态接收实例，发送时必须通过API指定接收者列表（此参数已自动勾选且不可取消）</p>
            <p>📋 发送顺序：实例配置的内容类型优先，若为空则按 <code class="bg-gray-100 dark:bg-gray-800 px-1 rounded">HTML → Markdown → Text</code> 顺序回退</p>
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