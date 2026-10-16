<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
import { CONSTANT } from '@/constant'
import {
  Search,
  ArrowLeft,
  Loader2,
  Info,
  Send,
  Smartphone,
  MessageSquare,
  Building2,
  AppWindow,
  Mail,
  Webhook,
  Radio,
  Layers,
  Bell,
  PhoneCall,
  QrCode,
  Server,
} from 'lucide-vue-next'

// 组件 props
interface Props {
  open?: boolean
  editData?: any // 编辑时传入的数据
  mode?: 'add' | 'edit' // 模式：新增或编辑
}

const props = withDefaults(defineProps<Props>(), {
  open: false,
  editData: null,
  mode: 'add'
})

// 组件 emits
const emit = defineEmits<{
  'update:open': [value: boolean]
  'save': [data: any]
}>()

// 核心数据源直接从 CONSTANT 读取
const waysData = CONSTANT.WAYS_DATA
const channelCategories = CONSTANT.CHANNEL_CATEGORIES || [
  { id: 'all', label: '全部' },
  { id: 'workplace', label: '办公协作' },
  { id: 'push', label: '移动推送' },
  { id: 'base', label: '基础服务' },
]

// 当前步骤：'select' (选择渠道) | 'configure' (配置渠道)
const currentStep = ref<'select' | 'configure'>(props.mode === 'edit' ? 'configure' : 'select')

// 搜索关键字
const searchQuery = ref('')

// 当前选中的分类
const activeCategory = ref<string>('all')

// 当前选中的渠道类型，默认从 waysData 第一个开始
const selectedType = ref<string>(waysData[0]?.type || 'Telegram')



// 表单数据
const formData = ref<Record<string, any>>({})

// 表单校验错误记录
const validationErrors = ref<Record<string, string>>({})

// 操作加载中状态
const isTesting = ref(false)
const isSaving = ref(false)

// 当前选中渠道的完整配置信息 (从 CONSTANT.WAYS_DATA 获取)
const currentChannel = computed(() => {
  const found = waysData.find(item => item.type === selectedType.value)
  return found || waysData[0]
})

// 当前渠道除去 name 以外的参数列表
const channelParamInputs = computed(() => {
  const inputs = currentChannel.value?.inputs || []
  return inputs.filter((item: any) => item.col !== 'name')
})

// 根据 CONSTANT.WAYS_DATA 进行分类和搜索过滤
const filteredChannels = computed(() => {
  return waysData.filter(item => {
    const categoryId = item.meta?.categoryId || item.categoryId
    const category = item.meta?.category || item.category || ''
    const subTitle = item.meta?.subTitle || item.subTitle || ''
    const label = item.label || ''
    const type = item.type || ''

    // 分类匹配
    if (activeCategory.value !== 'all' && categoryId !== activeCategory.value) {
      return false
    }

    // 搜索匹配
    const query = searchQuery.value.trim().toLowerCase()
    if (!query) return true

    const matchLabel = label.toLowerCase().includes(query)
    const matchSubTitle = subTitle.toLowerCase().includes(query)
    const matchType = type.toLowerCase().includes(query)
    const matchCategory = category.toLowerCase().includes(query)

    return matchLabel || matchSubTitle || matchType || matchCategory
  })
})

// 渠道图标与主题色样式映射
const getChannelVisual = (type: string) => {
  const visualMap: Record<string, { icon: any; iconBg: string }> = {
    Telegram: { icon: Send, iconBg: 'bg-[#229ED9]/15 text-[#229ED9]' },
    Bark: { icon: Smartphone, iconBg: 'bg-[#007AFF]/15 text-[#007AFF]' },
    Dtalk: { icon: MessageSquare, iconBg: 'bg-[#007FFF]/15 text-[#007FFF]' },
    QyWeiXin: { icon: Building2, iconBg: 'bg-[#2BA245]/15 text-[#2BA245]' },
    QyWeiXinApp: { icon: AppWindow, iconBg: 'bg-[#07C160]/15 text-[#07C160]' },
    Feishu: { icon: Send, iconBg: 'bg-[#00BFA5]/15 text-[#00BFA5]' },
    Email: { icon: Mail, iconBg: 'bg-[#EA4335]/15 text-[#EA4335]' },
    Custom: { icon: Webhook, iconBg: 'bg-[#8B5CF6]/15 text-[#8B5CF6]' },
    Ntfy: { icon: Radio, iconBg: 'bg-[#E11D48]/15 text-[#E11D48]' },
    Gotify: { icon: Layers, iconBg: 'bg-[#3B82F6]/15 text-[#3B82F6]' },
    PushMe: { icon: Bell, iconBg: 'bg-[#F59E0B]/15 text-[#F59E0B]' },
    AliyunSMS: { icon: PhoneCall, iconBg: 'bg-[#FF6A00]/15 text-[#FF6A00]' },
    WeChatOFAccount: { icon: QrCode, iconBg: 'bg-[#07C160]/15 text-[#07C160]' },
    MessageNest: { icon: Server, iconBg: 'bg-[#6366F1]/15 text-[#6366F1]' },
  }
  return visualMap[type] || { icon: Send, iconBg: 'bg-primary/10 text-primary' }
}

// 初始化表单数据
const initFormData = () => {
  const config = currentChannel.value
  if (!config) return

  const newFormData: Record<string, any> = {}

  // 编辑模式填充数据
  if (props.mode === 'edit' && props.editData) {
    selectedType.value = props.editData.type || waysData[0]?.type || ''
    currentStep.value = 'configure'

    let authData: Record<string, any> = {}
    try {
      authData = props.editData.auth ? JSON.parse(props.editData.auth) : {}
    } catch (e) {
      console.error('解析 auth 失败:', e)
    }

    newFormData.name = props.editData.name || ''

    Object.keys(authData).forEach(key => {
      newFormData[key] = authData[key]
    })
  } else {
    // 新增模式：默认使用渠道名称
    if (!newFormData.name) {
      newFormData.name = config.label || ''
    }
  }

  // 初始化各输入参数的默认值
  if (config.inputs) {
    config.inputs.forEach((input: any) => {
      if (newFormData[input.col] === undefined) {
        newFormData[input.col] = input.value || ''
      }
    })
  }

  formData.value = newFormData
  validationErrors.value = {}
}

// 选择渠道卡片，进入配置页
const handleSelectChannel = (channel: any) => {
  selectedType.value = channel.type
  initFormData()
  currentStep.value = 'configure'
}

// 返回渠道选择
const handleBackToSelect = () => {
  currentStep.value = 'select'
  validationErrors.value = {}
}

// 清除单个字段校验错误
const clearFieldError = (col: string) => {
  if (validationErrors.value[col]) {
    delete validationErrors.value[col]
  }
}

// 获取最终提交数据
const getFinalData = () => {
  const config = currentChannel.value
  const authData: Record<string, any> = {}

  if (config.inputs) {
    config.inputs.forEach((input: any) => {
      if (formData.value[input.col] !== undefined && input.col !== 'name') {
        let val = formData.value[input.col]
        if (config.type === 'Email' && input.col === 'port') {
          val = parseInt(val) || 0
        }
        if (config.type === 'Gotify' && input.col === 'priority') {
          val = parseInt(val) || 5
        }
        authData[input.col] = val
      }
    })
  }

  const postData: Record<string, any> = {
    auth: JSON.stringify(authData),
    type: selectedType.value,
    name: formData.value.name ? formData.value.name.trim() : config.label,
  }

  if (props.mode === 'edit' && props.editData?.id) {
    postData.id = props.editData.id
  }

  return postData
}

// 表单校验
const validateForm = (): boolean => {
  const errors: Record<string, string> = {}
  const config = currentChannel.value

  if (!formData.value.name || !formData.value.name.trim()) {
    errors.name = '请输入渠道名称'
  }

  if (config.inputs) {
    config.inputs.forEach((input: any) => {
      if (input.required && input.col !== 'name') {
        const val = formData.value[input.col]
        if (val === undefined || val === null || String(val).trim() === '') {
          errors[input.col] = `请输入 ${input.subLabel || input.label}`
        }
      }
    })
  }

  validationErrors.value = errors
  return Object.keys(errors).length === 0
}

// 测试连通性
const handleTest = async () => {
  if (!validateForm()) {
    toast.error('请先完善必填配置项')
    return
  }

  isTesting.value = true
  try {
    const postData = getFinalData()
    const rsp = await request.post('/sendways/test', postData)
    if (rsp.data && rsp.data.code === 200) {
      toast.success(rsp.data.msg || '渠道测试成功')
    } else {
      toast.error(rsp.data?.msg || '渠道测试失败')
    }
  } catch (error: any) {
    toast.error(error.message || '测试请求出错')
  } finally {
    isTesting.value = false
  }
}

// 保存提交
const handleSave = async () => {
  if (!validateForm()) {
    toast.error('请先完善必填配置项')
    return
  }

  isSaving.value = true
  try {
    const postData = getFinalData()
    const apiUrl = props.mode === 'edit' ? '/sendways/edit' : '/sendways/add'
    const successMessage = props.mode === 'edit' ? '更新渠道成功！' : '添加渠道成功！'

    const rsp = await request.post(apiUrl, postData)
    if (rsp.data && rsp.data.code === 200) {
      toast.success(successMessage)
      emit('save', postData)
      emit('update:open', false)
      setTimeout(() => {
        window.location.reload()
      }, 800)
    } else {
      toast.error(rsp.data?.msg || '保存失败，请检查参数')
    }
  } catch (error: any) {
    toast.error(error.message || '提交保存异常')
  } finally {
    isSaving.value = false
  }
}

// 关闭弹窗
const handleClose = () => {
  emit('update:open', false)
}

// 监听编辑数据或打开状态传入
watch(
  () => [props.editData, props.open],
  () => {
    if (props.open) {
      initFormData()
    }
  },
  { immediate: true }
)
</script>

<template>
  <div class="w-full flex flex-col h-full select-none">
    <!-- 弹窗 Header 区域 -->
    <div class="px-6 pt-6 pb-4 border-b border-border/40 shrink-0 pr-12">
      <DialogHeader>
        <DialogTitle class="text-lg font-bold text-foreground tracking-tight">
          {{ currentStep === 'select' ? '选择通知渠道' : '配置通知渠道' }}
        </DialogTitle>
        <DialogDescription class="text-xs text-muted-foreground mt-1">
          {{
            currentStep === 'select'
              ? '点击选择您需要接入的消息推送服务'
              : '配置推送参数并绑定消息通知'
          }}
        </DialogDescription>
      </DialogHeader>
    </div>

    <!-- 步骤一：选择通知渠道 (完全参考设计图1) -->
    <div
      v-if="currentStep === 'select'"
      class="flex-1 overflow-y-auto px-6 py-5 flex flex-col min-h-0"
    >
      <!-- 搜索与分类筛选栏 -->
      <div class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 mb-4 shrink-0">
        <!-- 搜索输入框 -->
        <div class="relative flex-1 max-w-sm">
          <Search class="w-4 h-4 text-muted-foreground absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索渠道，如 微信、钉钉、Server"
            class="w-full pl-9 pr-3 py-2 text-xs rounded-xl bg-muted/40 dark:bg-zinc-800/60 border border-border/50 text-foreground placeholder:text-muted-foreground/70 focus:outline-none focus:ring-1 focus:ring-primary/40 focus:border-primary/50 transition-all"
          />
        </div>

        <!-- 分类选择胶囊组 -->
        <div class="inline-flex items-center p-1 rounded-xl bg-muted/40 dark:bg-zinc-800/60 border border-border/40 text-xs shrink-0">
          <button
            v-for="cat in channelCategories"
            :key="cat.id"
            type="button"
            @click="activeCategory = cat.id"
            class="px-3 py-1 rounded-lg font-medium transition-all cursor-pointer"
            :class="[
              activeCategory === cat.id
                ? 'bg-background text-foreground shadow-xs'
                : 'text-muted-foreground hover:text-foreground'
            ]"
          >
            {{ cat.label }}
          </button>
        </div>
      </div>

      <!-- 渠道卡片网格列表 (3列网格) -->
      <div class="flex-1 overflow-y-auto pr-1">
        <div
          v-if="filteredChannels.length > 0"
          class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3"
        >
          <div
            v-for="channel in filteredChannels"
            :key="channel.type"
            @click="handleSelectChannel(channel)"
            class="group relative flex items-center gap-3.5 p-3 rounded-2xl border transition-all duration-200 cursor-pointer active:scale-[0.98] bg-muted/20 dark:bg-zinc-900/40 hover:bg-muted/60 dark:hover:bg-zinc-800/60 border-border/40 hover:border-border/80 hover:shadow-xs"
          >
            <!-- 渠道图标容器 -->
            <div
              class="w-11 h-11 rounded-xl flex items-center justify-center shrink-0 border border-border/20 transition-transform duration-200 group-hover:scale-105"
              :class="getChannelVisual(channel.type).iconBg"
            >
              <component :is="getChannelVisual(channel.type).icon" class="w-5 h-5" />
            </div>

            <!-- 渠道名称与副描述 -->
            <div class="flex-1 min-w-0 pr-1">
              <div class="flex items-center gap-1.5">
                <span class="text-sm font-semibold text-foreground group-hover:text-primary transition-colors truncate">
                  {{ channel.label }}
                </span>
                <span
                  v-if="channel.dynamicRecipient?.support"
                  class="inline-block w-1.5 h-1.5 rounded-full bg-emerald-500 shrink-0"
                  title="支持群发"
                />
              </div>
              <p class="text-xs text-muted-foreground mt-0.5 truncate">
                {{ channel.meta?.subTitle || channel.subTitle || channel.label }}
              </p>
            </div>
          </div>
        </div>

        <!-- 空搜索提示 -->
        <div
          v-else
          class="py-16 text-center text-muted-foreground flex flex-col items-center justify-center"
        >
          <Search class="w-8 h-8 opacity-30 mb-2" />
          <p class="text-sm">未找到与「{{ searchQuery }}」匹配的通知渠道</p>
          <button
            type="button"
            @click="searchQuery = ''; activeCategory = 'all'"
            class="mt-3 text-xs text-primary hover:underline"
          >
            清除搜索条件
          </button>
        </div>
      </div>
    </div>

    <!-- 步骤二：配置通知渠道 (完全参考设计图2) -->
    <div
      v-else
      class="flex-1 overflow-y-auto px-6 py-5 flex flex-col min-h-0"
    >
      <!-- 顶部已选渠道预览条卡片 -->
      <div class="rounded-2xl p-3.5 bg-muted/30 dark:bg-zinc-900/40 border border-border/50 flex items-center justify-between mb-5 shrink-0">
        <div class="flex items-center gap-3.5">
          <div
            class="w-11 h-11 rounded-xl flex items-center justify-center shrink-0 border border-border/20"
            :class="getChannelVisual(currentChannel.type).iconBg"
          >
            <component :is="getChannelVisual(currentChannel.type).icon" class="w-5 h-5" />
          </div>
          <div>
            <div class="flex items-center gap-2">
              <span class="text-sm font-bold text-foreground">{{ currentChannel.label }}</span>
              <span class="text-[10px] px-2 py-0.5 rounded-full bg-muted dark:bg-zinc-800 text-muted-foreground font-medium">
                {{ currentChannel.meta?.category || currentChannel.category || '通知渠道' }}
              </span>
            </div>
            <p class="text-xs text-muted-foreground mt-0.5">{{ currentChannel.meta?.subTitle || currentChannel.subTitle || currentChannel.label }}</p>
          </div>
        </div>

        <!-- 更换渠道按钮（仅在新增模式下展示） -->
        <button
          v-if="props.mode === 'add'"
          type="button"
          @click="handleBackToSelect"
          class="inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground px-2.5 py-1.5 rounded-lg border border-border/40 hover:bg-muted/40 transition-all cursor-pointer"
        >
          <ArrowLeft class="w-3.5 h-3.5" />
          <span>更换渠道</span>
        </button>
      </div>

      <!-- 表单输入区域 -->
      <div class="flex-1 space-y-4">
        <!-- 通用设置项 -->
        <div class="space-y-3.5">
          <!-- 渠道名称 -->
          <div class="flex flex-col sm:flex-row sm:items-center gap-1.5 sm:gap-4">
            <label class="w-24 sm:w-28 shrink-0 text-sm font-medium text-foreground">
              名称
            </label>
            <div class="flex-1">
              <Input
                v-model="formData.name"
                :placeholder="currentChannel.label"
                class="h-10 bg-muted/20 dark:bg-zinc-900/40 border-border/60 rounded-xl"
                @input="clearFieldError('name')"
              />
              <p v-if="validationErrors.name" class="text-xs text-red-500 mt-1 pl-0.5">
                {{ validationErrors.name }}
              </p>
            </div>
          </div>
        </div>

        <!-- 参数配置分区标题 -->
        <div class="pt-3 pb-1 border-t border-border/30">
          <div class="text-xs font-semibold text-muted-foreground tracking-wide">
            参数配置
          </div>
        </div>

        <!-- 各渠道专属参数字段 (来自 CONSTANT.WAYS_DATA 的 inputs) -->
        <div class="space-y-3.5">
          <div
            v-for="input in channelParamInputs"
            :key="input.col"
            class="flex flex-col sm:flex-row sm:items-start gap-1.5 sm:gap-4"
          >
            <label class="w-24 sm:w-28 shrink-0 text-sm font-medium text-foreground pt-2 flex items-center">
              <span>{{ input.subLabel || input.label }}</span>
              <span v-if="input.required" class="text-red-500 ml-1 font-bold">*</span>
            </label>
            <div class="flex-1 space-y-1">
              <Textarea
                v-if="input.isTextArea"
                v-model="formData[input.col]"
                :placeholder="input.placeholder || input.desc || input.subLabel"
                rows="3"
                class="bg-muted/20 dark:bg-zinc-900/40 border-border/60 rounded-xl resize-none text-xs leading-relaxed"
                @input="clearFieldError(input.col)"
              />
              <Input
                v-else
                v-model="formData[input.col]"
                :placeholder="input.placeholder || input.desc || input.subLabel"
                class="h-10 bg-muted/20 dark:bg-zinc-900/40 border-border/60 rounded-xl text-sm"
                @input="clearFieldError(input.col)"
              />
              <p v-if="validationErrors[input.col]" class="text-xs text-red-500 pl-0.5">
                {{ validationErrors[input.col] }}
              </p>
            </div>
          </div>

          <!-- 自托管消息等无参数渠道提示 -->
          <div
            v-if="channelParamInputs.length === 0"
            class="py-6 px-4 rounded-xl bg-muted/20 dark:bg-zinc-900/40 border border-border/40 text-center text-xs text-muted-foreground"
          >
            当前渠道无需额外参数，保存后即可直接使用。
          </div>
        </div>

        <!-- 群发支持提示 -->
        <div
          v-if="currentChannel.dynamicRecipient?.support"
          class="mt-4 p-3 rounded-xl bg-brand/5 dark:bg-brand/10 border border-brand/20 flex items-start gap-2.5"
        >
          <Info class="w-4 h-4 text-brand shrink-0 mt-0.5" />
          <div class="text-xs text-brand/90 leading-relaxed">
            <span class="font-medium">支持群发模式：</span>在发送任务中可通过 API 指定多个 {{ currentChannel.dynamicRecipient.label }} 进行批量投递。
          </div>
        </div>

        <!-- 详细使用说明 -->
        <div
          v-if="currentChannel.tips?.desc"
          class="mt-2 text-xs text-muted-foreground/80 leading-relaxed bg-muted/10 p-3 rounded-xl border border-border/30"
          v-html="currentChannel.tips.desc"
        />
      </div>

      <!-- 底部操作按钮栏 -->
      <div class="flex items-center justify-end gap-2.5 pt-4 mt-6 border-t border-border/40 shrink-0">
        <Button variant="outline" class="rounded-xl px-4" @click="handleClose">
          取消
        </Button>
        <Button
          variant="secondary"
          class="rounded-xl px-4"
          :disabled="isTesting"
          @click="handleTest"
        >
          <Loader2 v-if="isTesting" class="w-3.5 h-3.5 animate-spin mr-1.5" />
          <span>测试</span>
        </Button>
        <Button
          class="rounded-xl px-5 btn-brand-premium font-medium"
          :disabled="isSaving"
          @click="handleSave"
        >
          <Loader2 v-if="isSaving" class="w-3.5 h-3.5 animate-spin mr-1.5" />
          <span>{{ props.mode === 'edit' ? '更新' : '保存' }}</span>
        </Button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'WaysForm'
})
</script>