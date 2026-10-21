<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Fingerprint, Zap, KeyRound, ShieldCheck, Unlink, RefreshCw, CheckCircle2, Plus } from 'lucide-vue-next'
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
  <div class="space-y-6">
    <!-- 修改密码卡片 -->
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <KeyRound class="w-5 h-5 text-primary" />
          <span>修改密码</span>
        </CardTitle>
        <CardDescription>定期更新您的本地账号登录密码，增强账户安全</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="space-y-2">
          <Label for="current-password">旧密码</Label>
          <Input
            id="current-password"
            type="password"
            v-model="passwordForm.currentPassword"
            placeholder="请输入当前账号的旧密码"
          />
        </div>
        <div class="space-y-2">
          <Label for="new-password">新密码</Label>
          <Input
            id="new-password"
            type="password"
            v-model="passwordForm.newPassword"
            placeholder="请输入新密码（至少 4 位）"
          />
        </div>

        <Button @click="resetPassword" :disabled="pwdLoading" class="w-full sm:w-auto">
          {{ pwdLoading ? '提交中...' : '确认修改密码' }}
        </Button>
      </CardContent>
    </Card>

    <!-- 单点登录 (SSO) 多身份源绑定卡片 -->
    <Card>
      <CardHeader>
        <div class="flex items-center justify-between">
          <CardTitle class="flex items-center gap-2">
            <ShieldCheck class="w-5 h-5 text-emerald-500" />
            <span>第三方单点登录 (SSO) 绑定</span>
          </CardTitle>
          <Button variant="ghost" size="sm" @click="fetchSSOProfile" :disabled="ssoLoading" class="h-8 px-2 text-xs">
            <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': ssoLoading }" />
          </Button>
        </div>
        <CardDescription>
          支持同时绑定多个单点登录身份源（例如内网企业 OIDC 与个人 GitHub OAuth 2.0）。绑定后无论通过哪一渠道单点登录，均可一键直接进入当前账号。
        </CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div v-if="availableProviders && availableProviders.length > 0" class="grid gap-3">
          <div
            v-for="p in availableProviders"
            :key="p.id"
            class="p-4 rounded-xl border transition-all flex flex-col sm:flex-row sm:items-center justify-between gap-3"
            :class="getBindingForProvider(p.type) ? 'border-emerald-500/25 bg-emerald-500/5' : 'border-border bg-muted/20'"
          >
            <div class="flex items-start sm:items-center gap-3">
              <div
                class="w-9 h-9 rounded-lg flex items-center justify-center shrink-0 shadow-sm"
                :class="getBindingForProvider(p.type) ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400' : 'bg-muted text-muted-foreground'"
              >
                <Zap v-if="p.type === 'oauth2'" class="w-5 h-5 text-amber-500" />
                <Fingerprint v-else class="w-5 h-5 text-primary" />
              </div>

              <div class="space-y-0.5">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-semibold text-foreground">{{ p.name }}</span>
                  <Badge
                    v-if="getBindingForProvider(p.type)"
                    variant="secondary"
                    class="bg-emerald-600/15 text-emerald-700 dark:text-emerald-400 border border-emerald-500/30 text-[11px] py-0 px-1.5 flex items-center gap-1"
                  >
                    <CheckCircle2 class="w-3 h-3 text-emerald-600" />
                    <span>已绑定</span>
                  </Badge>
                  <Badge
                    v-else
                    variant="outline"
                    class="text-[11px] py-0 px-1.5 text-muted-foreground"
                  >
                    未绑定
                  </Badge>
                </div>

                <p v-if="getBindingForProvider(p.type)" class="text-xs text-muted-foreground font-mono">
                  身份标识 (Sub): {{ getBindingForProvider(p.type)?.subject }}
                </p>
                <p v-else class="text-xs text-muted-foreground/80">
                  尚未绑定该身份源，绑定后可在登录页通过 {{ p.name }} 快速登录本账号
                </p>
              </div>
            </div>

            <!-- 操作按钮 -->
            <div class="flex items-center gap-2 self-end sm:self-center">
              <Button
                v-if="getBindingForProvider(p.type)"
                variant="outline"
                size="sm"
                class="text-rose-600 hover:text-rose-700 hover:bg-rose-50 dark:hover:bg-rose-950/30 border-rose-200 dark:border-rose-900/50 gap-1.5 h-8 text-xs"
                @click="handleUnbindSSO(p)"
                :disabled="ssoLoading"
              >
                <Unlink class="w-3.5 h-3.5" />
                <span>解除绑定</span>
              </Button>

              <Button
                v-else
                variant="default"
                size="sm"
                class="gap-1.5 h-8 text-xs font-medium"
                @click="handleBindSSO(p)"
                :disabled="ssoLoading"
              >
                <Plus class="w-3.5 h-3.5" />
                <span>立即绑定</span>
              </Button>
            </div>
          </div>
        </div>

        <div v-else class="text-xs text-muted-foreground/70 italic p-3 text-center rounded-lg border border-dashed border-border bg-muted/10">
          系统暂未开启任何第三方单点登录通道（管理员可在左侧「单点登录」设置中开启 OIDC 或 OAuth 2.0 提供商）
        </div>
      </CardContent>
    </Card>
  </div>
</template>