<script lang="ts">
import { ref, defineComponent, watch, toRef } from 'vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
// @ts-ignore
import { TemplateApiStrGenerate } from '@/util/viewApi'
import { useInstanceData } from '@/composables/useInstanceData'
import { useApiCodeViewer } from '@/composables/useApiCodeViewer'

export default defineComponent({
  name: 'TemplateApiViewer',
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
    templateData: {
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
      'template',
      toRef(props, 'templateData'),
      toRef(props, 'open')
    )

    // 使用 API 代码查看器 composable
    const { activeTab, codeLanguages, copyToClipboard } = useApiCodeViewer()

    // 可选参数选项
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
      const templateId = props.templateData?.id || 'TEMPLATE_ID'
      const placeholders = props.templateData?.placeholders || '[]'
      const options = {
        recipients: showRecipients.value
      }

      switch (language) {
        case 'curl':
          return TemplateApiStrGenerate.getCurlString(templateId, placeholders, options)
        case 'javascript':
          return TemplateApiStrGenerate.getNodeString(templateId, placeholders, options)
        case 'python':
          return TemplateApiStrGenerate.getPythonString(templateId, placeholders, options)
        case 'php':
          return TemplateApiStrGenerate.getPHPString(templateId, placeholders, options)
        case 'golang':
          return TemplateApiStrGenerate.getGolangString(templateId, placeholders, options)
        case 'java':
          return TemplateApiStrGenerate.getJavaString(templateId, placeholders, options)
        case 'rust':
          return TemplateApiStrGenerate.getRustString(templateId, placeholders, options)
        default:
          return '// 请选择一种编程语言查看示例代码'
      }
    }

    return {
      handleUpdateOpen,
      activeTab,
      hasDynamicRecipientInstance,
      enabledChannelNames,
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
    <DialogContent class="w-[800px] sm:w-[900px] lg:w-[1000px] max-w-[90vw] max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
          <span>模板API接口</span>
          <Badge v-if="templateData" variant="outline">{{ templateData.name }}</Badge>
        </DialogTitle>
      </DialogHeader>

      <div class="space-y-4">
        <!-- API 信息概览 -->
        <div class="border rounded-lg p-4 space-y-2 bg-white dark:bg-slate-900">
          <div class="flex items-center gap-2">
            <Badge variant="default">POST</Badge>
            <code class="text-sm bg-gray-100 dark:bg-slate-800 px-2 py-1 rounded">/api/v2/message/send</code>
          </div>
          <p class="text-sm text-gray-600 dark:text-gray-400">使用模板发送消息（V2接口）</p>
          <div class="mt-3 space-y-1 text-xs text-gray-500 dark:text-gray-400">
            <p><strong>模板ID:</strong> <code class="bg-gray-100 dark:bg-slate-800 px-1 py-0.5 rounded">{{ templateData?.id }}</code></p>
            <p><strong>必填参数:</strong> token (加密token), title (消息标题), placeholders (占位符键值对)</p>
            <p><strong>可选参数:</strong> 根据模板配置的@提醒字段自动应用</p>
            <p class="text-amber-600 dark:text-amber-400"><strong>⚠️ 注意:</strong> V2接口使用加密token，不支持明文ID</p>
          </div>
          
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
            <p class="text-xs text-amber-600 dark:text-amber-400">⚠️ 该模板暂无启用的发送渠道</p>
          </div>
        </div>

        <!-- 可选参数 -->
        <div v-if="hasDynamicRecipientInstance" class="border rounded-lg p-4 bg-gray-50 dark:bg-slate-800/50">
          <h3 class="font-semibold mb-3">可选参数</h3>
          <div class="flex flex-wrap gap-4">
            <label class="flex items-center gap-2 cursor-not-allowed opacity-75">
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
          <div class="space-y-1 text-xs text-gray-500 dark:text-gray-400 mt-3">
            <p>📧 动态接收者：该模板配置了动态接收实例，发送时必须通过API指定接收者列表（群发模式）</p>
            <p class="text-amber-600 dark:text-amber-400">⚠️ 此参数已自动勾选且不可取消，因为模板已配置动态接收实例</p>
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
                  class="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-xs leading-relaxed max-w-full whitespace-pre-wrap break-words"><code class="text-xs font-mono">{{ generateApiCode(lang.value) }}</code></pre>
              </div>
            </TabsContent>
          </Tabs>
        </div>

        <!-- 说明 -->
        <div class="border-l-4 border-blue-500 bg-blue-50 dark:bg-blue-950 p-3 rounded text-xs space-y-1">
          <p class="font-semibold text-blue-900 dark:text-blue-200">💡 使用说明</p>
          <ul class="text-blue-800 dark:text-blue-300 space-y-1 ml-4 list-disc">
            <li><strong>token 参数：</strong>需要使用加密后的 token，不能直接使用明文模板ID（安全考虑）</li>
            <li><strong>placeholders 参数：</strong>用于替换模板中的占位符，格式为 <code class="bg-blue-100 dark:bg-blue-900 px-1 rounded">{"key": "value"}</code></li>
            <li>如果模板配置了@提醒，会自动应用到发送的消息中</li>
            <li>支持 Text、HTML、Markdown 三种格式，根据实例配置精确发送对应类型</li>
            <li>系统会自动遍历所有启用的实例进行发送</li>
          </ul>
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
