<script setup lang="ts">
import { reactive, onMounted } from 'vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet'
import { Github, FileText } from 'lucide-vue-next'
import { request } from '@/api/api'
import { toast } from 'vue-sonner'

const state = reactive({
  version: '1.0.0',
  description: '一个现代化的消息推送管理平台，支持多种推送渠道和灵活的消息管理功能。',
  features: [
    '多渠道消息推送',
    '定时消息管理', 
    '托管消息服务',
    '发信日志追踪',
    '渠道配置管理'
  ],
  techStack: ['Vue 3', 'TypeScript', 'Vite', 'Tailwind CSS', 'Shadcn/ui'],
  githubUrl: 'https://github.com/engigu/Message-Push-Nest',
  copyright: '保留所有权利.',
  versionLog: ''
})

// 获取关于页面配置
const getAboutConfig = async () => {
  try {
    const params = { params: { section: 'about' } }
    const response = await request.get('/settings/getsetting', params)
    if (response.data.code === 200) {
      const data = response.data.data
      if (data.version) state.version = data.version
      if (data.desc) state.versionLog = data.desc
    }
  } catch (error) {
    toast.error('获取关于信息失败')
  }
}

onMounted(() => {
  getAboutConfig()
})
</script>

<script lang="ts">
export default {
  name: 'AboutSettings'
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>站点关于</CardTitle>
      <CardDescription>{{ state.description }}</CardDescription>
    </CardHeader>
    <CardContent class="space-y-6">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div class="space-y-4">
          <div>
            <h3 class="font-medium text-gray-900 mb-2">技术栈</h3>
            <div class="flex flex-wrap gap-2">
              <Badge v-for="tech in state.techStack" :key="tech">{{ tech }}</Badge>
            </div>
          </div>
          
          <div>
            <h3 class="font-medium text-gray-900 mb-2">功能特性</h3>
            <ul class="text-sm text-gray-600 space-y-1">
              <li v-for="feature in state.features" :key="feature">• {{ feature }}</li>
            </ul>
          </div>
        </div>
        
        <div class="space-y-4">
          <div>
            <h3 class="font-medium text-gray-900 mb-2">系统信息</h3>
            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-gray-600">版本:</span>
                <Badge variant="outline">{{ state.version }}</Badge>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-600">运行环境:</span>
                <span>Vue 3 + TypeScript</span>
              </div>
            </div>
          </div>
          
          <div>
            <h3 class="font-medium text-gray-900 mb-2">版本日志</h3>
            <Sheet>
              <SheetTrigger as-child>
                <Button variant="outline" size="sm" class="inline-flex items-center gap-2">
                  <FileText class="w-4 h-4" />
                  查看更新日志
                </Button>
              </SheetTrigger>
              <SheetContent class="w-[600px] sm:w-[800px]">
                <SheetHeader>
                  <SheetTitle>版本更新日志</SheetTitle>
                </SheetHeader>
                <div class="mt-6">
                  <div class="bg-gray-50 p-4 rounded-lg">
                    <pre class="whitespace-pre-wrap text-sm font-mono">{{ state.versionLog }}</pre>
                  </div>
                </div>
              </SheetContent>
            </Sheet>
          </div>
        </div>
      </div>
      
      <div class="border-t border-gray-200 my-4"></div>
      
      <div class="text-center text-sm text-gray-500">
        <p>© {{ new Date().getFullYear() }} {{ state.copyright }}</p>
        <p class="mt-1">如有问题请联系系统管理员</p>
        <p class="mt-2">
          <a :href="state.githubUrl" target="_blank" class="inline-flex items-center gap-1 text-blue-500 hover:text-blue-700 underline">
            <Github class="w-4 h-4" />
            GitHub 仓库
          </a>
        </p>
      </div>
    </CardContent>
  </Card>
</template>