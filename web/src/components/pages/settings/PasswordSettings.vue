<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Fingerprint, Zap, KeyRound, ShieldCheck, Unlink, RefreshCw, Plus } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
import { checkIsDemoMode } from '@/util/demo'
// @ts-ignore
import config from '../../../../config.js'

import { CONSTANT } from '@/constant'

// 重置密码相关
const passwordForm = ref({
  newPassword: '',
  currentPassword: ''
})
const pwdLoading = ref(false)

// SSO 绑定相关
interface SSOBinding {
  id: string
  provider: string
  subject: string
  email: string
  created_on: string
}

interface SSOProfile {
  id: string
  username: string
  email: string
  role: string
  is_oidc: boolean
  auth_provider: string
  oidc_subject: string
  bindings: SSOBinding[]
}

const ssoLoading = ref(false)
const ssoProfile = ref<SSOProfile | null>(null)
const availableProviders = ref<any[]>([])

// 获取当前用户 SSO 绑定状态
const fetchSSOProfile = async () => {
  ssoLoading.value = true
  try {
    const [profileRes, statusRes] = await Promise.all([
      request.get('/api/v1/auth/sso/profile'),
      request.get('/auth/sso/status')
    ])

    if (profileRes.data && profileRes.data.code === 200) {
      ssoProfile.value = profileRes.data.data
    }
    if (statusRes.data && statusRes.data.code === 200 && statusRes.data.data?.providers) {
      availableProviders.value = statusRes.data.data.providers
    }
  } catch (e: any) {
    console.warn('获取 SSO 绑定信息失败:', e)
  } finally {
    ssoLoading.value = false
  }
}

// 检查特定 Provider 是否已绑定
const getBindingForProvider = (providerType: string): SSOBinding | undefined => {
  if (!ssoProfile.value) return undefined
  if (ssoProfile.value.bindings && ssoProfile.value.bindings.length > 0) {
    return ssoProfile.value.bindings.find(b => b.provider === providerType)
  }
  // 兜底单绑定存量字段
  if (ssoProfile.value.is_oidc && ssoProfile.value.auth_provider === providerType) {
    return {
      id: 'legacy',
      provider: providerType,
      subject: ssoProfile.value.oidc_subject,
      email: ssoProfile.value.email,
      created_on: ''
    }
  }
  return undefined
}

// 发起指定 Provider 的绑定
const handleBindSSO = (provider: any) => {
  const prefix = config.pathPrefix || ''
  const token = localStorage.getItem(CONSTANT.STORE_TOKEN_NAME) || ''
  const bindUrl = `${prefix}/api/v1/auth/sso/bind?type=${provider.type}&token=${encodeURIComponent(token)}`
  window.location.href = bindUrl
}

// 解除指定 Provider 绑定
const handleUnbindSSO = async (provider: any) => {
  if (!confirm(`确定要解除与 [${provider.name}] 的单点登录绑定吗？解除后该平台无法再直接登录本账号。`)) {
    return
  }

  ssoLoading.value = true
  try {
    const res = await request.post('/api/v1/auth/sso/unbind', { provider: provider.type })
    if (res.data && res.data.code === 200) {
      toast.success(`已成功解除 [${provider.name}] 单点登录绑定`)
      fetchSSOProfile()
    } else {
      toast.error(res.data?.msg || '解除绑定失败')
    }
  } catch (e: any) {
    toast.error('请求失败: ' + (e?.message || '网络异常'))
  } finally {
    ssoLoading.value = false
  }
}

// 重置密码
const resetPassword = async () => {
  if (checkIsDemoMode()) {
    toast.error('演示模式下无法重置密码')
    return
  }

  if (!passwordForm.value.currentPassword.trim()) {
    toast.error('请输入旧密码')
    return
  }
  if (!passwordForm.value.newPassword.trim() || passwordForm.value.newPassword.length < 4) {
    toast.error('新密码长度至少为 4 位')
    return
  }

  pwdLoading.value = true
  try {
    const postData = {
      new_passwd: passwordForm.value.newPassword,
      old_passwd: passwordForm.value.currentPassword
    }
    const rsp = await request.post('/settings/setpasswd', postData)
    if (rsp.data.code == 200) {
      toast.success(rsp.data.msg || '密码修改成功')
      passwordForm.value.currentPassword = ''
      passwordForm.value.newPassword = ''
    } else {
      toast.error(rsp.data.msg || '密码修改失败')
    }
  } catch (error) {
    toast.error('密码修改失败，请稍后重试')
  } finally {
    pwdLoading.value = false
  }
}

onMounted(() => {
  fetchSSOProfile()

  // 检查 URL 中是否有绑定回调返回的参数
  const url = new URL(window.location.href)
  const hash = window.location.hash || ''
  let ssoBindStatus = url.searchParams.get('sso_bind')
  let ssoBindMsg = url.searchParams.get('msg')
  let ssoBindErr = url.searchParams.get('error')

  if (!ssoBindStatus && hash.includes('?')) {
    const hashParams = new URLSearchParams(hash.substring(hash.indexOf('?') + 1))
    ssoBindStatus = hashParams.get('sso_bind')
    ssoBindMsg = hashParams.get('msg')
    ssoBindErr = hashParams.get('error')
  }

  if (ssoBindStatus === 'success') {
    toast.success(ssoBindMsg ? decodeURIComponent(ssoBindMsg) : '已成功绑定第三方单点登录账号！')
    cleanUrlParam()
  } else if (ssoBindStatus === 'failed') {
    toast.error(ssoBindErr ? decodeURIComponent(ssoBindErr) : '绑定第三方账号失败')
    cleanUrlParam()
  }
})

const cleanUrlParam = () => {
  try {
    const cleanHash = window.location.hash.split('?')[0] || '#/settings'
    window.history.replaceState(null, '', `${window.location.pathname}${cleanHash}`)
  } catch (e) {}
}
</script>

<script lang="ts">
export default {
  name: 'PasswordSettings'
}
</script>

<template>
  <div class="space-y-4 sm:space-y-6">
    <!-- 修改密码卡片 -->
    <Card class="overflow-hidden shadow-xs">
      <CardHeader class="px-3.5 py-3.5 sm:px-6 sm:py-5 border-b border-border/40 bg-muted/[0.06]">
        <div class="flex items-center gap-2.5">
          <div class="w-7 h-7 sm:w-8 sm:h-8 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0">
            <KeyRound class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
          </div>
          <div class="min-w-0">
            <CardTitle class="text-sm sm:text-base font-semibold tracking-tight">修改本地密码</CardTitle>
            <CardDescription class="text-xs text-muted-foreground mt-0.5 text-pretty">
              定期更新系统本地登录密码，保障账户安全
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent class="p-3.5 sm:p-6 space-y-3.5 sm:space-y-4">
        <div class="space-y-1.5 sm:space-y-2">
          <Label for="current-password" class="text-xs font-medium">旧密码</Label>
          <Input
            id="current-password"
            type="password"
            v-model="passwordForm.currentPassword"
            placeholder="请输入当前账号的旧密码"
            autocomplete="current-password"
            class="h-9 text-xs"
          />
        </div>
        <div class="space-y-1.5 sm:space-y-2">
          <Label for="new-password" class="text-xs font-medium">新密码</Label>
          <Input
            id="new-password"
            type="password"
            v-model="passwordForm.newPassword"
            placeholder="请输入新密码（至少 4 位）"
            autocomplete="new-password"
            class="h-9 text-xs"
          />
        </div>

        <Button @click="resetPassword" :disabled="pwdLoading" class="w-full sm:w-auto h-8 sm:h-9 text-xs font-medium active:scale-95 transition-all">
          {{ pwdLoading ? '提交中...' : '确认修改密码' }}
        </Button>
      </CardContent>
    </Card>

    <!-- 单点登录 (SSO) 多身份源绑定卡片 -->
    <Card class="overflow-hidden shadow-xs">
      <CardHeader class="px-3.5 py-3.5 sm:px-6 sm:py-5 border-b border-border/40 bg-muted/[0.06]">
        <div class="flex items-center justify-between gap-2.5">
          <div class="flex items-center gap-2.5 min-w-0">
            <div class="w-7 h-7 sm:w-8 sm:h-8 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center shrink-0">
              <ShieldCheck class="w-3.5 h-3.5 sm:w-4 sm:h-4" />
            </div>
            <div class="min-w-0">
              <CardTitle class="text-sm sm:text-base font-semibold tracking-tight truncate">第三方单点登录 (SSO) 绑定</CardTitle>
              <CardDescription class="text-xs text-muted-foreground mt-0.5 text-pretty line-clamp-1 sm:line-clamp-none">
                绑定外部单点登录身份源，可通过第三方账号一键快捷登录
              </CardDescription>
            </div>
          </div>
          <Button
            variant="outline"
            size="sm"
            @click="fetchSSOProfile"
            :disabled="ssoLoading"
            class="h-7 sm:h-8 px-2 sm:px-2.5 text-xs text-muted-foreground hover:text-foreground border-border/70 gap-1.5 shrink-0 transition-all cursor-pointer"
            title="刷新绑定状态"
          >
            <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': ssoLoading }" />
            <span class="hidden sm:inline">刷新状态</span>
          </Button>
        </div>
      </CardHeader>

      <CardContent class="p-3.5 sm:p-6 space-y-3 sm:space-y-4">
        <!-- 身份源卡片列表 -->
        <div v-if="availableProviders && availableProviders.length > 0" class="grid gap-2.5 sm:gap-3.5">
          <div
            v-for="p in availableProviders"
            :key="p.id"
            class="group relative p-3 sm:p-4 rounded-xl border transition-all duration-200 flex flex-col sm:flex-row sm:items-center justify-between gap-3 sm:gap-4"
            :class="[
              getBindingForProvider(p.type)
                ? 'border-emerald-500/35 bg-emerald-500/[0.03] dark:bg-emerald-500/[0.05] hover:border-emerald-500/50'
                : 'border-border/80 bg-card/60 hover:border-primary/40 hover:bg-muted/[0.12] hover:shadow-xs'
            ]"
          >
            <div class="flex items-start sm:items-center gap-3 sm:gap-3.5 min-w-0">
              <!-- 身份源图标/徽标容器 -->
              <div
                class="w-9 h-9 sm:w-10 sm:h-10 rounded-lg sm:rounded-xl flex items-center justify-center shrink-0 shadow-xs transition-transform group-hover:scale-105"
                :class="[
                  getBindingForProvider(p.type)
                    ? 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30'
                    : p.type === 'oauth2'
                      ? 'bg-amber-500/10 text-amber-500 border border-amber-500/25'
                      : 'bg-primary/10 text-primary border border-primary/20'
                ]"
              >
                <Zap v-if="p.type === 'oauth2'" class="w-4 h-4 sm:w-5 sm:h-5 text-amber-500" />
                <Fingerprint v-else class="w-4 h-4 sm:w-5 sm:h-5" />
              </div>

              <!-- 身份源详情 -->
              <div class="space-y-0.5 sm:space-y-1 min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-1.5 sm:gap-2">
                  <span class="text-xs sm:text-sm font-semibold text-foreground tracking-tight">{{ p.name }}</span>
                  <span class="text-[9px] sm:text-[10px] uppercase font-mono px-1.5 py-0.5 rounded bg-muted text-muted-foreground font-medium">
                    {{ p.type }}
                  </span>

                  <!-- 绑定状态胶囊 -->
                  <span
                    v-if="getBindingForProvider(p.type)"
                    class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] sm:text-[11px] font-medium bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border border-emerald-500/30"
                  >
                    <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                    <span>已绑定</span>
                  </span>
                  <span
                    v-else
                    class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] sm:text-[11px] font-medium bg-muted text-muted-foreground border border-border/80"
                  >
                    <span class="w-1.5 h-1.5 rounded-full bg-muted-foreground/50"></span>
                    <span>未绑定</span>
                  </span>
                </div>

                <!-- 已绑定模式：显示 sub / email 信息 -->
                <div v-if="getBindingForProvider(p.type)" class="flex flex-wrap items-center gap-x-2.5 gap-y-1 text-[11px] sm:text-xs text-muted-foreground pt-0.5">
                  <span class="inline-flex items-center gap-1 font-mono text-[10px] sm:text-[11px] text-foreground bg-muted/60 px-1.5 py-0.5 rounded border border-border/60">
                    <span class="text-muted-foreground/70">Sub:</span>
                    <span class="truncate max-w-[160px] sm:max-w-[220px] select-all">{{ getBindingForProvider(p.type)?.subject }}</span>
                  </span>
                  <span v-if="getBindingForProvider(p.type)?.email" class="text-[10px] sm:text-[11px] text-muted-foreground/80 truncate max-w-[150px] sm:max-w-[200px]">
                    ✉️ {{ getBindingForProvider(p.type)?.email }}
                  </span>
                  <span v-if="getBindingForProvider(p.type)?.created_on" class="text-[10px] sm:text-[11px] text-muted-foreground/60 hidden sm:inline">
                    绑定于 {{ getBindingForProvider(p.type)?.created_on }}
                  </span>
                </div>

                <!-- 未绑定模式：简洁提示语 -->
                <p v-else class="text-[11px] sm:text-xs text-muted-foreground leading-relaxed pt-0.5 text-pretty">
                  尚未绑定该身份源，绑定后可在登录页通过 {{ p.name }} 授权一键进入本系统
                </p>
              </div>
            </div>

            <!-- 操作按钮 -->
            <div class="flex items-center gap-2 sm:shrink-0 self-stretch sm:self-center w-full sm:w-auto pt-2 sm:pt-0 border-t sm:border-0 border-border/40">
              <Button
                v-if="getBindingForProvider(p.type)"
                variant="outline"
                size="sm"
                class="w-full sm:w-auto h-8 px-3 text-xs text-rose-600 hover:text-rose-700 hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-rose-950/40 border-rose-200/80 dark:border-rose-900/60 gap-1.5 transition-all cursor-pointer"
                @click="handleUnbindSSO(p)"
                :disabled="ssoLoading"
              >
                <Unlink class="w-3.5 h-3.5" />
                <span>解除绑定</span>
              </Button>

              <Button
                v-else
                variant="outline"
                size="sm"
                class="w-full sm:w-auto h-8 px-3.5 text-xs font-medium border-primary/30 text-primary bg-primary/5 hover:bg-primary/15 hover:border-primary/50 gap-1.5 transition-all active:scale-95 cursor-pointer shadow-2xs"
                @click="handleBindSSO(p)"
                :disabled="ssoLoading"
              >
                <Plus class="w-3.5 h-3.5" />
                <span>立即绑定</span>
              </Button>
            </div>
          </div>
        </div>

        <!-- 暂无通道提示 -->
        <div
          v-else-if="!ssoLoading"
          class="text-xs text-muted-foreground p-5 sm:p-6 text-center rounded-xl border border-dashed border-border bg-muted/[0.06] space-y-2"
        >
          <div class="w-9 h-9 sm:w-10 sm:h-10 rounded-full bg-muted flex items-center justify-center mx-auto text-muted-foreground">
            <Fingerprint class="w-4 h-4 sm:w-5 sm:h-5 opacity-60" />
          </div>
          <p class="font-medium text-foreground text-xs sm:text-sm">暂未开放第三方单点登录通道</p>
          <p class="text-[11px] sm:text-xs text-muted-foreground/80 max-w-md mx-auto leading-relaxed text-pretty">
            管理员可在左侧「单点登录」设置中配置并启用 OIDC 或 OAuth 2.0 身份提供商。
          </p>
        </div>
      </CardContent>
    </Card>
  </div>
</template>