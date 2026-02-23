<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'

interface Placeholder {
  key: string
  label: string
  default: string
}

interface TemplateData {
  id?: string  // 模板ID是字符串类型（UUID）
  name: string
  description: string
  text_template: string
  html_template: string
  markdown_template: string
  placeholders: string
  at_mobiles?: string
  at_user_ids?: string
  is_at_all?: boolean
  status: string
}

// 组件props
interface Props {
  open?: boolean
  isEditing?: boolean
  templateData?: TemplateData | null
}
const props = withDefaults(defineProps<Props>(), {
  open: false,
  isEditing: false,
  templateData: null
})

// 组件emits
const emit = defineEmits<{
  'update:open': [value: boolean]
  'saved': []
}>()

// Textarea refs for inserting placeholders
const textTemplateRef = ref<any>(null)
const htmlTemplateRef = ref<any>(null)
const markdownTemplateRef = ref<any>(null)

// 表单数据
const formData = ref<TemplateData>({
  name: '',
  description: '',
  text_template: '',
  html_template: '',
  markdown_template: '',
  placeholders: '[]',
  at_mobiles: '',
  at_user_ids: '',
  is_at_all: false,
  status: 'enabled'
})

// 使用独立的响应式数组来管理占位符，避免频繁的 JSON 序列化
const placeholdersList = ref<Placeholder[]>([])

// 过滤出有效的占位符（key 不为空）
const validPlaceholders = computed(() => {
  return placeholdersList.value.filter(ph => ph.key && ph.key.trim())
})

// 预览数据
const previewData = ref({
  text: '',
  html: '',
  markdown: '',
  params: {} as Record<string, string>
})

// 是否显示预览
const showPreview = ref(false)

// 预览防抖定时器
let previewDebounceTimer: number | null = null

// 刷新预览
const refreshPreview = async () => {
  if (!props.isEditing || !formData.value.id) {
    // 新建模板时，直接使用当前输入的内容作为预览
    previewData.value.text = replacePreviewPlaceholders(formData.value.text_template)
    previewData.value.html = replacePreviewPlaceholders(formData.value.html_template)
    previewData.value.markdown = replacePreviewPlaceholders(formData.value.markdown_template)
    return
  }

  try {
    const rsp = await request.post('/templates/preview', {
      id: formData.value.id,
      params: previewData.value.params
    })
    previewData.value.text = rsp.data.data.text || ''
    previewData.value.html = rsp.data.data.html || ''
    previewData.value.markdown = rsp.data.data.markdown || ''
  } catch (error: any) {
    console.error('预览失败:', error)
  }
}

// 替换预览占位符（用于新建模板）
const replacePreviewPlaceholders = (template: string) => {
  if (!template) return ''
  let result = template
  Object.keys(previewData.value.params).forEach(key => {
    const value = previewData.value.params[key] || `{{${key}}}`
    result = result.replace(new RegExp(`{{${key}}}`, 'g'), value)
  })
  return result
}

// 监听模板内容变化，自动刷新预览（防抖）
watch([
  () => formData.value.text_template,
  () => formData.value.html_template,
  () => formData.value.markdown_template,
  () => previewData.value.params
], () => {
  if (!showPreview.value) return
  
  if (previewDebounceTimer) {
    clearTimeout(previewDebounceTimer)
  }
  previewDebounceTimer = window.setTimeout(() => {
    refreshPreview()
  }, 500)
}, { deep: true })

// 监听占位符列表变化，同步到 formData（使用防抖）
let placeholderDebounceTimer: number | null = null
watch(placeholdersList, () => {
  if (placeholderDebounceTimer) {
    clearTimeout(placeholderDebounceTimer)
  }
  placeholderDebounceTimer = window.setTimeout(() => {
    formData.value.placeholders = JSON.stringify(placeholdersList.value)
  }, 300)
}, { deep: true })

// 检查占位符 key 是否重复
const isDuplicateKey = (key: string, currentIndex: number): boolean => {
  if (!key.trim()) return false
  return placeholdersList.value.some((p, index) => 
    index !== currentIndex && p.key.trim() === key.trim()
  )
}

// 获取重复的 key 列表
const getDuplicateKeys = computed(() => {
  const keys = placeholdersList.value.map(p => p.key.trim()).filter(k => k)
  const duplicates = new Set<string>()
  const seen = new Set<string>()
  
  keys.forEach(key => {
    if (seen.has(key)) {
      duplicates.add(key)
    }
    seen.add(key)
  })
  
  return duplicates
})

// 添加占位符
const addPlaceholder = () => {
  placeholdersList.value.push({ key: '', label: '', default: '' })
}

// 删除占位符
const removePlaceholder = (index: number) => {
  placeholdersList.value.splice(index, 1)
}

// 插入占位符到模板
const insertPlaceholder = async (type: 'text' | 'html' | 'markdown', key: string) => {
  const placeholder = `{{${key}}}`
  let targetRef: any = null
  
  if (type === 'text') targetRef = textTemplateRef.value
  else if (type === 'html') targetRef = htmlTemplateRef.value
  else if (type === 'markdown') targetRef = markdownTemplateRef.value
  
  if (!targetRef) return
  
  await nextTick()
  
  const textarea = targetRef.$el || targetRef
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const text = formData.value[`${type}_template`]
  
  const before = text.substring(0, start)
  const after = text.substring(end)
  
  formData.value[`${type}_template`] = before + placeholder + after
  
  await nextTick()
  textarea.focus()
  const newPosition = start + placeholder.length
  textarea.setSelectionRange(newPosition, newPosition)
}

// 重置表单
const resetForm = () => {
  formData.value = {
    name: '',
    description: '',
    text_template: '',
    html_template: '',
    markdown_template: '',
    placeholders: '[]',
    at_mobiles: '',
    at_user_ids: '',
    is_at_all: false,
    status: 'enabled'
  }
  placeholdersList.value = []
}

// 加载模板数据
const loadTemplateData = (template: TemplateData) => {
  formData.value = {
    id: template.id,
    name: template.name,
    description: template.description,
    text_template: template.text_template,
    html_template: template.html_template,
    markdown_template: template.markdown_template,
    placeholders: template.placeholders,
    at_mobiles: template.at_mobiles || '',
    at_user_ids: template.at_user_ids || '',
    is_at_all: Boolean(template.is_at_all),
    status: template.status
  }
  
  // 解析占位符
  try {
    placeholdersList.value = JSON.parse(template.placeholders || '[]')
  } catch {
    placeholdersList.value = []
  }
  
  // 初始化预览参数
  previewData.value.params = {}
  placeholdersList.value.forEach(p => {
    previewData.value.params[p.key] = p.default || ''
  })
}

// 保存模板
const saveTemplate = async () => {
  if (!formData.value.name.trim()) {
    toast.error('请输入模板名称')
    return
  }
  
  // 验证至少填写一种格式的模板内容
  if (!formData.value.text_template && !formData.value.html_template && !formData.value.markdown_template) {
    toast.error('至少需要填写一种格式的模板内容')
    return
  }
  
  // 验证占位符 key 不能为空且不能重复
  const emptyKeys = placeholdersList.value.filter(p => p.key.trim() === '')
  if (emptyKeys.length > 0) {
    toast.error('占位符 key 不能为空')
    return
  }
  
  if (getDuplicateKeys.value.size > 0) {
    const duplicates = Array.from(getDuplicateKeys.value).join('、')
    toast.error(`占位符 key 不能重复：${duplicates}`)
    return
  }

  // 同步占位符数据
  formData.value.placeholders = JSON.stringify(placeholdersList.value)

  try {
    const url = props.isEditing ? '/templates/edit' : '/templates/add'
    const response = await request.post(url, formData.value)
    if (response.data.code === 200) {
      toast.success(props.isEditing ? '更新模板成功' : '添加模板成功')
      emit('update:open', false)
      emit('saved')
    } else {
      toast.error(response.data.msg || '操作失败')
    }
  } catch (error: any) {
    toast.error(error.response?.data?.msg || error.response?.data?.message || '操作失败')
  }
}

// 监听对话框打开状态
watch(() => props.open, (newVal) => {
  if (newVal) {
    if (props.isEditing && props.templateData) {
      loadTemplateData(props.templateData)
    } else {
      resetForm()
    }
  }
})
</script>

<template>
  <Dialog :open="open" @update:open="(value) => $emit('update:open', value)">
    <DialogContent class="max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
      <DialogHeader>
        <DialogTitle>{{ isEditing ? '编辑模板' : '新建模板' }}</DialogTitle>
      </DialogHeader>
      <div class="space-y-4 py-4 flex-1 overflow-y-auto pr-2">
        <!-- 基本信息 -->
        <div class="grid grid-cols-10 gap-4">
          <div class="col-span-7 space-y-2">
            <Label for="name">模板名称 *</Label>
            <Input id="name" v-model="formData.name" placeholder="请输入模板名称" />
          </div>
          <div class="col-span-3 space-y-2">
            <Label>状态</Label>
            <Select v-model="formData.status">
              <SelectTrigger class="w-full">
                <SelectValue placeholder="全部" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="enabled">启用</SelectItem>
                  <SelectItem value="disabled">禁用</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div class="space-y-2">
          <Label for="description">描述</Label>
          <Textarea id="description" v-model="formData.description" placeholder="请输入模板描述" />
        </div>

        <!-- 占位符配置 -->
        <div class="space-y-2">
          <div class="flex justify-between items-center">
            <Label>占位符配置</Label>
            <Button size="sm" variant="outline" @click="addPlaceholder">添加占位符</Button>
          </div>
          <div v-for="(placeholder, index) in placeholdersList" :key="index" class="flex flex-col gap-1">
            <div class="flex gap-2 items-center">
              <div class="flex-1 relative">
                <Input
                  v-model="placeholder.key"
                  placeholder="key (如: username)"
                  :class="{ 'border-red-500 focus-visible:ring-red-500': isDuplicateKey(placeholder.key, index) }"
                />
                <p v-if="isDuplicateKey(placeholder.key, index)" class="text-xs text-red-500 mt-1">
                  该 key 已存在
                </p>
              </div>
              <Input
                v-model="placeholder.label"
                placeholder="标签 (如: 用户名)"
                class="flex-1"
              />
              <Input
                v-model="placeholder.default"
                placeholder="默认值"
                class="flex-1"
              />
              <Button size="sm" variant="ghost" @click="removePlaceholder(index)">删除</Button>
            </div>
          </div>
          <p class="text-xs text-muted-foreground">
            在模板中使用 <code v-text="'{{key}}'"></code> 来引用占位符，例如：Hello <code v-text="'{{username}}'"></code>
          </p>
        </div>

        <!-- @提醒配置 -->
        <div class="space-y-2">
          <Label>@提醒配置 <span class="text-xs text-muted-foreground font-normal">（适用于钉钉、企业微信等）</span></Label>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-2">
            <div class="flex items-center gap-2 px-3 py-2 border rounded-md">
              <Checkbox 
                id="is_at_all" 
                :model-value="formData.is_at_all"
                @update:model-value="(newVal: boolean | 'indeterminate') => formData.is_at_all = newVal === true"
              />
              <Label for="is_at_all" class="cursor-pointer text-sm">@所有人</Label>
            </div>
            <Input
              v-model="formData.at_mobiles"
              placeholder="@手机号（逗号分隔）"
              class="text-sm"
            />
            <Input
              v-model="formData.at_user_ids"
              placeholder="@用户ID（逗号分隔）"
              class="text-sm"
            />
          </div>
        </div>

        <!-- 模板内容 -->
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <Label class="text-base font-semibold">模板内容</Label>
            <Button 
              size="sm" 
              variant="outline" 
              @click="showPreview = !showPreview; if (showPreview) refreshPreview()"
            >
              {{ showPreview ? '隐藏预览' : '显示预览' }}
            </Button>
          </div>
          
          <!-- 占位符参数输入（仅在显示预览时） -->
          <div v-if="showPreview && validPlaceholders.length > 0" class="p-3 bg-muted rounded-lg space-y-2">
            <Label class="text-sm font-medium">填写占位符参数（用于预览）</Label>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
              <div v-for="ph in validPlaceholders" :key="ph.key" class="flex gap-2 items-center">
                <Label class="text-xs w-24 flex-shrink-0">{{ ph.key }}</Label>
                <Input
                  v-model="previewData.params[ph.key]"
                  :placeholder="ph.default || `请输入 ${ph.key}`"
                  class="text-sm h-8"
                  size="sm"
                />
              </div>
            </div>
          </div>
        </div>
        
        <Tabs default-value="text" class="w-full">
          <TabsList class="grid w-full grid-cols-3">
            <TabsTrigger value="text">Text</TabsTrigger>
            <TabsTrigger value="html">HTML</TabsTrigger>
            <TabsTrigger value="markdown">Markdown</TabsTrigger>
          </TabsList>
          <TabsContent value="text" class="space-y-2">
            <div class="flex flex-col gap-1">
              <Label>纯文本模板</Label>
              <div v-if="validPlaceholders.length > 0" class="flex gap-1 overflow-x-auto pb-1 scrollbar-thin">
                <Button
                  v-for="ph in validPlaceholders"
                  :key="ph.key"
                  size="sm"
                  variant="outline"
                  class="h-7 text-xs whitespace-nowrap flex-shrink-0"
                  @click="insertPlaceholder('text', ph.key)"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                  </svg>
                  {{ph.key}}
                </Button>
              </div>
            </div>
            <Textarea
              ref="textTemplateRef"
              v-model="formData.text_template"
              placeholder="请输入纯文本模板内容，可使用 {{key}} 作为占位符"
              :rows="showPreview ? 10 : 15"
            />
            
            <!-- 预览区 -->
            <div v-if="showPreview" class="space-y-2">
              <Label>预览效果</Label>
              <div class="p-4 border rounded-md bg-muted/50">
                <pre class="whitespace-pre-wrap text-sm">{{ previewData.text || '无内容' }}</pre>
              </div>
            </div>
          </TabsContent>
          <TabsContent value="html" class="space-y-2">
            <div class="flex flex-col gap-1">
              <Label>HTML模板</Label>
              <div v-if="validPlaceholders.length > 0" class="flex gap-1 overflow-x-auto pb-1 scrollbar-thin">
                <Button
                  v-for="ph in validPlaceholders"
                  :key="ph.key"
                  size="sm"
                  variant="outline"
                  class="h-7 text-xs whitespace-nowrap flex-shrink-0"
                  @click="insertPlaceholder('html', ph.key)"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                  </svg>
                  {{ph.key}}
                </Button>
              </div>
            </div>
            <Textarea
              ref="htmlTemplateRef"
              v-model="formData.html_template"
              placeholder="请输入HTML模板内容，可使用 {{key}} 作为占位符"
              :rows="showPreview ? 10 : 15"
            />
            
            <!-- 预览区 -->
            <div v-if="showPreview" class="space-y-2">
              <Label>预览效果（基础渲染）</Label>
              <div class="p-4 border rounded-md bg-muted/50">
                <div v-html="previewData.html || '无内容'"></div>
              </div>
              <p class="text-xs text-muted-foreground">
                💡 HTML 预览仅显示基础结构，实际发送时可能包含邮件样式等
              </p>
            </div>
          </TabsContent>
          <TabsContent value="markdown" class="space-y-2">
            <div class="flex flex-col gap-1">
              <Label>Markdown模板</Label>
              <div v-if="validPlaceholders.length > 0" class="flex gap-1 overflow-x-auto pb-1 scrollbar-thin">
                <Button
                  v-for="ph in validPlaceholders"
                  :key="ph.key"
                  size="sm"
                  variant="outline"
                  class="h-7 text-xs whitespace-nowrap flex-shrink-0"
                  @click="insertPlaceholder('markdown', ph.key)"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                  </svg>
                  {{ph.key}}
                </Button>
              </div>
            </div>
            <Textarea
              ref="markdownTemplateRef"
              v-model="formData.markdown_template"
              placeholder="请输入Markdown模板内容，可使用 {{key}} 作为占位符"
              :rows="showPreview ? 10 : 15"
            />
            
            <!-- 预览区 -->
            <div v-if="showPreview" class="space-y-2">
              <Label>预览效果（原始格式）</Label>
              <div class="p-4 border rounded-md bg-muted/50">
                <pre class="whitespace-pre-wrap text-sm">{{ previewData.markdown || '无内容' }}</pre>
              </div>
              <p class="text-xs text-muted-foreground">
                💡 Markdown 在发送时会被渲染为对应格式（钉钉、企业微信等平台支持）
              </p>
            </div>
          </TabsContent>
        </Tabs>
      </div>
      <DialogFooter>
        <Button variant="outline" @click="$emit('update:open', false)">取消</Button>
        <Button @click="saveTemplate">保存</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
