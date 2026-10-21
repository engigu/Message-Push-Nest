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
import { Switch } from '@/components/ui/switch'
import { checkIsDemoMode } from '@/util/demo'

const isDemo = ref(checkIsDemoMode())
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
  enable_public_preview: 'true',
  allow_registration: 'false',
  enable_captcha: 'true',
  section: 'site_config',
})

// 提交设置
const handleSubmit = async () => {
  if (isDemo.value || checkIsDemoMode()) {
    toast.error('当前处于演示模式，已禁止修改站点配置！')
    return
  }
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
        enable_public_preview: state.enable_public_preview,
        allow_registration: state.allow_registration,
        enable_captcha: state.enable_captcha,
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
  if (isDemo.value || checkIsDemoMode()) {
    toast.error('当前处于演示模式，已禁止恢复默认设置！')
    return
  }
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
      state.enable_public_preview = data.enable_public_preview !== undefined ? data.enable_public_preview : 'true'
      state.allow_registration = data.allow_registration !== undefined ? data.allow_registration : 'false'
      state.enable_captcha = data.enable_captcha !== undefined ? data.enable_captcha : 'true'

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
            <label class="text-sm font-medium text-foreground">站点标题</label>
            <Input v-model="state.title" placeholder="请输入自定义的网站标题" />
          </div>

          <!-- 站点标语 -->
          <div class="space-y-2">
            <label class="text-sm font-medium text-foreground">站点标语</label>
            <Input v-model="state.slogan" placeholder="请输入自定义的网站slogan" />
          </div>

          <!-- 站点图标 -->
          <div class="space-y-2">
            <label class="text-sm font-medium text-foreground">站点图标(只支持svg文本)</label>
            <div class="flex items-center gap-2">
              <div class="flex-1">
                <Input v-model="state.logo" placeholder="请输入自定义的网站logo（svg文本）" />
              </div>
              <div v-if="state.logo"
                class="flex-shrink-0 w-9 h-9 border border-border rounded bg-muted/50 flex items-center justify-center p-1.5 shadow-sm overflow-hidden"
                v-html="state.logo">
              </div>
            </div>
          </div>

          <!-- 主题色 -->
          <div class="space-y-3">
            <label class="text-sm font-medium text-foreground">主题色</label>
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
              <button v-for="t in THEMES" :key="t.key" @click="changeTheme(t.key)"
                class="group relative flex items-center gap-2.5 p-2.5 rounded-lg border transition-all duration-200"
                :class="[
                  currentThemeColor === t.key
                    ? 'border-brand bg-brand/10 dark:bg-brand/20 shadow-sm ring-1 ring-brand/30'
                    : 'border-border bg-card hover:border-brand/40 hover:bg-muted/40'
                ]">
                <div class="w-4 h-4 rounded-full shadow-inner border border-white/20 flex-shrink-0"
                  :style="{ backgroundColor: t.light }"></div>
                <span class="text-xs font-medium truncate"
                  :class="currentThemeColor === t.key ? 'text-brand font-semibold' : 'text-foreground/80 group-hover:text-foreground'">{{ t.name }}</span>

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
              <label class="text-sm font-medium text-foreground">分页大小</label>
              <Input v-model="state.pagesize" placeholder="页面分页大小" />
            </div>

            <!-- Cookie过期天数 -->
            <div class="space-y-2">
              <label class="text-sm font-medium text-foreground">Cookie过期天数</label>
              <Input v-model="state.cookieExpDays" type="number" min="1" max="365" placeholder="Cookie过期天数（默认1天）" />
            </div>
          </div>

          <!-- 是否启用公开预览 -->
          <div class="flex items-center justify-between p-3 border border-dashed border-border rounded-lg bg-muted/20">
            <div class="space-y-0.5">
              <label class="text-sm font-semibold text-foreground">公开预览功能</label>
              <p class="text-xs text-muted-foreground">允许未登录用户通过防遍历的公开 URL 访问预览托管消息</p>
            </div>
            <Switch 
              :modelValue="state.enable_public_preview === 'true'" 
              @update:modelValue="(val: boolean) => state.enable_public_preview = val ? 'true' : 'false'" 
            />
          </div>

          <!-- 开放用户注册 -->
          <div class="flex items-center justify-between p-3 border border-dashed border-border rounded-lg bg-muted/20">
            <div class="space-y-0.5">
              <label class="text-sm font-semibold text-foreground flex items-center gap-2">
                <span>开放用户注册</span>
                <span v-if="isDemo" class="text-[11px] font-normal text-amber-600 dark:text-amber-400 bg-amber-500/10 px-1.5 py-0.5 rounded border border-amber-500/20">
                  演示模式已强制关闭
                </span>
              </label>
              <p class="text-xs text-muted-foreground">
                {{ isDemo ? '当前系统处于演示模式，自主注册已全局强制禁用' : '允许访客在前台登录页自主注册账号（注册用户默认为普通用户，数据独立隔离）' }}
              </p>
            </div>
            <Switch 
              :disabled="isDemo"
              :modelValue="!isDemo && state.allow_registration === 'true'" 
              @update:modelValue="(val: boolean) => state.allow_registration = val ? 'true' : 'false'" 
            />
          </div>

          <!-- 人机验证码控制 -->
          <div class="flex items-center justify-between p-3 border border-dashed border-border rounded-lg bg-muted/20">
            <div class="space-y-0.5">
              <label class="text-sm font-semibold text-foreground">登录与注册人机验证码</label>
              <p class="text-xs text-muted-foreground">登录和注册时要求输入算术验证码，防御自动化机器人撞库与暴力破解攻击</p>
            </div>
            <Switch 
              :modelValue="state.enable_captcha === 'true'" 
              @update:modelValue="(val: boolean) => state.enable_captcha = val ? 'true' : 'false'" 
            />
          </div>

        </div>

        <!-- 底部操作区域 -->
        <div class="flex items-center justify-between mt-6">
          <div class="flex items-center space-x-2">
            <span class="text-sm text-muted-foreground">说明</span>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <HelpCircleIcon class="w-4 h-4 text-muted-foreground/70 hover:text-foreground transition-colors" />
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
