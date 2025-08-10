<script setup lang="ts">
import { reactive, onMounted } from 'vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
import { useRouter } from 'vue-router'
import { CONSTANT } from '@/constant'
import { HelpCircleIcon } from 'lucide-vue-next'

const router = useRouter()

const state = reactive({
  section: 'log_config',
  cron: '',
  keepNum: '1000',
})

// 提交配置
const handleSubmit = async () => {
  try {
    const postData = {
      section: state.section,
      data: {
        cron: state.cron.trim(),
        keep_num: state.keepNum.trim(),
      },
    }
    const response = await request.post('/settings/set', postData)
    if (response.data.code === 200) {
      const msg = response.data.msg
      toast.success(msg)
    }
  } catch (error) {
    toast.error('保存失败，请稍后重试')
  }
}

// 查看日志
const handleView = async () => {
  router.push({ path: '/sendlogs', query: { taskid: CONSTANT.LOG_TASK_ID } })
}

// 获取站点配置
const getSiteConfig = async () => {
  try {
    const params = { params: { section: 'log_config' } }
    const response = await request.get('/settings/getsetting', params)
    if (response.data.code === 200) {
      const data = response.data.data
      state.cron = data.cron || ''
      state.keepNum = data.keep_num || '1000'
    }
  } catch (error) {
    toast.error('获取配置失败')
  }
}

onMounted(() => {
  getSiteConfig()
})
</script>

<script lang="ts">
export default {
  name: 'LogsSettings'
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>日志设置</CardTitle>
      <CardDescription>配置定时日志清除和保留策略</CardDescription>
    </CardHeader>
    <CardContent>
      <div class="setting-container">
        <div class="space-y-4">
          <!-- Cron表达式输入 -->
          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700">定时清除Cron表达式</label>
            <div class="flex">
              <!-- <div class="inline-flex items-center px-3 text-sm text-gray-500 bg-gray-50 border border-r-0 border-gray-300 rounded-l-md">
                cron://
              </div> -->
              <Input 
                v-model="state.cron" 
                placeholder="请输入定时日志清除的Cron表达式"
              />
            </div>
          </div>
          
          <!-- 保留数量输入 -->
          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700">保留日志条数</label>
            <div class="flex">
              <!-- <div class="inline-flex items-center px-3 text-sm text-gray-500 bg-gray-50 border border-r-0 border-gray-300 rounded-l-md">
                保留数
              </div> -->
              <Input 
                v-model="state.keepNum" 
                placeholder="请输入要保留的最近的日志条数"
               
              />
            </div>
          </div>
        </div>
        
        <!-- 底部操作区域 -->
        <div class="flex items-center justify-between mt-6">
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
                  </div>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
          
          <div class="flex space-x-2">
            <Button variant="outline" size="sm" @click="handleView">
              查看日志
            </Button>
            <Button size="sm" @click="handleSubmit">
              确定
            </Button>
          </div>
        </div>
      </div>
    </CardContent>
  </Card>
</template>

