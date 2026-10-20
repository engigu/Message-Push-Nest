<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

defineOptions({
  name: 'OidcSettings'
})
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'
// @ts-ignore
import config from '../../../../config.js'
import {
  HelpCircleIcon,
  Loader2,
  Copy,
  Check,
  Eye,
  EyeOff,
  ExternalLink
} from 'lucide-vue-next'
import { checkIsDemoMode } from '@/util/demo'

const isDemo = ref(checkIsDemoMode())
const loading = ref(false)
const saving = ref(false)
const testingOidc = ref(false)
const testingOAuth2 = ref(false)
const showSecretOidc = ref(false)
const showSecretOAuth2 = ref(false)
const copiedOidc = ref(false)
const copiedOAuth2 = ref(false)

// 当前激活的选项卡 ('oidc' | 'oauth2')
const activeTab = ref<'oidc' | 'oauth2'>('oidc')

const form = ref({
  // 1. 标准 OIDC 通道
  oidc_enabled: false,
  oidc_issuer: '',
  oidc_client_id: '',
  oidc_client_secret: '',
  oidc_scopes: '',
  oidc_auto_create: false,

  // 2. 通用 OAuth 2.0 通道
  oauth2_enabled: false,
  oauth2_name: '',
  oauth2_auth_url: '',
  oauth2_token_url: '',
  oauth2_userinfo_url: '',
  oauth2_user_key: '',
  oauth2_client_id: '',
  oauth2_client_secret: '',
  oauth2_scopes: '',
  oauth2_auto_create: false,
})

// 计算标准回调地址
const oidcRedirectUrl = computed(() => {
  if (typeof window === 'undefined') return ''
  const base = window.location.origin
  const prefix = config.pathPrefix || ''
  return `${base}${prefix}/api/v1/auth/sso/oidc/callback`
})

const oauth2RedirectUrl = computed(() => {
  if (typeof window === 'undefined') return ''
  const base = window.location.origin
  const prefix = config.pathPrefix || ''
  return `${base}${prefix}/api/v1/auth/sso/oauth2/callback`
})

function cleanOidcIssuer() {
  if (form.value.oidc_issuer) {
    form.value.oidc_issuer = form.value.oidc_issuer
      .trim()
      .replace(/\/+\.well-known\/openid-configuration\/?$/i, '')
      .replace(/\/+$/, '')
  }
}

// 探测发现文档地址
const discoveryUrl = computed(() => {
  let iss = form.value.oidc_issuer.trim()
  if (!iss || (!iss.startsWith('http://') && !iss.startsWith('https://'))) return ''
  iss = iss.replace(/\/+\.well-known\/openid-configuration\/?$/i, '').replace(/\/+$/, '')
  return `${iss}/.well-known/openid-configuration`
})

// 常用 OAuth 2.0 平台预设
interface OAuth2Preset {
  name: string
  displayName: string
  authUrl: string
  tokenUrl: string
  userinfoUrl: string
  scopes: string
  userKey?: string
}

const oauth2Presets: OAuth2Preset[] = [
  {
    name: 'GitHub',
    displayName: 'GitHub',
    authUrl: 'https://github.com/login/oauth/authorize',
    tokenUrl: 'https://github.com/login/oauth/access_token',
    userinfoUrl: 'https://api.github.com/user',
    scopes: 'read:user user:email',
    userKey: 'id',
  },
  {
    name: 'Gitee (码云)',
    displayName: 'Gitee',
    authUrl: 'https://gitee.com/oauth/authorize',
    tokenUrl: 'https://gitee.com/oauth/token',
    userinfoUrl: 'https://gitee.com/api/v5/user',
    scopes: 'user_info emails',
    userKey: 'id',
  },
  {
    name: 'GitLab',
    displayName: 'GitLab',
    authUrl: 'https://gitlab.com/oauth/authorize',
    tokenUrl: 'https://gitlab.com/oauth/token',
    userinfoUrl: 'https://gitlab.com/api/v4/user',
    scopes: 'read_user',
    userKey: 'id',
  },
]

function applyOAuth2Preset(p: OAuth2Preset) {
  form.value.oauth2_name = p.displayName
  form.value.oauth2_auth_url = p.authUrl
  form.value.oauth2_token_url = p.tokenUrl
  form.value.oauth2_userinfo_url = p.userinfoUrl
  form.value.oauth2_scopes = p.scopes
  if (p.userKey) form.value.oauth2_user_key = p.userKey
  toast.success(`已自动填入 ${p.name} 平台端点`)
}

// 快捷 Scope 列表
const oidcCommonScopes = ['groups', 'roles', 'offline_access']
const oauth2CommonScopes = ['read:user', 'user:email', 'profile', 'email']

function addOidcScope(scope: string) {
  const current = form.value.oidc_scopes.trim()
  const list = current ? current.split(/[\s,]+/).filter(Boolean) : []
  if (!list.includes(scope)) {
    list.push(scope)
    form.value.oidc_scopes = list.join(' ')
  }
}

function addOAuth2Scope(scope: string) {
  const current = form.value.oauth2_scopes.trim()
  const list = current ? current.split(/[\s,]+/).filter(Boolean) : []
  if (!list.includes(scope)) {
    list.push(scope)
    form.value.oauth2_scopes = list.join(' ')
  }
}

async function copyUrl(url: string, type: 'oidc' | 'oauth2') {
  if (!url) return
  try {
    await navigator.clipboard.writeText(url)
    if (type === 'oidc') {
      copiedOidc.value = true
      setTimeout(() => (copiedOidc.value = false), 2000)
    } else {
      copiedOAuth2.value = true
      setTimeout(() => (copiedOAuth2.value = false), 2000)
    }
    toast.success('回调地址已复制到剪贴板')
  } catch (e) {
    toast.error('复制失败，请手动选取复制')
  }
}

async function loadSettings() {
  loading.value = true
  try {
    const params = { params: { section: 'security' } }
    const response = await request.get('/settings/getsetting', params)
    if (response.data.code === 200) {
      const res = response.data.data || {}
      form.value = {
        // OIDC
        oidc_enabled: res.oidc_enabled === 'true',
        oidc_issuer: res.oidc_issuer || '',
        oidc_client_id: res.oidc_client_id || '',
        oidc_client_secret: res.oidc_client_secret || '',
        oidc_scopes: res.oidc_scopes || '',
        oidc_auto_create: res.oidc_auto_create === 'true',

        // OAuth 2.0
        oauth2_enabled: res.oauth2_enabled === 'true',
        oauth2_name: res.oauth2_name || '',
        oauth2_auth_url: res.oauth2_auth_url || '',
        oauth2_token_url: res.oauth2_token_url || '',
        oauth2_userinfo_url: res.oauth2_userinfo_url || '',
        oauth2_user_key: res.oauth2_user_key || '',
        oauth2_client_id: res.oauth2_client_id || '',
        oauth2_client_secret: res.oauth2_client_secret || '',
        oauth2_scopes: res.oauth2_scopes || '',
        oauth2_auto_create: res.oauth2_auto_create === 'true',
      }
    }
  } catch (e: any) {
    toast.error(e?.message || '获取单点登录设置失败')
  } finally {
    loading.value = false
  }
}

async function save() {
  cleanOidcIssuer()

  // 校验 OIDC 配置
  if (form.value.oidc_enabled) {
    if (!form.value.oidc_issuer.trim()) {
      toast.error('已开启标准 OIDC，请填写 Issuer 服务提供商地址')
      activeTab.value = 'oidc'
      return
    }
    if (!form.value.oidc_client_id.trim()) {
      toast.error('已开启标准 OIDC，请填写 Client ID 客户端标识')
      activeTab.value = 'oidc'
      return
    }
  }

  // 校验 OAuth 2.0 配置
  if (form.value.oauth2_enabled) {
    if (!form.value.oauth2_auth_url.trim() || !form.value.oauth2_token_url.trim() || !form.value.oauth2_userinfo_url.trim()) {
      toast.error('已开启通用 OAuth 2.0，请填写完整的授权、令牌及用户信息端点')
      activeTab.value = 'oauth2'
      return
    }
    if (!form.value.oauth2_client_id.trim()) {
      toast.error('已开启通用 OAuth 2.0，请填写 Client ID 客户端标识')
      activeTab.value = 'oauth2'
      return
    }
  }

  if (isDemo.value || checkIsDemoMode()) {
    toast.error('当前处于演示模式，已禁止修改单点登录配置！')
    return
  }

  saving.value = true
  try {
    const postData = {
      section: 'security',
      data: {
        // OIDC
        oidc_enabled: form.value.oidc_enabled ? 'true' : 'false',
        oidc_issuer: form.value.oidc_issuer.trim(),
        oidc_client_id: form.value.oidc_client_id.trim(),
        oidc_client_secret: form.value.oidc_client_secret,
        oidc_scopes: form.value.oidc_scopes.trim(),
        oidc_auto_create: form.value.oidc_auto_create ? 'true' : 'false',

        // OAuth 2.0
        oauth2_enabled: form.value.oauth2_enabled ? 'true' : 'false',
        oauth2_name: form.value.oauth2_name.trim(),
        oauth2_auth_url: form.value.oauth2_auth_url.trim(),
        oauth2_token_url: form.value.oauth2_token_url.trim(),
        oauth2_userinfo_url: form.value.oauth2_userinfo_url.trim(),
        oauth2_user_key: form.value.oauth2_user_key.trim(),
        oauth2_client_id: form.value.oauth2_client_id.trim(),
        oauth2_client_secret: form.value.oauth2_client_secret,
        oauth2_scopes: form.value.oauth2_scopes.trim(),
        oauth2_auto_create: form.value.oauth2_auto_create ? 'true' : 'false',
      },
    }

    const response = await request.post('/settings/set', postData)
    if (response.data.code === 200) {
      toast.success(response.data.msg || '单点登录配置已保存')
      await loadSettings()
    } else {
      toast.error(response.data.msg || '保存失败')
    }
  } catch (e: any) {
    toast.error(e?.message || '保存失败，请稍后重试')
  } finally {
    saving.value = false
  }
}

async function handleTestOidc() {
  if (isDemo.value || checkIsDemoMode()) {
    toast.error('当前处于演示模式，已禁止测试第三方连接！')
    return
  }
  cleanOidcIssuer()
  if (!form.value.oidc_issuer.trim()) {
    toast.error('请先填写 Issuer 服务提供商根地址')
    return
  }

  testingOidc.value = true
  try {
    const response = await request.post('/auth/oidc/test', {
      type: 'oidc',
      issuer: form.value.oidc_issuer.trim(),
    })
    if (response.data.code === 200) {
      toast.success(response.data.msg || 'OIDC 发现文档解析通过，连通正常')
    } else {
      toast.error(response.data.msg || '连接失败')
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.msg || e?.message || '连接测试失败，请核对 Issuer 地址')
  } finally {
    testingOidc.value = false
  }
}

async function handleTestOAuth2() {
  if (isDemo.value || checkIsDemoMode()) {
    toast.error('当前处于演示模式，已禁止测试第三方连接！')
    return
  }
  if (!form.value.oauth2_auth_url.trim() || !form.value.oauth2_token_url.trim() || !form.value.oauth2_userinfo_url.trim()) {
    toast.error('请完整填写授权、令牌及用户信息端点')
    return
  }

  testingOAuth2.value = true
  try {
    const response = await request.post('/auth/oauth2/test', {
      type: 'oauth2',
      oauth2_auth_url: form.value.oauth2_auth_url.trim(),
      oauth2_token_url: form.value.oauth2_token_url.trim(),
      oauth2_userinfo_url: form.value.oauth2_userinfo_url.trim(),
    })
    if (response.data.code === 200) {
      toast.success(response.data.msg || 'OAuth 2.0 端点连通正常')
    } else {
      toast.error(response.data.msg || '端点连接失败')
    }
  } catch (e: any) {
    toast.error(e?.response?.data?.msg || e?.message || 'OAuth 2.0 端点测试失败')
  } finally {
    testingOAuth2.value = false
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>单点登录设置</CardTitle>
      <CardDescription>配置 OIDC 与 OAuth 2.0 登录参数</CardDescription>
    </CardHeader>
    <CardContent>
      <div class="setting-container">
        <!-- 演示模式提示横幅 -->
        <div v-if="isDemo" class="mb-4 p-3 rounded-lg border border-amber-500/30 bg-amber-500/10 text-xs text-amber-700 dark:text-amber-400 flex items-center gap-2">
          <span>⚠️ 当前系统运行于演示模式，已全局禁用第三方单点登录 (SSO) 与注册功能，相关配置仅供展示与只读预览。</span>
        </div>

        <!-- 标签页切换 -->
        <Tabs v-model="activeTab" class="w-full">
          <TabsList class="grid w-full grid-cols-2 mb-4 sm:mb-6 h-9 sm:h-10">
            <TabsTrigger value="oidc" class="text-xs sm:text-sm">标准 OIDC</TabsTrigger>
            <TabsTrigger value="oauth2" class="text-xs sm:text-sm">通用 OAuth 2.0</TabsTrigger>
          </TabsList>

          <!-- ==================== 1. 标准 OIDC ==================== -->
          <TabsContent value="oidc" class="space-y-4">
            <!-- 启用 OIDC 开关卡片 -->
            <div class="flex items-center justify-between p-3 sm:p-3.5 border border-dashed border-border rounded-lg bg-muted/20 gap-3">
              <div class="space-y-0.5 flex-1 min-w-0 pr-2">
                <label class="text-xs sm:text-sm font-semibold text-foreground block">启用 OIDC 登录</label>
                <p class="text-[11px] sm:text-xs text-muted-foreground leading-relaxed break-words">
                  适用于 Keycloak、Casdoor、Pocket ID、Authentik 等标准身份提供商
                </p>
              </div>
              <Switch v-model="form.oidc_enabled" :disabled="loading" class="shrink-0" />
            </div>

            <!-- 授权重定向地址 -->
            <div class="space-y-1.5">
              <div class="flex items-center justify-between gap-2">
                <label class="text-xs sm:text-sm font-medium text-foreground truncate">
                  授权重定向地址 (Redirect URI)
                </label>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  class="h-7 px-2 text-xs shrink-0 hover:bg-muted"
                  @click="copyUrl(oidcRedirectUrl, 'oidc')"
                >
                  <Check v-if="copiedOidc" class="w-3.5 h-3.5 mr-1 text-emerald-500" />
                  <Copy v-else class="w-3.5 h-3.5 mr-1" />
                  {{ copiedOidc ? '已复制' : '复制' }}
                </Button>
              </div>
              <Input
                :model-value="oidcRedirectUrl"
                readonly
                class="font-mono bg-muted/40 cursor-pointer text-xs sm:text-sm h-9 sm:h-10 truncate"
                @click="copyUrl(oidcRedirectUrl, 'oidc')"
                title="点击复制"
              />
              <p class="text-[11px] sm:text-xs text-muted-foreground leading-relaxed">
                请在 OIDC IdP 后台的 Redirect URIs 白名单中填入此地址
              </p>
            </div>

            <!-- Issuer 地址 -->
            <div class="space-y-1.5">
              <div class="flex flex-wrap items-center justify-between gap-1">
                <label class="text-xs sm:text-sm font-medium text-foreground">
                  Issuer（身份提供商根地址）
                  <span class="text-destructive">*</span>
                </label>
                <a
                  v-if="discoveryUrl"
                  :href="discoveryUrl"
                  target="_blank"
                  class="text-xs text-brand hover:underline inline-flex items-center gap-1 shrink-0"
                >
                  <span>查看发现文档</span>
                  <ExternalLink class="w-3 h-3" />
                </a>
              </div>
              <Input
                v-model="form.oidc_issuer"
                @blur="cleanOidcIssuer"
                placeholder="例如 https://id.example.com 或 https://auth.corp.com/realms/master"
                class="font-mono text-xs sm:text-sm h-9 sm:h-10"
                autocomplete="off"
                :disabled="loading"
              />
              <p class="text-[11px] sm:text-xs text-muted-foreground leading-relaxed break-words">
                系统将自动请求 <code class="font-mono text-brand">/.well-known/openid-configuration</code> 探测元数据
              </p>
            </div>

            <!-- Client ID & Client Secret (使用文本框+密文遮罩完全杜绝浏览器密码填充) -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 sm:gap-4">
              <div class="space-y-1.5">
                <label class="text-xs sm:text-sm font-medium text-foreground block">
                  Client ID（客户端标识）
                  <span class="text-destructive">*</span>
                </label>
                <Input
                  v-model="form.oidc_client_id"
                  name="oidc_client_id"
                  placeholder="请输入 Client ID"
                  class="font-mono text-xs sm:text-sm h-9 sm:h-10"
                  autocomplete="off"
                  data-lpignore="true"
                  data-1p-ignore="true"
                  :disabled="loading"
                />
              </div>

              <div class="space-y-1.5">
                <label class="text-xs sm:text-sm font-medium text-foreground block">Client Secret（客户端密钥）</label>
                <div class="relative">
                  <Input
                    v-model="form.oidc_client_secret"
                    type="text"
                    name="oidc_client_secret"
                    placeholder="留空保持不变"
                    class="font-mono pr-9 text-xs sm:text-sm h-9 sm:h-10"
                    :class="{ 'secret-mask': !showSecretOidc }"
                    autocomplete="off"
                    data-lpignore="true"
                    data-1p-ignore="true"
                    :disabled="loading"
                  />
                  <button
                    type="button"
                    @click="showSecretOidc = !showSecretOidc"
                    class="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors p-1"
                  >
                    <EyeOff v-if="showSecretOidc" class="w-4 h-4" />
                    <Eye v-else class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Scopes (授权范围与快捷按钮，移动端自适应换行防溢出) -->
            <div class="space-y-1.5">
              <div class="flex flex-wrap items-center justify-between gap-1.5">
                <label class="text-xs sm:text-sm font-medium text-foreground">授权范围 (Scopes)</label>
                <div class="flex flex-wrap items-center gap-1.5">
                  <button
                    v-for="s in oidcCommonScopes"
                    :key="s"
                    type="button"
                    @click="addOidcScope(s)"
                    class="text-[11px] sm:text-xs px-2 py-0.5 rounded bg-muted hover:bg-muted/80 text-foreground border border-border/60 transition-colors shrink-0"
                  >
                    +{{ s }}
                  </button>
                </div>
              </div>
              <Input
                v-model="form.oidc_scopes"
                placeholder="默认包含 openid profile email，以空格分隔其他 scope"
                class="font-mono text-xs sm:text-sm h-9 sm:h-10"
                autocomplete="off"
                :disabled="loading"
              />
            </div>

            <!-- 首次登录自动建号 -->
            <div class="flex items-center justify-between p-3 sm:p-3.5 border border-dashed border-border rounded-lg bg-muted/20 gap-3">
              <div class="space-y-0.5 flex-1 min-w-0 pr-2">
                <label class="text-xs sm:text-sm font-semibold text-foreground block">首次登录自动创建账号</label>
                <p class="text-[11px] sm:text-xs text-muted-foreground leading-relaxed break-words">
                  未关联系统账号的 OIDC 身份首次成功验证后自动开通本地账号
                </p>
              </div>
              <Switch v-model="form.oidc_auto_create" :disabled="loading" class="shrink-0" />
            </div>
          </TabsContent>

          <!-- ==================== 2. 通用 OAuth 2.0 ==================== -->
          <TabsContent value="oauth2" class="space-y-4">
            <!-- 启用 OAuth 2.0 开关卡片 -->
            <div class="flex items-center justify-between p-3 sm:p-3.5 border border-dashed border-border rounded-lg bg-muted/20 gap-3">
              <div class="space-y-0.5 flex-1 min-w-0 pr-2">
                <label class="text-xs sm:text-sm font-semibold text-foreground block">启用通用 OAuth 2.0 登录</label>
                <p class="text-[11px] sm:text-xs text-muted-foreground leading-relaxed break-words">
                  适用于 GitHub、Gitee、GitLab 等通用 OAuth 2.0 第三方快捷登录
                </p>
              </div>
              <Switch v-model="form.oauth2_enabled" :disabled="loading" class="shrink-0" />
            </div>

            <!-- 常见平台一键预设 (移动端自适应换行) -->
            <div class="flex flex-col sm:flex-row sm:items-center justify-between p-3 sm:p-3.5 border border-border rounded-lg bg-muted/20 gap-2.5">
              <div class="space-y-0.5">
                <label class="text-xs sm:text-sm font-semibold text-foreground block">常见平台一键预设</label>
                <p class="text-[11px] sm:text-xs text-muted-foreground">快速填入常用服务商的端点和 Scope</p>
              </div>
              <div class="flex flex-wrap items-center gap-1.5 sm:gap-2 shrink-0">
                <Button
                  v-for="p in oauth2Presets"
                  :key="p.name"
                  type="button"
                  variant="outline"
                  size="sm"
                  class="h-7 sm:h-8 text-xs px-2.5"
                  @click="applyOAuth2Preset(p)"
                  :disabled="loading"
                >
                  {{ p.name }}
                </Button>
              </div>
            </div>

            <!-- OAuth 2.0 授权重定向地址 -->
            <div class="space-y-1.5">
              <div class="flex items-center justify-between gap-2">
                <label class="text-xs sm:text-sm font-medium text-foreground truncate">
                  授权重定向地址 (Redirect URI)
                </label>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  class="h-7 px-2 text-xs shrink-0 hover:bg-muted"
                  @click="copyUrl(oauth2RedirectUrl, 'oauth2')"
                >
                  <Check v-if="copiedOAuth2" class="w-3.5 h-3.5 mr-1 text-emerald-500" />
                  <Copy v-else class="w-3.5 h-3.5 mr-1" />
                  {{ copiedOAuth2 ? '已复制' : '复制' }}
                </Button>
              </div>
              <Input
                :model-value="oauth2RedirectUrl"
                readonly
                class="font-mono bg-muted/40 cursor-pointer text-xs sm:text-sm h-9 sm:h-10 truncate"
                @click="copyUrl(oauth2RedirectUrl, 'oauth2')"
                title="点击复制"
              />
              <p class="text-[11px] sm:text-xs text-muted-foreground leading-relaxed">
                请在 OAuth 2.0 应用的回调地址 (Callback URL) 中填写此地址
              </p>
            </div>

            <!-- 平台显示名称 & 用户标识 Key -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 sm:gap-4">
              <div class="space-y-1.5">
                <label class="text-xs sm:text-sm font-medium text-foreground block">平台显示名称</label>
                <Input
                  v-model="form.oauth2_name"
                  placeholder="例如 GitHub、Gitee"
                  class="text-xs sm:text-sm h-9 sm:h-10"
                  autocomplete="off"
                  :disabled="loading"
                />
              </div>

              <div class="space-y-1.5">
                <label class="text-xs sm:text-sm font-medium text-foreground block">用户唯一标识 Key</label>
                <Input
                  v-model="form.oauth2_user_key"
                  placeholder="留空自动探测（默认尝试 id、sub、login）"
                  class="font-mono text-xs sm:text-sm h-9 sm:h-10"
                  autocomplete="off"
                  :disabled="loading"
                />
              </div>
            </div>

            <!-- Client ID & Client Secret (使用文本框+密文遮罩完全杜绝浏览器密码填充) -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3 sm:gap-4">
              <div class="space-y-1.5">
                <label class="text-xs sm:text-sm font-medium text-foreground block">
                  Client ID（客户端标识）
                  <span class="text-destructive">*</span>
                </label>
                <Input
                  v-model="form.oauth2_client_id"
                  name="oauth2_client_id"
                  placeholder="请输入 Client ID"
                  class="font-mono text-xs sm:text-sm h-9 sm:h-10"
                  autocomplete="off"
                  data-lpignore="true"
                  data-1p-ignore="true"
                  :disabled="loading"
                />
              </div>

              <div class="space-y-1.5">
                <label class="text-xs sm:text-sm font-medium text-foreground block">Client Secret（客户端密钥）</label>
                <div class="relative">
                  <Input
                    v-model="form.oauth2_client_secret"
                    type="text"
                    name="oauth2_client_secret"
                    placeholder="留空保持不变"
                    class="font-mono pr-9 text-xs sm:text-sm h-9 sm:h-10"
                    :class="{ 'secret-mask': !showSecretOAuth2 }"
                    autocomplete="off"
                    data-lpignore="true"
                    data-1p-ignore="true"
                    :disabled="loading"
                  />
                  <button
                    type="button"
                    @click="showSecretOAuth2 = !showSecretOAuth2"
                    class="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors p-1"
                  >
                    <EyeOff v-if="showSecretOAuth2" class="w-4 h-4" />
                    <Eye v-else class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>

            <!-- 端点配置 -->
            <div class="space-y-1.5">
              <label class="text-xs sm:text-sm font-medium text-foreground block">
                授权端点 (Authorize URL)
                <span class="text-destructive">*</span>
              </label>
              <Input
                v-model="form.oauth2_auth_url"
                placeholder="例如 https://github.com/login/oauth/authorize"
                class="font-mono text-xs sm:text-sm h-9 sm:h-10"
                autocomplete="off"
                :disabled="loading"
              />
            </div>

            <div class="space-y-1.5">
              <label class="text-xs sm:text-sm font-medium text-foreground block">
                令牌端点 (Token URL)
                <span class="text-destructive">*</span>
              </label>
              <Input
                v-model="form.oauth2_token_url"
                placeholder="例如 https://github.com/login/oauth/access_token"
                class="font-mono text-xs sm:text-sm h-9 sm:h-10"
                autocomplete="off"
                :disabled="loading"
              />
            </div>

            <div class="space-y-1.5">
              <label class="text-xs sm:text-sm font-medium text-foreground block">
                用户信息端点 (Userinfo URL)
                <span class="text-destructive">*</span>
              </label>
              <Input
                v-model="form.oauth2_userinfo_url"
                placeholder="例如 https://api.github.com/user"
                class="font-mono text-xs sm:text-sm h-9 sm:h-10"
                autocomplete="off"
                :disabled="loading"
              />
            </div>

            <!-- Scopes (授权范围与快捷按钮) -->
            <div class="space-y-1.5">
              <div class="flex flex-wrap items-center justify-between gap-1.5">
                <label class="text-xs sm:text-sm font-medium text-foreground">授权范围 (Scopes)</label>
                <div class="flex flex-wrap items-center gap-1.5">
                  <button
                    v-for="s in oauth2CommonScopes"
                    :key="s"
                    type="button"
                    @click="addOAuth2Scope(s)"
                    class="text-[11px] sm:text-xs px-2 py-0.5 rounded bg-muted hover:bg-muted/80 text-foreground border border-border/60 transition-colors shrink-0"
                  >
                    +{{ s }}
                  </button>
                </div>
              </div>
              <Input
                v-model="form.oauth2_scopes"
                placeholder="以空格分隔其他 scope"
                class="font-mono text-xs sm:text-sm h-9 sm:h-10"
                autocomplete="off"
                :disabled="loading"
              />
            </div>

            <!-- 首次登录自动建号 -->
            <div class="flex items-center justify-between p-3 sm:p-3.5 border border-dashed border-border rounded-lg bg-muted/20 gap-3">
              <div class="space-y-0.5 flex-1 min-w-0 pr-2">
                <label class="text-xs sm:text-sm font-semibold text-foreground block">首次登录自动创建账号</label>
                <p class="text-[11px] sm:text-xs text-muted-foreground leading-relaxed break-words">
                  未关联系统账号的 OAuth 2.0 身份首次成功验证后自动开通本地账号
                </p>
              </div>
              <Switch v-model="form.oauth2_auto_create" :disabled="loading" class="shrink-0" />
            </div>
          </TabsContent>
        </Tabs>

        <!-- 底部操作区域 (严格统一于 SiteSettings 与 CleanSettings，小屏自适应) -->
        <div class="flex items-center justify-between gap-2 mt-6 pt-3 border-t border-border/40">
          <div class="flex items-center space-x-1.5 shrink-0">
            <span class="text-xs sm:text-sm text-muted-foreground">说明</span>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <HelpCircleIcon class="w-4 h-4 text-muted-foreground/70 hover:text-foreground transition-colors" />
                </TooltipTrigger>
                <TooltipContent class="max-w-sm">
                  <div class="text-xs sm:text-sm space-y-1">
                    <p>1. 支持同时开启 OIDC 与通用 OAuth 2.0，登录页将自适应展示对应的快捷入口</p>
                    <p>2. 标准 OIDC 填入根 Issuer 后系统将自动通过发现文档协商端点与 JWKS 密钥</p>
                    <p>3. 授权回调地址需要在身份提供商后台的回调白名单中严格一致配置</p>
                    <p>4. 首次登录自动创建账号若关闭，未绑定的第三方用户需先在系统内有对应账号</p>
                  </div>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>

          <div class="flex items-center space-x-2 shrink-0">
            <Button
              variant="outline"
              size="sm"
              class="h-8 sm:h-9 text-xs sm:text-sm px-2.5 sm:px-3"
              @click="activeTab === 'oidc' ? handleTestOidc() : handleTestOAuth2()"
              :disabled="testingOidc || testingOAuth2 || saving || loading"
            >
              <Loader2 v-if="testingOidc || testingOAuth2" class="w-3.5 h-3.5 mr-1 animate-spin" />
              <span>{{ activeTab === 'oidc' ? (testingOidc ? '测试中...' : '测试连通性') : (testingOAuth2 ? '测试中...' : '测试连通性') }}</span>
            </Button>
            <Button size="sm" class="h-8 sm:h-9 text-xs sm:text-sm px-3.5 sm:px-5" @click="save" :disabled="saving || loading">
              <Loader2 v-if="saving" class="w-3.5 h-3.5 mr-1 animate-spin" />
              <span>确定</span>
            </Button>
          </div>
        </div>
      </div>
    </CardContent>
  </Card>
</template>

<style scoped>
:deep(.secret-mask),
.secret-mask {
  -webkit-text-security: disc !important;
}

/* 避免浏览器 autofill 给输入框注入破坏主题样式的奇怪背景色 */
:deep(input:-webkit-autofill),
:deep(input:-webkit-autofill:hover),
:deep(input:-webkit-autofill:focus),
:deep(input:-webkit-autofill:active) {
  transition: background-color 50000s ease-in-out 0s, color 50000s ease-in-out 0s;
  -webkit-text-fill-color: inherit !important;
}
</style>
