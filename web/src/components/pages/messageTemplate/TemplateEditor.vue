<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
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
    is_at_all: template.is_at_all || false,
    status: template.status
  }
  
  // 解析占位符
  try {
    placeholdersList.value = JSON.parse(template.placeholders || '[]')
  } catch {
    placeholdersList.value = []
  }
}

// 保存模板
const saveTemplate = async () => {
  if (!formData.value.name.trim()) {
    toast.error('请输入模板名称')
    return
  }

  // 同步占位符数据
  formData.value.placeholders = JSON.stringify(placeholdersList.value)

  try {
    const url = props.isEditing ? '/templates/edit' : '/templates/add'
    await request.post(url, formData.value)
    toast.success(props.isEditing ? '更新模板成功' : '添加模板成功')
    emit('update:open', false)
    emit('saved')
  } catch (error: any) {
    toast.error(error.response?.data?.message || '操作失败')
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
    <DialogContent class="max-w-4xl max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>{{ isEditing ? '编辑模板' : '新建模板' }}</DialogTitle>
      </DialogHeader>
      <div class="space-y-4 py-4">
        <!-- 基本信息 -->
        <div class="space-y-2">
          <Label for="name">模板名称 *</Label>
          <Input id="name" v-model="formData.name" placeholder="请输入模板名称" />
        </div>

        <div class="space-y-2">
          <Label for="description">描述</Label>
          <Textarea id="description" v-model="formData.description" placeholder="请输入模板描述" />
        </div>

        <div class="space-y-2">
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

        <!-- 占位符配置 -->
        <div class="space-y-2">
          <div class="flex justify-between items-center">
            <Label>占位符配置</Label>
            <Button size="sm" variant="outline" @click="addPlaceholder">添加占位符</Button>
          </div>
          <div v-for="(placeholder, index) in placeholdersList" :key="index" class="flex gap-2 items-center">
            <Input
              v-model="placeholder.key"
              placeholder="key (如: username)"
              class="flex-1"
            />
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
          <p class="text-xs text-muted-foreground">
            在模板中使用 <code v-text="'{{key}}'"></code> 来引用占位符，例如：Hello <code v-text="'{{username}}'"></code>
          </p>
        </div>

        <!-- @提醒配置 -->
        <div class="space-y-2">
          <Label>@提醒配置</Label>
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <input 
                type="checkbox" 
                id="is_at_all" 
                v-model="formData.is_at_all"
                class="w-4 h-4 rounded border-gray-300"
              />
              <Label for="is_at_all" class="cursor-pointer">@所有人</Label>
            </div>
            <div class="space-y-1">
              <Label for="at_mobiles">@手机号（多个用逗号分隔）</Label>
              <Input
                id="at_mobiles"
                v-model="formData.at_mobiles"
                placeholder="例如：13800138000,13900139000"
              />
            </div>
            <div class="space-y-1">
              <Label for="at_user_ids">@用户ID（多个用逗号分隔）</Label>
              <Input
                id="at_user_ids"
                v-model="formData.at_user_ids"
                placeholder="例如：user001,user002"
              />
            </div>
          </div>
          <p class="text-xs text-muted-foreground">
            配置后，使用此模板发送消息时会自动@指定的用户（适用于支持@功能的渠道，如钉钉、企业微信等）
          </p>
        </div>

        <!-- 模板内容 -->
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
              rows="10"
            />
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
              rows="10"
            />
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
              rows="10"
            />
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
