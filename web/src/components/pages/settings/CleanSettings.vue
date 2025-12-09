<script setup lang="ts">
import { reactive, onMounted } from 'vue'
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
import { HelpCircleIcon } from 'lucide-vue-next'

const router = useRouter()

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

// 提交日志清理配置
const handleLogsSubmit = async () => {
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
      <!-- 大屏横向，小屏竖向 -->
      <div class="flex flex-col lg:flex-row gap-6 lg:gap-8">
        <!-- 日志清理部分 -->
        <div class="setting-section flex-1">
          <div class="flex items-center space-x-2 mb-4">
            <h3 class="text-lg font-semibold">日志清理</h3>
          </div>
          
          <div class="space-y-4">
            <!-- 启用开关 -->
            <div class="flex items-center justify-between space-x-2 p-4 border rounded-lg">
              <div class="space-y-0.5">
                <Label class="text-base font-medium">启用日志清理</Label>
                <div class="text-sm text-muted-foreground">
                  开启后将按照规则清理清理日志
                </div>
              </div>
              <Switch v-model="logsState.enabled" @update:model-value="handleLogsSubmit" />
            </div>

            <!-- Cron表达式输入 -->
            <div class="space-y-2">
              <label class="text-sm font-medium text-gray-700">定时清除Cron表达式</label>
              <Input 
                v-model="logsState.cron" 
                placeholder="请输入定时日志清除的Cron表达式"
                :disabled="!logsState.enabled"
              />
            </div>
            
            <!-- 保留数量输入 -->
            <div class="space-y-2">
              <label class="text-sm font-medium text-gray-700">保留日志条数</label>
              <Input 
                v-model="logsState.keepNum" 
                placeholder="请输入要保留的最近的日志条数"
                :disabled="!logsState.enabled"
              />
            </div>

            <!-- 底部操作区域 -->
            <div class="flex items-center justify-between pt-2">
              <div class="flex items-center space-x-2">
                <span class="text-sm text-gray-600">说明</span>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <HelpCircleIcon class="w-4 h-4 text-gray-400 hover:text-gray-600" />
                    </TooltipTrigger>
                    <TooltipContent class="max-w-xs">
                      <div class="text-sm">
                        <p>cron如果不设置，默认是在每天的0点1分进行清理</p>
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
                <Button size="sm" @click="handleLogsSubmit">
                  保存
                </Button>
              </div>
            </div>
          </div>
        </div>

        <!-- 响应式分隔线：小屏横向，大屏竖向 -->
        <div class="lg:hidden w-full border-t border-gray-200 dark:border-gray-700 my-6"></div>
        <div class="hidden lg:block w-px bg-border self-stretch"></div>

        <!-- 托管消息清理部分 -->
        <div class="setting-section flex-1">
          <div class="flex items-center space-x-2 mb-4">
            <h3 class="text-lg font-semibold">托管消息清理</h3>
          </div>
          
          <div class="space-y-4">
            <!-- 启用开关 -->
            <div class="flex items-center justify-between space-x-2 p-4 border rounded-lg">
              <div class="space-y-0.5">
                <Label class="text-base font-medium">启用托管消息清理</Label>
                <div class="text-sm text-muted-foreground">
                  开启后将按照规则清理托管消息
                </div>
              </div>
              <Switch v-model="hostedMsgState.enabled" @update:model-value="handleHostedMsgSubmit" />
            </div>

            <!-- Cron表达式输入 -->
            <div class="space-y-2">
              <label class="text-sm font-medium text-gray-700">定时清除Cron表达式</label>
              <Input 
                v-model="hostedMsgState.cron" 
                placeholder="请输入定时托管消息清除的Cron表达式"
                :disabled="!hostedMsgState.enabled"
              />
            </div>
            
            <!-- 保留数量输入 -->
            <div class="space-y-2">
              <label class="text-sm font-medium text-gray-700">保留托管消息条数</label>
              <Input 
                v-model="hostedMsgState.keepNum" 
                placeholder="请输入要保留的最近的托管消息条数"
                :disabled="!hostedMsgState.enabled"
              />
            </div>

            <!-- 底部操作区域 -->
            <div class="flex items-center justify-between pt-2">
              <div class="flex items-center space-x-2">
                <span class="text-sm text-gray-600">说明</span>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <HelpCircleIcon class="w-4 h-4 text-gray-400 hover:text-gray-600" />
                    </TooltipTrigger>
                    <TooltipContent class="max-w-xs">
                      <div class="text-sm">
                        <p>cron如果不设置，默认是在每天的0点30分进行清理</p>
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
                <Button size="sm" @click="handleHostedMsgSubmit">
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
