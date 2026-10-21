<script setup lang="ts">
import { reactive, onMounted, onUnmounted, ref } from 'vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
import { useRouter } from 'vue-router'
import { CONSTANT } from '@/constant'
import { usePageState } from '@/store/page_sate'
import { checkIsDemoMode } from '@/util/demo'
import {
  HelpCircleIcon,
  ClockIcon,
  PlayIcon,
  ShieldCheckIcon,
  FileTextIcon,
  MessageSquareIcon,
  Loader2,
  ExternalLinkIcon
} from 'lucide-vue-next'

const router = useRouter()
const pageState = usePageState()
const isDemo = ref(checkIsDemoMode())

// 全局系统维护 Cron 配置（管理员专有调度状态）
const systemCronState = reactive({
  cron: '0 3 * * *',
  saving: false,
  executing: false,
})

// 立即维护防抖与冷却倒计时 (秒)
const cooldownSeconds = ref(0)
let cooldownTimer: any = null

// 用户自身日志清理策略状态
const logsState = reactive({
  section: 'log_config',
  keepNum: '1000',
  enabled: true,
  saving: false,
})

// 用户自身托管消息清理策略状态
const hostedMsgState = reactive({
  section: 'hosted_msg_config',
  keepNum: '50000',
  enabled: false,
  saving: false,
})

// 管理员更新全局调度统一维护周期
const handleAdminCronSubmit = async () => {
  if (isDemo.value || checkIsDemoMode()) {
    toast.error('演示模式下禁止修改清理配置')
    return
  }
  const cronExpr = systemCronState.cron.trim()
  if (!cronExpr) {
    toast.error('Cron 表达式不能为空')
    return
  }
  try {
    systemCronState.saving = true
    const postData = {
      section: logsState.section,
      data: {
        cron: cronExpr,
        keep_num: logsState.keepNum.trim() || '1000',
        enabled: logsState.enabled ? 'true' : 'false',
      },
    }
    const response = await request.post('/settings/set', postData)
    if (response.data.code === 200) {
      toast.success(`系统统一调度周期已更新为：${cronExpr}`)
    } else {
      toast.error(response.data.message || '更新系统调度周期失败')
    }
  } catch (error: any) {
    toast.error(error.response?.data?.message || '更新系统调度周期失败，请稍后重试')
  } finally {
    systemCronState.saving = false
  }
}

// 快速应用推荐的系统 Cron 周期
const applyPresetCron = (expr: string) => {
  if (isDemo.value || checkIsDemoMode()) {
    toast.error('演示模式下禁止修改清理配置')
    return
  }
  systemCronState.cron = expr
}

// 管理员立即手动触发一次系统统一维护清理 (带防抖与冷却限制)
const handleExecuteCleanNow = async () => {
  if (isDemo.value || checkIsDemoMode()) {
    toast.error('演示模式下禁止执行数据清理')
    return
  }
  if (systemCronState.executing || cooldownSeconds.value > 0) {
    toast.info(`操作过于频繁，请在 ${cooldownSeconds.value} 秒冷却后再试`)
    return
  }
  try {
    systemCronState.executing = true
    const response = await request.post('/settings/execute_clean', {})
    if (response.data.code === 200) {
      toast.success(response.data.message || '已触发系统数据维护清理，可在推送日志中查看执行结果！')
      cooldownSeconds.value = 10
      if (cooldownTimer) clearInterval(cooldownTimer)
      cooldownTimer = setInterval(() => {
        if (cooldownSeconds.value > 1) {
          cooldownSeconds.value--
        } else {
          cooldownSeconds.value = 0
          clearInterval(cooldownTimer)
          cooldownTimer = null
        }
      }, 1000)
    } else {
      toast.error(response.data.message || '触发清理任务失败')
    }
  } catch (error: any) {
    toast.error(error.response?.data?.message || '触发清理任务失败，请稍后重试')
  } finally {
    systemCronState.executing = false
  }
}

// 提交个人日志清理策略
const handleLogsSubmit = async () => {
  if (isDemo.value || checkIsDemoMode()) {
    toast.error('演示模式下禁止修改清理配置')
    return
  }
  const keep = parseInt(logsState.keepNum.trim(), 10)
  if (isNaN(keep) || keep <= 0) {
    toast.error('保留日志条数必须为大于 0 的正整数')
    return
  }
  try {
    logsState.saving = true
    const postData = {
      section: logsState.section,
      data: {
        keep_num: String(keep),
        enabled: logsState.enabled ? 'true' : 'false',
      },
    }
    const response = await request.post('/settings/set', postData)
    if (response.data.code === 200) {
      const statusText = logsState.enabled ? '已启用' : '已关闭'
      toast.success(`推送日志清理策略已保存（${statusText}，保留最近 ${keep} 条）`)
    } else {
      toast.error(response.data.message || '保存失败')
    }
  } catch (error: any) {
    toast.error(error.response?.data?.message || '保存失败，请稍后重试')
  } finally {
    logsState.saving = false
  }
}

// 提交个人托管消息清理策略
const handleHostedMsgSubmit = async () => {
  if (isDemo.value || checkIsDemoMode()) {
    toast.error('演示模式下禁止修改清理配置')
    return
  }
  const keep = parseInt(hostedMsgState.keepNum.trim(), 10)
  if (isNaN(keep) || keep <= 0) {
    toast.error('保留托管消息条数必须为大于 0 的正整数')
    return
  }
  try {
    hostedMsgState.saving = true
    const postData = {
      section: hostedMsgState.section,
      data: {
        keep_num: String(keep),
        enabled: hostedMsgState.enabled ? 'true' : 'false',
      },
    }
    const response = await request.post('/settings/set', postData)
    if (response.data.code === 200) {
      const statusText = hostedMsgState.enabled ? '已启用' : '已关闭'
      toast.success(`托管消息清理策略已保存（${statusText}，保留最近 ${keep} 条）`)
    } else {
      toast.error(response.data.message || '保存失败')
    }
  } catch (error: any) {
    toast.error(error.response?.data?.message || '保存失败，请稍后重试')
  } finally {
    hostedMsgState.saving = false
  }
}

// 查看日志清理执行历史
const handleLogsView = () => {
  router.push({ path: '/sendlogs', query: { taskid: CONSTANT.LOG_TASK_ID } })
}

// 查看托管消息清理执行历史
const handleHostedMsgView = () => {
  router.push({ path: '/sendlogs', query: { taskid: CONSTANT.HOSTED_MSG_TASK_ID } })
}

// 获取日志清理配置
const getLogsConfig = async () => {
  try {
    const params = { params: { section: 'log_config' } }
    const response = await request.get('/settings/getsetting', params)
    if (response.data.code === 200) {
      const data = response.data.data
      logsState.keepNum = data.keep_num || '1000'
      logsState.enabled = data.enabled === 'true' || data.enabled === true
      if (data.cron) {
        systemCronState.cron = data.cron
      }
    }
  } catch (error) {
    toast.error('获取日志清理配置失败')
  }
}

// 获取托管消息清理配置
const getHostedMsgConfig = async () => {
  try {
    const params = { params: { section: 'hosted_msg_config' } }
    const response = await request.get('/settings/getsetting', params)
    if (response.data.code === 200) {
      const data = response.data.data
      hostedMsgState.keepNum = data.keep_num || '50000'
      hostedMsgState.enabled = data.enabled === 'true' || data.enabled === true
      if (data.cron && !systemCronState.cron) {
        systemCronState.cron = data.cron
      }
    }
  } catch (error) {
    toast.error('获取托管消息清理配置失败')
  }
}

onMounted(() => {
  getLogsConfig()
  getHostedMsgConfig()
})

onUnmounted(() => {
  if (cooldownTimer) {
    clearInterval(cooldownTimer)
    cooldownTimer = null
  }
})
</script>

<script lang="ts">
export default {
  name: 'CleanSettings'
}
</script>

<template>
  <Card class="overflow-hidden">
    <CardHeader class="px-4 py-4 sm:px-6 sm:py-5">
      <CardTitle class="text-lg sm:text-xl">数据清理设置</CardTitle>
      <CardDescription class="text-xs sm:text-sm">配置统一调度周期与各用户独立的数据保留策略</CardDescription>
    </CardHeader>

    <CardContent class="px-3.5 pb-4 sm:px-6 sm:pb-6 space-y-4 sm:space-y-6">
      <!-- 演示模式提示横幅 -->
      <div
        v-if="isDemo"
        class="p-3 sm:p-3.5 rounded-lg border border-amber-500/30 bg-amber-500/10 text-xs text-amber-700 dark:text-amber-400 flex items-center gap-2"
      >
        <span>⚠️ 当前系统处于演示模式，数据清理配置仅供只读预览展示，禁止修改配置。</span>
      </div>

      <!-- 架构机制提示横幅 -->
      <div class="p-3 sm:p-3.5 rounded-lg border border-primary/20 bg-primary/5 flex items-start gap-2 sm:gap-2.5">
        <ShieldCheckIcon class="w-4 h-4 text-primary shrink-0 mt-0.5" />
        <div class="text-[11px] sm:text-xs text-muted-foreground leading-relaxed">
          <span class="font-medium text-foreground">统一调度 + 租户隔离机制：</span>
          系统每日低峰期自动运行单一守护作业，依次读取每位用户的专属保留条数并执行隔离清理。普通用户只需配置保留条数与开关，无需关心复杂的 Cron 执行周期。
        </div>
      </div>

      <!-- 管理员全局调度控制台 (仅管理员可见) -->
      <div
        v-if="pageState.isAdmin"
        class="p-3.5 sm:p-4 rounded-xl border border-border bg-muted/20 space-y-3"
      >
        <div class="flex flex-wrap items-center justify-between gap-1.5 sm:gap-2 border-b border-border/60 pb-2.5">
          <div class="flex items-center gap-1.5 sm:gap-2">
            <ClockIcon class="w-4 h-4 text-primary" />
            <span class="text-xs sm:text-sm font-semibold text-foreground">系统统一维护调度周期</span>
            <span class="text-[10px] sm:text-[11px] px-1.5 py-0.5 rounded-full bg-primary/10 text-primary font-medium">管理员专享</span>
          </div>
          <div class="text-[11px] sm:text-xs text-muted-foreground flex items-center gap-1">
            <span>当前生效：</span>
            <code class="px-1.5 py-0.5 rounded bg-muted font-mono text-foreground font-semibold text-[11px]">
              {{ systemCronState.cron || '0 3 * * *' }}
            </code>
          </div>
        </div>

        <!-- 调度控制行：输入框 + 预设胶囊 + 操作按钮 -->
        <div class="flex flex-col lg:flex-row items-stretch lg:items-center gap-2.5 sm:gap-3 pt-0.5">
          <!-- Cron 输入框 -->
          <div class="relative w-full lg:w-56 shrink-0">
            <Input
              v-model="systemCronState.cron"
              placeholder="例如: 0 3 * * *"
              class="font-mono text-xs pr-7 h-9 w-full"
              :disabled="isDemo || systemCronState.saving"
            />
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger as-child>
                  <button type="button" class="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground/70 hover:text-foreground">
                    <HelpCircleIcon class="w-3.5 h-3.5" />
                  </button>
                </TooltipTrigger>
                <TooltipContent class="max-w-xs text-xs">
                  标准 5 位 Cron 表达式：分 时 日 月 周。例如 "0 3 * * *" 表示每天凌晨 03:00。
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>

          <!-- 低峰期推荐预设按钮 -->
          <div class="flex items-center gap-1.5 flex-wrap flex-1 min-w-0">
            <span class="text-[11px] sm:text-xs text-muted-foreground shrink-0 mr-0.5">常用预设:</span>
            <div class="flex items-center gap-1.5 flex-wrap">
              <button
                v-for="preset in [
                  { label: '每日 03:00', expr: '0 3 * * *' },
                  { label: '每日 04:00', expr: '0 4 * * *' },
                  { label: '每日 02:30', expr: '30 2 * * *' },
                ]"
                :key="preset.expr"
                type="button"
                class="text-[11px] sm:text-xs px-2 sm:px-2.5 py-1 rounded-md border border-border bg-background hover:bg-muted text-muted-foreground hover:text-foreground transition-all shrink-0 whitespace-nowrap disabled:opacity-40"
                :class="{ 'border-primary/60 text-primary font-medium bg-primary/5': systemCronState.cron === preset.expr }"
                :disabled="isDemo"
                @click="applyPresetCron(preset.expr)"
              >
                {{ preset.label }}
              </button>
            </div>
          </div>

          <!-- 管理员操作按钮 (小屏网格平分，大屏右对齐) -->
          <div class="grid grid-cols-2 sm:flex sm:items-center gap-2 shrink-0 pt-1 sm:pt-0">
            <Button
              variant="outline"
              size="sm"
              class="h-9 text-xs transition-all w-full sm:w-auto"
              :disabled="isDemo || systemCronState.executing || cooldownSeconds > 0"
              :title="cooldownSeconds > 0 ? `冷却中，请在 ${cooldownSeconds} 秒后重试` : '立即唤醒调度引擎执行一次全系统数据清理维护'"
              @click="handleExecuteCleanNow"
            >
              <Loader2 v-if="systemCronState.executing" class="w-3.5 h-3.5 mr-1 animate-spin shrink-0" />
              <PlayIcon v-else class="w-3.5 h-3.5 mr-1 shrink-0" />
              <span class="truncate">
                {{
                  systemCronState.executing
                    ? '维护中...'
                    : cooldownSeconds > 0
                      ? `${cooldownSeconds}s 后可维护`
                      : '立即维护'
                }}
              </span>
            </Button>
            <Button
              size="sm"
              class="h-9 text-xs w-full sm:w-auto"
              :disabled="isDemo || systemCronState.saving"
              @click="handleAdminCronSubmit"
            >
              <span>{{ systemCronState.saving ? '保存中...' : '保存周期' }}</span>
            </Button>
          </div>
        </div>
      </div>

      <!-- 分隔线 -->
      <div class="border-t border-border/60"></div>

      <!-- 租户独立数据保留策略：大屏左右双栏，小屏纵向 -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-6">
        <!-- 1. 推送日志清理模块 -->
        <div class="p-3.5 sm:p-5 rounded-xl border border-border/80 bg-muted/10 flex flex-col justify-between space-y-3.5 sm:space-y-4">
          <div class="space-y-3 sm:space-y-4">
            <!-- 头部标题与状态 -->
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-1.5 sm:gap-2">
                <FileTextIcon class="w-4 h-4 text-primary shrink-0" />
                <h3 class="text-sm sm:text-base font-semibold text-foreground">推送日志清理</h3>
              </div>
              <span
                class="text-[11px] sm:text-xs px-2 py-0.5 rounded-full font-medium shrink-0"
                :class="logsState.enabled ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400' : 'bg-muted text-muted-foreground'"
              >
                {{ logsState.enabled ? '已启用' : '已停用' }}
              </span>
            </div>

            <!-- 开关卡片行 -->
            <div class="flex items-center justify-between p-3 sm:p-3.5 border border-border rounded-lg bg-card/60 gap-2">
              <div class="space-y-0.5 flex-1 min-w-0 pr-1">
                <Label class="text-xs sm:text-sm font-medium text-foreground cursor-pointer block truncate">自动清理推送日志</Label>
                <div class="text-[11px] sm:text-xs text-muted-foreground leading-snug">
                  超出保留数量的历史日志将被自动清理
                </div>
              </div>
              <Switch
                v-model="logsState.enabled"
                :disabled="isDemo"
                class="shrink-0"
                @update:model-value="handleLogsSubmit"
              />
            </div>

            <!-- 保留条数输入与快速预设 -->
            <div class="space-y-1.5 sm:space-y-2">
              <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-0.5 sm:gap-2">
                <label class="text-xs sm:text-sm font-medium text-foreground">保留日志条数</label>
                <span class="text-[11px] sm:text-xs text-muted-foreground">超出该条数的数据将被清除</span>
              </div>
              <Input
                v-model="logsState.keepNum"
                type="number"
                min="1"
                placeholder="默认 1000"
                :disabled="!logsState.enabled || isDemo"
                class="font-mono text-xs sm:text-sm h-9"
              />
              <!-- 快捷预设标签：自适应换行，防止文字被竖向折叠 -->
              <div class="flex flex-wrap items-center gap-1.5 pt-1">
                <span class="text-[11px] text-muted-foreground shrink-0 select-none mr-0.5">快速设置:</span>
                <div class="flex flex-wrap gap-1.5">
                  <button
                    v-for="preset in [
                      { label: '500条', val: '500' },
                      { label: '1000条', val: '1000' },
                      { label: '3000条', val: '3000' },
                      { label: '5000条', val: '5000' }
                    ]"
                    :key="preset.val"
                    type="button"
                    class="text-[11px] px-2 py-0.5 rounded border border-border bg-card hover:bg-muted text-muted-foreground hover:text-foreground transition-colors shrink-0 whitespace-nowrap disabled:opacity-40 disabled:cursor-not-allowed"
                    :class="{ 'border-primary/50 text-primary bg-primary/5 font-medium': logsState.keepNum === preset.val }"
                    :disabled="!logsState.enabled || isDemo"
                    @click="logsState.keepNum = preset.val"
                  >
                    {{ preset.label }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- 底部操作与说明 (弹性布局，杜绝窄屏贴边推挤) -->
          <div class="pt-3 border-t border-border/60 flex items-center justify-between gap-2">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger as-child>
                  <button type="button" class="text-xs text-muted-foreground hover:text-foreground flex items-center gap-1 transition-colors shrink-0 py-1">
                    <HelpCircleIcon class="w-3.5 h-3.5" />
                    <span class="text-[11px] sm:text-xs">执行规则</span>
                  </button>
                </TooltipTrigger>
                <TooltipContent class="max-w-xs text-xs">
                  <p>1. 数据按当前账号严格隔离，不会影响其他用户的日志记录。</p>
                  <p class="mt-1">2. 维护任务由系统守护进程统一调度，按时间倒序保留最新的 N 条。</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>

            <div class="flex items-center gap-2 shrink-0">
              <Button variant="outline" size="sm" class="h-8 px-2.5 sm:px-3 text-xs" @click="handleLogsView">
                <ExternalLinkIcon class="w-3 h-3 mr-1 shrink-0" />
                <span>日志</span>
              </Button>
              <Button
                size="sm"
                class="h-8 px-3 sm:px-3.5 text-xs"
                :disabled="isDemo || logsState.saving"
                @click="handleLogsSubmit"
              >
                {{ logsState.saving ? '保存中...' : '保存配置' }}
              </Button>
            </div>
          </div>
        </div>

        <!-- 2. 托管消息清理模块 -->
        <div class="p-3.5 sm:p-5 rounded-xl border border-border/80 bg-muted/10 flex flex-col justify-between space-y-3.5 sm:space-y-4">
          <div class="space-y-3 sm:space-y-4">
            <!-- 头部标题与状态 -->
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-1.5 sm:gap-2">
                <MessageSquareIcon class="w-4 h-4 text-primary shrink-0" />
                <h3 class="text-sm sm:text-base font-semibold text-foreground">托管消息清理</h3>
              </div>
              <span
                class="text-[11px] sm:text-xs px-2 py-0.5 rounded-full font-medium shrink-0"
                :class="hostedMsgState.enabled ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400' : 'bg-muted text-muted-foreground'"
              >
                {{ hostedMsgState.enabled ? '已启用' : '已停用' }}
              </span>
            </div>

            <!-- 开关卡片行 -->
            <div class="flex items-center justify-between p-3 sm:p-3.5 border border-border rounded-lg bg-card/60 gap-2">
              <div class="space-y-0.5 flex-1 min-w-0 pr-1">
                <Label class="text-xs sm:text-sm font-medium text-foreground cursor-pointer block truncate">自动清理托管消息</Label>
                <div class="text-[11px] sm:text-xs text-muted-foreground leading-snug">
                  超出保留数量的托管消息将被自动清理
                </div>
              </div>
              <Switch
                v-model="hostedMsgState.enabled"
                :disabled="isDemo"
                class="shrink-0"
                @update:model-value="handleHostedMsgSubmit"
              />
            </div>

            <!-- 保留条数输入与快速预设 -->
            <div class="space-y-1.5 sm:space-y-2">
              <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-0.5 sm:gap-2">
                <label class="text-xs sm:text-sm font-medium text-foreground">保留托管消息条数</label>
                <span class="text-[11px] sm:text-xs text-muted-foreground">超出该条数的数据将被清除</span>
              </div>
              <Input
                v-model="hostedMsgState.keepNum"
                type="number"
                min="1"
                placeholder="默认 50000"
                :disabled="!hostedMsgState.enabled || isDemo"
                class="font-mono text-xs sm:text-sm h-9"
              />
              <!-- 快捷预设标签：针对大数提供 1万/3万/5万/10万 紧凑标签，杜绝窄屏爆版折行 -->
              <div class="flex flex-wrap items-center gap-1.5 pt-1">
                <span class="text-[11px] text-muted-foreground shrink-0 select-none mr-0.5">快速设置:</span>
                <div class="flex flex-wrap gap-1.5">
                  <button
                    v-for="preset in [
                      { label: '1万条', val: '10000' },
                      { label: '3万条', val: '30000' },
                      { label: '5万条', val: '50000' },
                      { label: '10万条', val: '100000' }
                    ]"
                    :key="preset.val"
                    type="button"
                    class="text-[11px] px-2 py-0.5 rounded border border-border bg-card hover:bg-muted text-muted-foreground hover:text-foreground transition-colors shrink-0 whitespace-nowrap disabled:opacity-40 disabled:cursor-not-allowed"
                    :class="{ 'border-primary/50 text-primary bg-primary/5 font-medium': hostedMsgState.keepNum === preset.val }"
                    :disabled="!hostedMsgState.enabled || isDemo"
                    @click="hostedMsgState.keepNum = preset.val"
                  >
                    {{ preset.label }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- 底部操作与说明 (弹性布局，杜绝窄屏贴边推挤) -->
          <div class="pt-3 border-t border-border/60 flex items-center justify-between gap-2">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger as-child>
                  <button type="button" class="text-xs text-muted-foreground hover:text-foreground flex items-center gap-1 transition-colors shrink-0 py-1">
                    <HelpCircleIcon class="w-3.5 h-3.5" />
                    <span class="text-[11px] sm:text-xs">执行规则</span>
                  </button>
                </TooltipTrigger>
                <TooltipContent class="max-w-xs text-xs">
                  <p>1. 数据按当前账号严格隔离，不会影响其他用户的托管消息。</p>
                  <p class="mt-1">2. 维护任务由系统守护进程统一调度，按时间倒序保留最新的 N 条。</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>

            <div class="flex items-center gap-2 shrink-0">
              <Button variant="outline" size="sm" class="h-8 px-2.5 sm:px-3 text-xs" @click="handleHostedMsgView">
                <ExternalLinkIcon class="w-3 h-3 mr-1 shrink-0" />
                <span>日志</span>
              </Button>
              <Button
                size="sm"
                class="h-8 px-3 sm:px-3.5 text-xs"
                :disabled="isDemo || hostedMsgState.saving"
                @click="handleHostedMsgSubmit"
              >
                {{ hostedMsgState.saving ? '保存中...' : '保存配置' }}
              </Button>
            </div>
          </div>
        </div>
      </div>
    </CardContent>
  </Card>
</template>
