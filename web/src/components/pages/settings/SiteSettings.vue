<script setup lang="ts">
import { reactive, onMounted, ref } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
// @ts-ignore
import { LocalStieConfigUtils } from '@/util/localSiteConfig'
import { HelpCircleIcon, CheckIcon } from 'lucide-vue-next'
import { THEMES, applyTheme, getStoredTheme } from '@/util/theme'

const currentThemeColor = ref(getStoredTheme())

const changeTheme = (themeKey: string) => {
  currentThemeColor.value = themeKey
  state.theme_color = themeKey
  applyTheme(themeKey)
}

const state = reactive({
  title: '',
  slogan: '',
  logo: '',
  pagesize: '',
  cookieExpDays: '',
  theme_color: getStoredTheme(),
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
        pagesize: state.pagesize.toString(),
        cookie_exp_days: state.cookieExpDays.toString(),
        theme_color: state.theme_color,
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
      state.cookieExpDays = data.cookie_exp_days || '1'
      state.theme_color = data.theme_color || 'blue'

      // 同步当前选中的主题状态
      currentThemeColor.value = state.theme_color
      applyTheme(state.theme_color)

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
            <label class="text-sm font-medium text-gray-700">站点图标(只支持svg文本)</label>
            <div class="flex items-center gap-2">
              <div class="flex-1">
                <Input v-model="state.logo" placeholder="请输入自定义的网站logo（svg文本）" />
              </div>
              <div v-if="state.logo"
                class="flex-shrink-0 w-9 h-9 border border-border rounded bg-white dark:bg-white flex items-center justify-center p-1.5 shadow-sm overflow-hidden"
                v-html="state.logo">
              </div>
            </div>
          </div>

          <!-- 主题色 -->
          <div class="space-y-3">
            <label class="text-sm font-medium text-gray-700">主题色</label>
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
              <button v-for="t in THEMES" :key="t.key" @click="changeTheme(t.key)"
                class="group relative flex items-center gap-2.5 p-2.5 rounded-lg border transition-all duration-200"
                :class="[
                  currentThemeColor === t.key
                    ? 'border-brand bg-brand/5 shadow-sm ring-1 ring-brand/20'
                    : 'border-border hover:border-brand/40 hover:bg-muted/30'
                ]">
                <div class="w-4 h-4 rounded-full shadow-inner border border-white/20 flex-shrink-0"
                  :style="{ backgroundColor: t.light }"></div>
                <span class="text-xs font-medium truncate"
                  :class="currentThemeColor === t.key ? 'text-brand' : 'text-foreground/80'">{{ t.name }}</span>

                <!-- 选中状态标志 -->
                <div v-if="currentThemeColor === t.key"
                  class="absolute -top-1 -right-1 w-3.5 h-3.5 rounded-full bg-brand text-white flex items-center justify-center shadow-sm">
                  <CheckIcon class="w-2 h-2" />
                </div>
              </button>
            </div>
          </div>

          <!-- 分页大小和Cookie过期天数 -->
          <div class="grid grid-cols-2 gap-4">
            <!-- 分页大小 -->
            <div class="space-y-2">
              <label class="text-sm font-medium text-gray-700">分页大小</label>
              <Input v-model="state.pagesize" placeholder="页面分页大小" />
            </div>

            <!-- Cookie过期天数 -->
            <div class="space-y-2">
              <label class="text-sm font-medium text-gray-700">Cookie过期天数</label>
              <Input v-model="state.cookieExpDays" type="number" min="1" max="365" placeholder="Cookie过期天数（默认1天）" />
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
                <TooltipContent class="max-w-sm">
                  <div class="text-sm space-y-1">
                    <p>1. logo请输入svg文本，替换后登录页面，ico，导航栏logo将全部一起更换</p>
                    <p>2. slogan将在登录页面展示</p>
                    <p>3. Cookie过期天数设置用户登录后的有效期，修改后下次登录时生效</p>
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
