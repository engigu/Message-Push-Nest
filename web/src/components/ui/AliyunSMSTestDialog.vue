<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { request } from '@/api/api'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { toast } from 'vue-sonner'
import { Copy } from 'lucide-vue-next'
import { buildAliyunTestPayload, getAliyunTestValues, parseTemplatePlaceholders } from '@/lib/aliyunSmsConfig.mjs'

interface Props {
  open: boolean
  scenario: 'new' | 'edit' | 'saved'
  instanceId: string
  templateId: string
  wayId: string
  config: Record<string, any>
  placeholders: Array<{ key: string; label: string; default: string }>
}

const props = defineProps<Props>()
const emit = defineEmits<{ 'update:open': [value: boolean] }>()

const testPhone = ref('')
const testValues = ref<Record<string, string>>({})
const isSending = ref(false)
const errorMessage = ref('')
const providerResult = ref<any>(null)
const providerRequestJson = computed(() => {
  const params = providerResult.value?.request_params
  return params ? JSON.stringify(params, null, 2) : ''
})
const providerTemplateParamJson = computed(() => {
  const raw = providerResult.value?.request_params?.TemplateParam
  if (typeof raw !== 'string') return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
})

const selectedPlaceholders = computed(() => {
  const all = parseTemplatePlaceholders(props.placeholders)
  if (props.config.template_param_mode !== 'structured') return all
  const selected = new Set(props.config.selected_params || [])
  return all.filter((item) => selected.has(item.key))
})

const mappingPreview = computed(() => selectedPlaceholders.value.map((item) => ({
  key: item.key,
  target: props.config.param_name_overrides?.[item.key] || item.key,
})))

const modeLabel = computed(() => props.config.template_param_mode === 'structured' ? 'structured 多参数' : 'legacy content 兼容模式')

watch(() => [props.open, props.scenario, props.config, props.placeholders], () => {
  if (!props.open) return
  testPhone.value = ''
  testValues.value = getAliyunTestValues(props.config, parseTemplatePlaceholders(props.placeholders))
  errorMessage.value = ''
  providerResult.value = null
}, { immediate: true, deep: true })

const copyProviderParams = async () => {
  try {
    await navigator.clipboard.writeText(providerRequestJson.value)
    toast.success('已复制阿里云 SDK 参数')
  } catch {
    toast.error('复制失败，请手动选择 JSON 内容复制')
  }
}

const sendTest = async () => {
  if (!/^\+?[0-9]{5,20}$/.test(testPhone.value.trim())) {
    errorMessage.value = '请输入有效的测试手机号'
    return
  }
  if (props.scenario !== 'saved' && (!props.templateId || !props.wayId)) {
    errorMessage.value = '模板或阿里云发送渠道信息缺失，请关闭后重新选择渠道'
    return
  }

  isSending.value = true
  errorMessage.value = ''
  providerResult.value = null
  const payload = buildAliyunTestPayload({
    scenario: props.scenario,
    instanceId: props.instanceId,
    templateId: props.templateId,
    wayId: props.wayId,
    config: props.config,
    testPhone: testPhone.value.trim(),
    placeholders: testValues.value,
  })
  try {
    const response = await request.post('/templates/ins/aliyun/test', payload)
    const body = response.data || {}
    if (response.status === 200 && body.code === 200) {
      providerResult.value = body.data || {}
      if (providerResult.value.accepted) {
        toast.success(body.msg || '阿里云已受理测试请求')
      } else {
        toast.error(providerResult.value.message || body.msg || '阿里云未受理测试请求')
      }
    } else {
      errorMessage.value = body.msg || '发送测试失败'
      providerResult.value = body.data || null
    }
  } catch (error: any) {
    const body = error.response?.data
    errorMessage.value = body?.msg || '发送测试请求失败，请稍后重试'
    providerResult.value = body?.data || null
  } finally {
    isSending.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="(value) => emit('update:open', value)">
    <DialogContent class="max-h-[90vh] max-w-xl overflow-y-auto">
      <DialogHeader>
        <DialogTitle>阿里云短信发送测试</DialogTitle>
        <DialogDescription>
          {{ scenario === 'new' ? '使用当前未保存的新建配置' : scenario === 'edit' ? '使用当前尚未保存的编辑配置' : '使用实例当前已保存配置' }}。
          测试会调用供应商接口并可能产生短信费用。
        </DialogDescription>
      </DialogHeader>

      <div class="space-y-4">
        <div class="rounded-md border border-amber-300 bg-amber-50 p-3 text-sm text-amber-950">
          确认后会向测试手机号发起真实供应商调用。阿里云接口受理成功不代表短信最终送达。
        </div>
        <div class="space-y-2">
          <Label for="aliyun-test-phone">测试手机号</Label>
          <Input id="aliyun-test-phone" v-model="testPhone" inputmode="tel" autocomplete="off" placeholder="仅用于本次测试，不修改正式接收者" />
        </div>
        <div class="rounded-md bg-muted p-3 text-sm">
          <div><span class="font-medium">TemplateCode：</span>{{ config.template_code || '未设置' }}</div>
          <div class="mt-1"><span class="font-medium">参数模式：</span>{{ modeLabel }}</div>
          <div v-if="config.template_param_mode === 'structured' && mappingPreview.length" class="mt-2 space-y-1">
            <div class="font-medium">参数映射预览</div>
            <div v-for="mapping in mappingPreview" :key="mapping.key">{{ mapping.key }} → {{ mapping.target }}</div>
          </div>
          <div v-else class="mt-2 text-xs text-muted-foreground">
            模板变量会先用于渲染短信内容，再按 legacy content 参数模式传递。
          </div>
        </div>
        <div v-if="selectedPlaceholders.length" class="space-y-3">
          <div class="text-sm font-medium">本次测试变量值</div>
          <div v-for="placeholder in selectedPlaceholders" :key="placeholder.key" class="space-y-1">
            <Label :for="`aliyun-test-value-${placeholder.key}`">{{ placeholder.label }}（{{ placeholder.key }}）</Label>
            <Input :id="`aliyun-test-value-${placeholder.key}`" v-model="testValues[placeholder.key]" :placeholder="placeholder.default || `请输入 ${placeholder.key} 的测试值`" />
          </div>
        </div>
        <p v-if="errorMessage" role="alert" class="text-sm text-destructive">{{ errorMessage }}</p>
        <div v-if="providerResult" class="space-y-1 rounded-md border p-3 text-sm" aria-live="polite">
          <div class="font-medium">供应商调用结果</div>
          <div>Code：{{ providerResult.code || '-' }}</div>
          <div>Message：{{ providerResult.message || '-' }}</div>
          <div>RequestId：{{ providerResult.request_id || '-' }}</div>
          <div>BizId：{{ providerResult.biz_id || '-' }}</div>
          <div>状态：{{ providerResult.accepted ? '阿里云已受理' : '阿里云未确认受理' }}</div>
          <div v-if="providerResult.troubleshooting" class="mt-2 rounded bg-muted p-2 text-xs leading-relaxed">
            排查建议：{{ providerResult.troubleshooting }}
            <a
              v-if="providerResult.request_id"
              class="ml-1 underline underline-offset-2"
              :href="`https://api.aliyun.com/troubleshoot?q=${encodeURIComponent(providerResult.code || '')}&product=Dysmsapi&requestId=${encodeURIComponent(providerResult.request_id)}`"
              target="_blank"
              rel="noopener noreferrer"
            >打开阿里云诊断</a>
          </div>
          <div class="text-xs text-muted-foreground">{{ providerResult.delivery_notice || '接口受理不代表短信最终送达。' }}</div>
        </div>
        <details v-if="providerRequestJson" class="rounded-md border p-3">
          <summary class="cursor-pointer text-sm font-medium">查看本系统传给阿里云 SDK 的参数</summary>
          <p class="mt-2 text-xs text-muted-foreground">
            这是后端实际传入 SendSmsRequest 的参数，不含 AccessKey。阿里云 Go SDK 会据此生成签名 RPC 请求，线上请求不是 JSON body。
            PhoneNumbers 是本次测试手机号；正式任务发送时使用任务收件人。
          </p>
          <div class="mt-2 flex justify-end">
            <Button size="sm" variant="outline" @click.prevent="copyProviderParams">
              <Copy class="mr-1 h-3.5 w-3.5" />复制 SDK 参数 JSON
            </Button>
          </div>
          <pre class="mt-2 max-h-64 overflow-auto rounded bg-muted p-3 text-xs leading-relaxed">{{ providerRequestJson }}</pre>
          <div v-if="providerTemplateParamJson" class="mt-3">
            <div class="text-xs font-medium">TemplateParam 解码预览</div>
            <pre class="mt-1 max-h-40 overflow-auto rounded bg-muted p-3 text-xs leading-relaxed">{{ providerTemplateParamJson }}</pre>
          </div>
        </details>
      </div>

      <DialogFooter>
        <Button variant="outline" :disabled="isSending" @click="emit('update:open', false)">关闭</Button>
        <Button :disabled="isSending" @click="sendTest">{{ isSending ? '调用中...' : '确认并发送测试' }}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
