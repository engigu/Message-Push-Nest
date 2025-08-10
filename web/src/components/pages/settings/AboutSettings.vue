<script setup lang="ts">
import { reactive, onMounted, computed } from 'vue'
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
    '渠道配置管理',
    '站点信息配置',
  ],
  techStack: ['Golang','Vue 3', 'TypeScript', 'Vite', 'Tailwind CSS', 'Shadcn/ui'],
  githubUrl: 'https://github.com/engigu/Message-Push-Nest',
  copyright: '保留所有权利.',
  versionLog: '',
  buildTime: ''
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

// 获取构建时间
const buildTime = computed(() => {
  try {
    return (globalThis as any).__BUILD_TIME__ || '开发模式 - 未构建'
  } catch {
    return '开发模式 - 未构建'
  }
})

onMounted(() => {
  getAboutConfig()
  state.buildTime = buildTime.value
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
            <div class="flex flex-wrap gap-2">
              <Badge v-for="feature in state.features" :key="feature" variant="secondary">{{ feature }}</Badge>
            </div>
          </div>
        </div>

        <div class="space-y-4">
          <div>
            <h3 class="font-medium text-gray-900 mb-2">系统信息</h3>
            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-gray-600">系统版本:</span>
                <Badge variant="outline">{{ state.version }}</Badge>
              </div>
              <!-- <div class="flex justify-between">
                <span class="text-gray-600">运行环境:</span>
                <span>Vue 3 + TypeScript</span>
              </div> -->
              <div class="flex justify-between">
                <span class="text-gray-600">构建时间:</span>
                <span>{{ state.buildTime.includes('开发模式') ? state.buildTime : new Date(state.buildTime).toLocaleString('zh-CN') }}</span>
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
                  <div class="bg-card text-card-foreground rounded-xl border shadow-sm p-6">
                    <div class="space-y-2 max-h-[70vh] overflow-y-auto">
                      <div v-for="(line, index) in state.versionLog.split('\n').reverse().filter(line => line.trim())" :key="index" 
                           class="flex items-start gap-3 p-3 rounded-lg border bg-background hover:bg-accent/50 transition-colors">
                        <div class="w-1.5 h-1.5 rounded-full bg-primary mt-2 shrink-0"></div>
                        <div class="text-sm text-foreground leading-relaxed font-mono">
                          {{ line }}
                        </div>
                      </div>
                      <div v-if="!state.versionLog || state.versionLog.trim() === ''" class="text-center py-8 text-muted-foreground">
                        <div class="text-lg mb-2">📝</div>
                        <div class="text-sm">暂无版本日志</div>
                      </div>
                    </div>
                  </div>
                </div>
              </SheetContent>
            </Sheet>
          </div>
        </div>
      </div>

      <div class="border-t border-gray-200 my-4"></div>

      <div class="text-center text-sm text-gray-500">
        <p>© {{ new Date().getFullYear() }} {{ state.copyright }}
          <a :href="state.githubUrl" target="_blank"
            class="inline-flex items-center gap-1 text-blue-500 hover:text-blue-700 underline ml-3">
            <Github class="w-4 h-4" />
            GitHub 仓库
          </a>
        </p>
        <p class="mt-1">如有问题请联系系统管理员</p>
     
      </div>
    </CardContent>
  </Card>
</template>