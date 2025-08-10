<script setup lang="ts">
import { reactive, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
// @ts-ignore
import { LocalStieConfigUtils } from '@/util/localSiteConfig'
import { HelpCircleIcon } from 'lucide-vue-next'

const state = reactive({
  title: '',
  slogan: '',
  logo: '',
  pagesize: '',
  section: 'site_config',
})

// 提交设置
const handleSubmit = async () => {
  try {
    const postData = {
      section: state.section,
      data: {
        title: state.title.trim(),
        slogan: state.slogan.trim(),
        logo: state.logo.trim(),
        pagesize: state.pagesize,
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

// 恢复默认设置
const handleSubmitReset = async () => {
  try {
    const response = await request.post('/settings/reset', {})
    if (response.data.code === 200) {
      const msg = response.data.msg
      toast.success(msg)
      // 重新获取设置
      await getSiteConfig()
    }
  } catch (error) {
    toast.error('恢复默认设置失败，请稍后重试')
  }
}

// 获取站点配置
const getSiteConfig = async () => {
  try {
    const params = { params: { section: 'site_config' } }
    const response = await request.get('/settings/getsetting', params)
    if (response.data.code === 200) {
      const data = response.data.data
      state.title = data.title || ''
      state.logo = data.logo || ''
      state.slogan = data.slogan || ''
      state.pagesize = data.pagesize || ''

      LocalStieConfigUtils.updateLocalConfig(data)
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
  name: 'SiteSettings'
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>站点设置</CardTitle>
      <CardDescription>配置站点基本信息和参数</CardDescription>
    </CardHeader>
    <CardContent>
      <div class="setting-container">
        <div class="space-y-4">
          <!-- 站点标题 -->
          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700">站点标题</label>
            <Input v-model="state.title" placeholder="请输入自定义的网站标题" />
          </div>

          <!-- 站点标语 -->
          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700">站点标语</label>
            <Input v-model="state.slogan" placeholder="请输入自定义的网站slogan" />
          </div>

          <!-- 站点图标 -->
          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700">站点图标</label>
            <Input v-model="state.logo" placeholder="请输入自定义的网站logo（svg文本）" />
            <!-- SVG预览 -->
            <div v-if="state.logo" class="mt-2 p-3 border border-gray-200 rounded-md bg-gray-50">
              <div class="text-xs text-gray-500 mb-2">预览效果：</div>
              <div class="flex items-center justify-center w-16 h-16 bg-white border border-gray-300 rounded"
                v-html="state.logo"></div>
            </div>
          </div>

          <!-- 分页大小 -->
          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-700">分页大小</label>
            <Input v-model="state.pagesize" placeholder="页面分页大小" />
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
                <TooltipContent class="max-w-sm">
                  <div class="text-sm space-y-1">
                    <p>1. logo请输入svg文本，替换后登录页面，ico，导航栏logo将全部一起更换</p>
                    <p>2. slogan将在登录页面展示</p>
                    <p>** 将在下一次登录的时候生效，如果不生效请在登录页面Ctrl+F5强制刷新</p>
                    <p>** logo将替换网页ico，登录页面logo，导航栏logo</p>
                  </div>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>

          <div class="flex space-x-2">
            <Button variant="outline" size="sm" @click="handleSubmitReset">
              恢复默认
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

