<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { AlertTriangle, AlertCircle, CheckCircle2, Info, ArrowLeft, RotateCw } from 'lucide-vue-next'
import { CONSTANT } from '@/constant'

const route = useRoute()
const router = useRouter()

// 解析 URL 参数（兼容 search 和 hash query）
function getParam(key: string): string {
  // 1. 从 vue-router query 获取
  if (route.query && route.query[key]) {
    return String(route.query[key])
  }
  // 2. 从 search 获取
  if (typeof window !== 'undefined') {
    const searchParams = new URLSearchParams(window.location.search)
    if (searchParams.get(key)) {
      return searchParams.get(key) || ''
    }
    // 3. 从 hash query 获取
    const hash = window.location.hash || ''
    if (hash.includes('?')) {
      const hashParams = new URLSearchParams(hash.substring(hash.indexOf('?') + 1))
      if (hashParams.get(key)) {
        return hashParams.get(key) || ''
      }
    }
  }
  return ''
}

const type = computed(() => {
  const t = getParam('type') || ''
  if (['warning', 'error', 'info', 'success'].includes(t)) {
    return t
  }
  if (getParam('error')) return 'warning'
  return 'warning'
})

const title = computed(() => {
  const t = getParam('title')
  if (t) return decodeSafe(t)
  if (type.value === 'warning' || type.value === 'error') {
    return '单点登录提示'
  }
  if (type.value === 'success') {
    return '操作成功'
  }
  return '提示信息'
})

const message = computed(() => {
  const msg = getParam('message') || getParam('error') || getParam('msg')
  if (msg) return decodeSafe(msg)
  return '未知提示信息，请返回登录页面重试。'
})

function decodeSafe(str: string): string {
  try {
    return decodeURIComponent(str)
  } catch (e) {
    return str
  }
}

// 返回登录页
function goToLogin() {
  // 清理可能导致冲突的临时登录信息
  if (typeof window !== 'undefined') {
    localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME)
  }
  router.push('/login')
}

// 刷新重试
function refreshPage() {
  if (typeof window !== 'undefined') {
    window.location.reload()
  }
}
</script>

<script lang="ts">
export default {
  name: 'NoticePage'
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-muted/40 selection:bg-primary/20">
    <div class="w-full max-w-md">
      <Card class="border border-border shadow-xl bg-card">
        <CardHeader class="text-center pb-4 pt-8">
          <!-- 状态图标 -->
          <div class="mx-auto w-16 h-16 rounded-2xl flex items-center justify-center mb-3 transition-transform hover:scale-105 duration-200"
            :class="[
              type === 'warning' ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/30' : '',
              type === 'error' ? 'bg-destructive/10 text-destructive border border-destructive/30' : '',
              type === 'info' ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/30' : '',
              type === 'success' ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30' : '',
            ]"
          >
            <AlertTriangle v-if="type === 'warning'" class="w-8 h-8" />
            <AlertCircle v-else-if="type === 'error'" class="w-8 h-8" />
            <Info v-else-if="type === 'info'" class="w-8 h-8" />
            <CheckCircle2 v-else-if="type === 'success'" class="w-8 h-8" />
          </div>

          <CardTitle class="text-xl font-bold tracking-tight text-foreground">
            {{ title }}
          </CardTitle>
          <CardDescription class="text-xs text-muted-foreground mt-1">
            系统身份安全验证反馈信息
          </CardDescription>
        </CardHeader>

        <CardContent class="space-y-5 px-6 pb-8">
          <!-- 详细提示内容展示框 -->
          <div
            class="p-4 rounded-xl border leading-relaxed text-sm transition-colors"
            :class="[
              type === 'warning' ? 'bg-amber-500/5 border-amber-500/20 text-foreground' : '',
              type === 'error' ? 'bg-destructive/5 border-destructive/20 text-foreground' : '',
              type === 'info' ? 'bg-blue-500/5 border-blue-500/20 text-foreground' : '',
              type === 'success' ? 'bg-emerald-500/5 border-emerald-500/20 text-foreground' : '',
            ]"
          >
            <div class="flex items-start gap-2.5">
              <span class="text-base select-none shrink-0 mt-0.5">💡</span>
              <div class="space-y-1">
                <p class="font-medium text-sm leading-normal break-words">
                  {{ message }}
                </p>
              </div>
            </div>
          </div>

          <!-- 解决指引说明卡片 -->
          <div class="p-3.5 rounded-lg border border-border bg-muted/30 text-xs text-muted-foreground space-y-1.5">
            <div class="font-semibold text-foreground flex items-center gap-1.5">
              <span>常见解决方式</span>
            </div>
            <ul class="list-disc pl-4 space-y-1 leading-normal">
              <li>若初次使用此第三方账号登录，请联系系统管理员在后台开启「首次登录自动创建账号」；</li>
              <li>若已有系统账号，请联系管理员为您预先关联此身份提供商账号；</li>
              <li>或者使用用户名和独立密码直接登录系统。</li>
            </ul>
          </div>

          <!-- 操作按钮区域 -->
          <div class="flex flex-col sm:flex-row gap-2.5 pt-2">
            <Button
              class="flex-1 gap-2 font-medium"
              @click="goToLogin"
            >
              <ArrowLeft class="w-4 h-4" />
              <span>返回登录页面</span>
            </Button>
            <Button
              variant="outline"
              class="gap-1.5 font-medium sm:w-auto"
              @click="refreshPage"
            >
              <RotateCw class="w-3.5 h-3.5 text-muted-foreground" />
              <span>刷新重试</span>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
