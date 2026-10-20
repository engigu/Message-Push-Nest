<script setup lang="ts">
import { reactive, onMounted, ref } from 'vue'
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
import { HelpCircleIcon, ShuffleIcon } from 'lucide-vue-next'
import { checkIsDemoMode } from '@/util/demo'

const router = useRouter()
const isDemo = ref(checkIsDemoMode())

// 日志清理状态
const logsState = reactive({
  section: 'log_config',
  cron: '',
  keepNum: '1000',
  enabled: true,
})

// 托管消息清理状态
const hostedMsgState = reactive({
  section: 'hosted_msg_config',
  cron: '',
  keepNum: '50000',
  enabled: false,
})

// 随机生成凌晨低峰期的 Cron 表达式（1:00 ~ 4:59）
const randomizeLogsCron = () => {
  if (isDemo.value || checkIsDemoMode()) {
    toast.error('演示模式下禁止修改清理配置')
    return
  }
  const minute = Math.floor(Math.random() * 60)
  const hour = 1 + Math.floor(Math.random() * 4)
  logsState.cron = `${minute} ${hour} * * *`
  toast.info(`已随机分配日志清理时间：每天 ${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`)
}

const randomizeHostedMsgCron = () => {
  if (isDemo.value || checkIsDemoMode()) {
    toast.error('演示模式下禁止修改清理配置')
    return
  }
  const minute = Math.floor(Math.random() * 60)
  const hour = 1 + Math.floor(Math.random() * 4)
  hostedMsgState.cron = `${minute} ${hour} * * *`
  toast.info(`已随机分配托管消息清理时间：每天 ${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`)
}

// 提交日志清理配置
const handleLogsSubmit = async () => {
  if (isDemo.value || checkIsDemoMode()) {
    toast.error('演示模式下禁止修改清理配置')
    return
  }
  try {
    const postData = {
      section: logsState.section,
      data: {
        cron: logsState.cron.trim(),
        keep_num: logsState.keepNum.trim(),
        enabled: logsState.enabled ? 'true' : 'false',
      },
    }
    const response = await request.post('/settings/set', postData)
    if (response.data.code === 200) {
      const statusText = logsState.enabled ? '已打开' : '已关闭'
      toast.success(`日志清理${statusText}`)
    }
  } catch (error) {
    toast.error('保存失败，请稍后重试')
  }
}

// 提交托管消息清理配置
const handleHostedMsgSubmit = async () => {
  if (isDemo.value || checkIsDemoMode()) {
    toast.error('演示模式下禁止修改清理配置')
    return
  }
  try {
    const postData = {
      section: hostedMsgState.section,
      data: {
        cron: hostedMsgState.cron.trim(),
        keep_num: hostedMsgState.keepNum.trim(),
        enabled: hostedMsgState.enabled ? 'true' : 'false',
      },
    }
    const response = await request.post('/settings/set', postData)
    if (response.data.code === 200) {
      const statusText = hostedMsgState.enabled ? '已打开' : '已关闭'
      toast.success(`托管消息清理${statusText}`)
    }
  } catch (error) {
    toast.error('保存失败，请稍后重试')
  }
}

// 查看日志清理日志
const handleLogsView = () => {
  router.push({ path: '/sendlogs', query: { taskid: CONSTANT.LOG_TASK_ID } })
}

// 查看托管消息清理日志
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
      // 使用Object.assign确保响应式更新
      Object.assign(logsState, {
        cron: data.cron || '',
        keepNum: data.keep_num || '1000',
        enabled: data.enabled === 'true' || data.enabled === true
      })
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
      // 使用Object.assign确保响应式更新
      Object.assign(hostedMsgState, {
        cron: data.cron || '',
        keepNum: data.keep_num || '50000',
        enabled: data.enabled === 'true' || data.enabled === true
      })
    }
  } catch (error) {
    toast.error('获取托管消息清理配置失败')
  }
}

onMounted(() => {
  getLogsConfig()
  getHostedMsgConfig()
})
</script>

<script lang="ts">
export default {
  name: 'CleanSettings'
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>数据清理设置</CardTitle>
      <CardDescription>配置定时数据清除和保留策略</CardDescription>
    </CardHeader>
    <CardContent>
      <!-- 演示模式提示横幅 -->
      <div v-if="isDemo" class="mb-6 p-3 rounded-lg border border-amber-500/30 bg-amber-500/10 text-xs text-amber-700 dark:text-amber-400 flex items-center gap-2">
        <span>⚠️ 当前系统处于演示模式，数据清理配置仅供只读预览展示，禁止修改配置。</span>
      </div>

      <!-- 大屏横向，小屏竖向 -->
      <div class="flex flex-col lg:flex-row gap-6 lg:gap-8">
        <!-- 日志清理部分 -->
        <div class="setting-section flex-1">
          <div class="flex items-center space-x-2 mb-4">
            <h3 class="text-lg font-semibold">日志清理</h3>
          </div>
          
          <div class="space-y-4">
            <!-- 启用开关 -->
            <div class="flex items-center justify-between space-x-2 p-4 border border-border rounded-lg bg-muted/20">
              <div class="space-y-0.5">
                <Label class="text-base font-medium text-foreground">启用日志清理</Label>
                <div class="text-sm text-muted-foreground">
                  开启后将按照规则清理日志
                </div>
              </div>
              <Switch v-model="logsState.enabled" :disabled="isDemo" @update:model-value="handleLogsSubmit" />
            </div>

            <!-- Cron表达式输入 -->
            <div class="space-y-2">
              <div class="flex items-center justify-between">
                <label class="text-sm font-medium text-foreground">定时清除Cron表达式</label>
                <button
                  type="button"
                  class="text-xs text-primary hover:underline flex items-center gap-1 disabled:opacity-40 disabled:cursor-not-allowed"
                  :disabled="!logsState.enabled || isDemo"
                  :title="isDemo ? '演示模式已禁止修改' : '随机生成凌晨低峰期清理时间'"
                  @click="randomizeLogsCron"
                >
                  <ShuffleIcon class="w-3.5 h-3.5" />
                  <span>随机时间</span>
                </button>
              </div>
              <Input 
                v-model="logsState.cron" 
                placeholder="留空自动随机生成，或输入如 15 2 * * *"
                :disabled="!logsState.enabled || isDemo"
              />
            </div>
            
            <!-- 保留数量输入 -->
            <div class="space-y-2">
              <label class="text-sm font-medium text-foreground">保留日志条数</label>
              <Input 
                v-model="logsState.keepNum" 
                placeholder="请输入要保留的最近的日志条数"
                :disabled="!logsState.enabled || isDemo"
              />
            </div>

            <!-- 底部操作区域 -->
            <div class="flex items-center justify-between pt-2">
              <div class="flex items-center space-x-2">
                <span class="text-sm text-muted-foreground">说明</span>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <HelpCircleIcon class="w-4 h-4 text-muted-foreground/70 hover:text-foreground transition-colors" />
                    </TooltipTrigger>
                    <TooltipContent class="max-w-xs">
                      <div class="text-sm">
                        <p>为避免并发冲击，系统已为每位用户默认随机打散在凌晨低峰期（1:00~4:59）执行</p>
                        <p class="mt-1">点击“随机时间”可重新打散分配，亦可手动指定或留空自动随机分配</p>
                        <p class="mt-1">保留数目如果不设置，默认保留最近1000条</p>
                        <p class="mt-1">需要先启用开关才能生效</p>
                      </div>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
              
              <div class="flex space-x-2">
                <Button variant="outline" size="sm" @click="handleLogsView">
                  查看日志
                </Button>
                <Button
                  size="sm"
                  :disabled="isDemo"
                  :class="{ 'opacity-50 cursor-not-allowed': isDemo }"
                  :title="isDemo ? '演示模式已禁止修改清理配置' : '保存'"
                  @click="handleLogsSubmit"
                >
                  保存
                </Button>
              </div>
            </div>
          </div>
        </div>

        <!-- 响应式分隔线：小屏横向，大屏竖向 -->
        <div class="lg:hidden w-full border-t border-border my-6"></div>
        <div class="hidden lg:block w-px bg-border self-stretch"></div>

        <!-- 托管消息清理部分 -->
        <div class="setting-section flex-1">
          <div class="flex items-center space-x-2 mb-4">
            <h3 class="text-lg font-semibold text-foreground">托管消息清理</h3>
          </div>
          
          <div class="space-y-4">
            <!-- 启用开关 -->
            <div class="flex items-center justify-between space-x-2 p-4 border border-border rounded-lg bg-muted/20">
              <div class="space-y-0.5">
                <Label class="text-base font-medium text-foreground">启用托管消息清理</Label>
                <div class="text-sm text-muted-foreground">
                  开启后将按照规则清理托管消息
                </div>
              </div>
              <Switch v-model="hostedMsgState.enabled" :disabled="isDemo" @update:model-value="handleHostedMsgSubmit" />
            </div>

            <!-- Cron表达式输入 -->
            <div class="space-y-2">
              <div class="flex items-center justify-between">
                <label class="text-sm font-medium text-foreground">定时清除Cron表达式</label>
                <button
                  type="button"
                  class="text-xs text-primary hover:underline flex items-center gap-1 disabled:opacity-40 disabled:cursor-not-allowed"
                  :disabled="!hostedMsgState.enabled || isDemo"
                  :title="isDemo ? '演示模式已禁止修改' : '随机生成凌晨低峰期清理时间'"
                  @click="randomizeHostedMsgCron"
                >
                  <ShuffleIcon class="w-3.5 h-3.5" />
                  <span>随机时间</span>
                </button>
              </div>
              <Input 
                v-model="hostedMsgState.cron" 
                placeholder="留空自动随机生成，或输入如 30 2 * * *"
                :disabled="!hostedMsgState.enabled || isDemo"
              />
            </div>
            
            <!-- 保留数量输入 -->
            <div class="space-y-2">
              <label class="text-sm font-medium text-foreground">保留托管消息条数</label>
              <Input 
                v-model="hostedMsgState.keepNum" 
                placeholder="请输入要保留的最近的托管消息条数"
                :disabled="!hostedMsgState.enabled || isDemo"
              />
            </div>

            <!-- 底部操作区域 -->
            <div class="flex items-center justify-between pt-2">
              <div class="flex items-center space-x-2">
                <span class="text-sm text-muted-foreground">说明</span>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <HelpCircleIcon class="w-4 h-4 text-muted-foreground/70 hover:text-foreground transition-colors" />
                    </TooltipTrigger>
                    <TooltipContent class="max-w-xs">
                      <div class="text-sm">
                        <p>为避免并发冲击，系统已为每位用户默认随机打散在凌晨低峰期（1:00~4:59）执行</p>
                        <p class="mt-1">点击“随机时间”可重新打散分配，亦可手动指定或留空自动随机分配</p>
                        <p class="mt-1">保留数目如果不设置，默认保留最近5000条</p>
                        <p class="mt-1">需要先启用开关才能生效</p>
                      </div>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
              
              <div class="flex space-x-2">
                <Button variant="outline" size="sm" @click="handleHostedMsgView">
                  查看日志
                </Button>
                <Button
                  size="sm"
                  :disabled="isDemo"
                  :class="{ 'opacity-50 cursor-not-allowed': isDemo }"
                  :title="isDemo ? '演示模式已禁止修改清理配置' : '保存'"
                  @click="handleHostedMsgSubmit"
                >
                  保存
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </CardContent>
  </Card>
</template>
