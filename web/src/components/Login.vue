<script setup lang="ts">
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { toast } from "vue-sonner"
import { ref, onMounted } from "vue"
import { Card, CardContent } from '@/components/ui/card'
import { Fingerprint, Zap, AlertCircle, UserPlus, LogIn } from 'lucide-vue-next'
// @ts-ignore
import config from '../../config.js'

// @ts-ignore
import { CONSTANT } from '@/constant';
// @ts-ignore
import { request } from '@/api/api';
import { usePageState } from '@/store/page_sate';
import { useRouter } from 'vue-router';
// @ts-ignore
import { LocalStieConfigUtils } from '@/util/localSiteConfig';
import { checkIsDemoMode, setRuntimeDemoMode } from '@/util/demo';

const siteConfigData = JSON.parse(CONSTANT.DEFALUT_SITE_CONFIG);
let logo = ref('data:image/svg+xml;base64,' + btoa(siteConfigData.logo));
let slogan = ref(siteConfigData.slogan);
if (siteConfigData.title) {
  document.title = siteConfigData.title;
}
let pageState = usePageState();
let router = useRouter();

// 模式切换: 'login' | 'register' | 'sso-bind'
let activeMode = ref<'login' | 'register' | 'sso-bind'>('login');
let allowRegistration = ref(false);

// SSO 绑定引导状态
let ssoTicket = ref('');
let ssoProvider = ref('');
let ssoUsernameHint = ref('');
let ssoEmailHint = ref('');
let ssoSubMode = ref<'link' | 'create'>('link'); // link: 关联已有账号, create: 创建新账号

// SSO 关联已有账号表单
let linkAccount = ref('');
let linkPassword = ref('');

// SSO 创建新账号表单
let bindNewAccount = ref('');
let bindNewEmail = ref('');
let bindNewPassword = ref('');
let bindNewConfirmPassword = ref('');

// 演示模式智能判定
let isDemo = ref(checkIsDemoMode());
let account = ref(isDemo.value ? "admin" : "");
let password = ref(isDemo.value ? "123456" : "");
let loading = ref(false);
let ssoProviders = ref<any[]>([]);
let loginErrorMessage = ref('');

// 注册表单
let regAccount = ref('');
let regEmail = ref('');
let regPassword = ref('');
let regConfirmPassword = ref('');
let regAccountError = ref('');
let bindAccountError = ref('');

// 实时校验注册用户名唯一性
const checkRegAccount = async () => {
  const name = regAccount.value.trim();
  if (!name) {
    regAccountError.value = '';
    return;
  }
  if (name.length < 2 || name.length > 30) {
    regAccountError.value = '用户名长度需在 2 到 30 个字符之间';
    return;
  }
  try {
    const res = await request.get('/api/v1/auth/check-username', { params: { username: name } });
    if (res.data && res.data.code === 200 && res.data.data) {
      if (!res.data.data.available) {
        regAccountError.value = res.data.data.reason || '该用户名已被使用';
      } else {
        regAccountError.value = '';
      }
    }
  } catch (e) {}
};

// 实时校验 SSO 创建新账号用户名唯一性
const checkBindAccount = async () => {
  const name = bindNewAccount.value.trim();
  if (!name) {
    bindAccountError.value = '';
    return;
  }
  if (name.length < 2 || name.length > 30) {
    bindAccountError.value = '用户名长度需在 2 到 30 个字符之间';
    return;
  }
  try {
    const res = await request.get('/api/v1/auth/check-username', { params: { username: name } });
    if (res.data && res.data.code === 200 && res.data.data) {
      if (!res.data.data.available) {
        bindAccountError.value = res.data.data.reason || '该用户名已被使用';
      } else {
        bindAccountError.value = '';
      }
    }
  } catch (e) {}
};

function safeDecode(str: string): string {
  try {
    return decodeURIComponent(str);
  } catch (e) {
    return str;
  }
}

const handleSSOLogin = (provider: any) => {
  const prefix = config.pathPrefix || '';
  const redirect = encodeURIComponent(window.location.hash || '#/');
  const targetUrl = provider.login_url;
  const sep = targetUrl.includes('?') ? '&' : '?';
  window.location.href = `${prefix}${targetUrl}${sep}redirect=${redirect}`;
};

// 取消 SSO 引导并返回登录
const cancelSSOBind = () => {
  ssoTicket.value = '';
  activeMode.value = 'login';
  // 清理 URL query
  if (typeof window !== 'undefined' && window.history) {
    const cleanUrl = window.location.pathname + (window.location.hash.split('?')[0] || '#/login');
    window.history.replaceState({}, '', cleanUrl);
  }
};

// 演示模式填充
const demoSiteSet = (force = false, showToast = false) => {
  if (force || isDemo.value) {
    account.value = 'admin';
    password.value = '123456';
    if (showToast) {
      toast.success('已填入演示账号密码');
    }
  }
};

// 切换至注册模式
const switchToRegister = () => {
  if (isDemo.value) {
    toast.error('演示模式下已禁用自主注册功能');
    return;
  }
  if (!allowRegistration.value) {
    toast.error('当前系统未开放自主注册，请联系管理员分配账号');
    return;
  }
  activeMode.value = 'register';
};

// 切换至登录模式
const switchToLogin = () => {
  activeMode.value = 'login';
};

// 更新favicon
const updateFavicon = (logoSvg: string) => {
  if (logoSvg) {
    let link = document.querySelector("link[rel*='icon']") as HTMLLinkElement
    if (!link) {
      link = document.createElement('link')
      link.rel = 'icon'
      document.head.appendChild(link)
    }
    const svgBlob = new Blob([logoSvg], { type: 'image/svg+xml' })
    const svgUrl = URL.createObjectURL(svgBlob)
    link.href = svgUrl
    link.type = 'image/svg+xml'
  }
}

onMounted(async () => {
  if (isDemo.value) {
    demoSiteSet(true);
    allowRegistration.value = false;
    ssoProviders.value = [];
  }

  // 检查 SSO 回调参数
  const queryParams = new URLSearchParams(window.location.search);
  const hash = window.location.hash || '';
  let token = queryParams.get('token');
  let rawError = queryParams.get('error') || queryParams.get('message');
  let ticketVal = queryParams.get('sso_ticket');
  let providerVal = queryParams.get('sso_provider');
  let usernameVal = queryParams.get('sso_username');
  let emailVal = queryParams.get('sso_email');

  if (hash.includes('?')) {
    const hashParams = new URLSearchParams(hash.substring(hash.indexOf('?') + 1));
    if (!token) token = hashParams.get('token');
    if (!rawError) rawError = hashParams.get('error') || hashParams.get('message');
    if (!ticketVal) ticketVal = hashParams.get('sso_ticket');
    if (!providerVal) providerVal = hashParams.get('sso_provider');
    if (!usernameVal) usernameVal = hashParams.get('sso_username');
    if (!emailVal) emailVal = hashParams.get('sso_email');
  }

  if (ticketVal) {
    if (isDemo.value) {
      toast.error('演示模式下已禁用第三方单点登录功能');
      cancelSSOBind();
    } else {
      ssoTicket.value = ticketVal;
      ssoProvider.value = safeDecode(providerVal || 'SSO');
      ssoUsernameHint.value = safeDecode(usernameVal || '');
      ssoEmailHint.value = safeDecode(emailVal || '');
      bindNewAccount.value = safeDecode(usernameVal || '');
      bindNewEmail.value = safeDecode(emailVal || '');
      activeMode.value = 'sso-bind';
      toast.info('第三方身份验证通过，请关联内部账号完成绑定');
    }
  } else if (rawError) {
    loginErrorMessage.value = safeDecode(rawError);
    toast.error('登录提示：' + loginErrorMessage.value);
  } else if (token) {
    pageState.setToken(token);
    toast.success('登录成功');
    setTimeout(() => {
      router.push('/');
    }, 400);
    return;
  }

  // 加载单点登录可用身份源与服务端状态
  try {
    const ssoRes = await request.get('/auth/sso/status');
    if (ssoRes.data && ssoRes.data.code === 200 && ssoRes.data.data) {
      if (ssoRes.data.data.is_demo) {
        setRuntimeDemoMode(true);
        isDemo.value = true;
        demoSiteSet(true);
        allowRegistration.value = false;
        ssoProviders.value = [];
      }
      if (!isDemo.value && ssoRes.data.data.enabled && ssoRes.data.data.providers) {
        ssoProviders.value = ssoRes.data.data.providers;
      } else {
        ssoProviders.value = [];
      }
    }
  } catch (e) {
    console.warn('获取 SSO 配置失败:', e);
  }

  // 获取公开站点配置 (含 allow_registration 开关)
  try {
    const siteRes = await request.get('/public/site/config');
    if (siteRes.data && siteRes.data.code === 200 && siteRes.data.data) {
      const pData = siteRes.data.data;
      if (pData.is_demo) {
        setRuntimeDemoMode(true);
        isDemo.value = true;
        demoSiteSet(true);
      }
      if (isDemo.value) {
        allowRegistration.value = false;
        ssoProviders.value = [];
      } else {
        allowRegistration.value = pData.allow_registration === 'true' || pData.allow_registration === true;
      }
      if (pData.title) document.title = pData.title;
      if (pData.slogan) slogan.value = pData.slogan;
      if (pData.logo) {
        logo.value = 'data:image/svg+xml;base64,' + btoa(pData.logo);
        updateFavicon(pData.logo);
      }
      // 成功获取服务端最新配置后，同步更新本地存储，避免本地残留旧配置
      try {
        LocalStieConfigUtils.updateLocalConfig(pData);
      } catch (err) {}
    }
  } catch (e) {
    console.warn('获取公开站点配置失败，尝试本地配置兜底:', e);
    // 仅在网络请求失败时才启用本地配置兜底
    try {
      const localConfig = LocalStieConfigUtils.getLocalConfig();
      if (!isDemo.value && localConfig && localConfig.allow_registration !== undefined) {
        allowRegistration.value = localConfig.allow_registration === 'true' || localConfig.allow_registration === true;
      }
    } catch (err) {}
  }
});

// 登录提交
let clickLogin = async () => {
  if (loading.value) return;
  if (!account.value.trim() || !password.value.trim()) {
    toast.error('请输入账号和密码');
    return;
  }
  
  loading.value = true;
  try {
    const rspe = await request.post('/api/v1/auth/login', { username: account.value, passwd: password.value });
    const rsp = rspe.data;
    if (rsp.code != 200) {
      toast.error(rsp.msg || '登录失败');
    } else {
      pageState.setToken(rsp.data.token);
      toast.success('登录成功');
      setTimeout(() => {
        router.push('/');
      }, 400);
    }
  } catch (error: any) {
    toast.error('登录失败，请检查网络连接');
  } finally {
    loading.value = false;
  }
};

// 注册提交
let clickRegister = async () => {
  if (loading.value) return;
  if (isDemo.value) {
    toast.error('演示模式下已禁用自主注册功能');
    return;
  }
  if (!allowRegistration.value) {
    toast.error('当前系统未开放自主注册');
    return;
  }
  const u = regAccount.value.trim();
  const p = regPassword.value.trim();
  const cp = regConfirmPassword.value.trim();
  const em = regEmail.value.trim();

  if (!u) {
    toast.error('请输入用户名');
    return;
  }
  if (u.length < 2 || u.length > 30) {
    toast.error('用户名长度须在 2 到 30 个字符之间');
    return;
  }
  if (regAccountError.value) {
    toast.error(regAccountError.value);
    return;
  }
  if (!p || p.length < 4) {
    toast.error('密码长度至少为 4 位');
    return;
  }
  if (p !== cp) {
    toast.error('两次输入的密码不一致');
    return;
  }

  loading.value = true;
  try {
    const rsp = await request.post('/api/v1/auth/register', {
      username: u,
      password: p,
      email: em
    });
    if (rsp.data && rsp.data.code === 200) {
      const data = rsp.data.data;
      toast.success('注册成功！欢迎加入 Message Nest');
      if (data && data.token) {
        pageState.setToken(data.token);
        setTimeout(() => {
          router.push('/');
        }, 500);
      } else {
        activeMode.value = 'login';
        account.value = u;
        password.value = p;
      }
    } else {
      toast.error(rsp.data?.msg || '注册失败，请稍后重试');
    }
  } catch (e: any) {
    toast.error('注册请求失败: ' + (e?.message || '网络异常'));
  } finally {
    loading.value = false;
  }
};

// SSO 关联已有内部账号并登录
let clickSSOLinkAccount = async () => {
  if (loading.value) return;
  const u = linkAccount.value.trim();
  const p = linkPassword.value.trim();
  if (!u || !p) {
    toast.error('请输入内部账号与密码');
    return;
  }
  if (!ssoTicket.value) {
    toast.error('单点登录凭据已失效，请重新发起登录');
    return;
  }

  loading.value = true;
  try {
    const rsp = await request.post('/api/v1/auth/sso/bind-login', {
      ticket: ssoTicket.value,
      username: u,
      password: p
    });
    if (rsp.data && rsp.data.code === 200) {
      toast.success('已成功关联内部账号并登录！');
      if (rsp.data.data && rsp.data.data.token) {
        pageState.setToken(rsp.data.data.token);
        setTimeout(() => {
          router.push('/');
        }, 400);
      }
    } else {
      toast.error(rsp.data?.msg || '关联失败，请检查账号密码');
    }
  } catch (e: any) {
    toast.error('关联请求异常: ' + (e?.message || '网络错误'));
  } finally {
    loading.value = false;
  }
};

// SSO 创建内部账号并关联登录
let clickSSOCreateAccount = async () => {
  if (loading.value) return;
  if (!allowRegistration.value) {
    toast.error('当前系统未开放自主创建账号，请选择【关联已有账号】');
    return;
  }
  const u = bindNewAccount.value.trim();
  const p = bindNewPassword.value.trim();
  const cp = bindNewConfirmPassword.value.trim();
  const em = bindNewEmail.value.trim();

  if (!u) {
    toast.error('请输入内部账号名称');
    return;
  }
  if (u.length < 2 || u.length > 30) {
    toast.error('用户名长度须在 2 到 30 个字符之间');
    return;
  }
  if (bindAccountError.value) {
    toast.error(bindAccountError.value);
    return;
  }
  if (!p || p.length < 4) {
    toast.error('密码长度至少为 4 位');
    return;
  }
  if (p !== cp) {
    toast.error('两次输入的密码不一致');
    return;
  }
  if (!ssoTicket.value) {
    toast.error('单点登录凭据已失效，请重新发起登录');
    return;
  }

  loading.value = true;
  try {
    const rsp = await request.post('/api/v1/auth/sso/register-bind', {
      ticket: ssoTicket.value,
      username: u,
      password: p,
      email: em
    });
    if (rsp.data && rsp.data.code === 200) {
      toast.success('内部账号创建并关联成功！');
      if (rsp.data.data && rsp.data.data.token) {
        pageState.setToken(rsp.data.data.token);
        setTimeout(() => {
          router.push('/');
        }, 400);
      }
    } else {
      toast.error(rsp.data?.msg || '创建失败，请稍后重试');
    }
  } catch (e: any) {
    toast.error('创建请求异常: ' + (e?.message || '网络错误'));
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div class="flex min-h-svh flex-col items-center justify-center bg-muted p-6 md:p-10">
    <div class="w-full max-w-sm md:max-w-3xl">
      <div class="flex flex-col gap-6">
        <Card class="overflow-hidden shadow-lg border-border">
          <CardContent class="grid p-0 md:grid-cols-2">
            <form class="p-6 md:p-8">
              <div class="flex flex-col gap-5">
                <!-- 小屏标题与LOGO -->
                <div class="flex flex-col items-center text-center md:hidden">
                  <div class="w-16 h-16 mb-3 flex justify-center items-center bg-card dark:bg-muted border border-border rounded-full shadow-sm">
                    <img :src="logo" alt="logo" class="w-10 h-10 object-contain" />
                  </div>
                  <h1 class="text-2xl font-bold text-foreground">
                    {{ activeMode === 'login' ? '欢迎回来' : activeMode === 'register' ? '创建新账号' : '关联内部系统账号' }}
                  </h1>
                  <p class="text-xs text-muted-foreground mt-1">
                    {{ activeMode === 'sso-bind' ? '第三方单点登录授权成功，请关联内部账号完成接入' : slogan }}
                  </p>
                </div>
                
                <!-- 大屏标题 -->
                <div class="hidden md:flex flex-col items-center text-center">
                  <h1 class="text-2xl font-bold tracking-tight">
                    {{ activeMode === 'login' ? '欢迎回来' : activeMode === 'register' ? '创建新账号' : '关联内部系统账号' }}
                  </h1>
                  <p class="text-xs text-muted-foreground mt-1">
                    {{ activeMode === 'sso-bind' ? '第三方单点登录授权成功，请关联内部账号完成接入' : slogan }}
                  </p>
                </div>

                <!-- 模式切换 Tabs（如果系统开启注册、非 SSO 绑定引导模式且非演示环境） -->
                <div v-if="!isDemo && allowRegistration && activeMode !== 'sso-bind'" class="flex p-1 bg-muted rounded-lg border border-border">
                  <button
                    type="button"
                    class="flex-1 py-1.5 text-xs font-medium rounded-md transition-all flex items-center justify-center gap-1.5"
                    :class="activeMode === 'login' ? 'bg-card text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'"
                    @click="switchToLogin"
                  >
                    <LogIn class="w-3.5 h-3.5" />
                    <span>账号登录</span>
                  </button>
                  <button
                    type="button"
                    class="flex-1 py-1.5 text-xs font-medium rounded-md transition-all flex items-center justify-center gap-1.5"
                    :class="activeMode === 'register' ? 'bg-card text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'"
                    @click="switchToRegister"
                  >
                    <UserPlus class="w-3.5 h-3.5" />
                    <span>注册账号</span>
                  </button>
                </div>

                <!-- 登录失败提示 -->
                <div v-if="loginErrorMessage" class="p-3 rounded-xl border border-destructive/30 bg-destructive/10 text-destructive text-xs leading-relaxed flex items-start gap-2.5">
                  <AlertCircle class="w-4 h-4 shrink-0 mt-0.5" />
                  <div class="flex-1">
                    <div class="font-semibold text-xs mb-0.5">登录未成功</div>
                    <p class="break-words">{{ loginErrorMessage }}</p>
                  </div>
                  <button type="button" @click="loginErrorMessage = ''" class="text-destructive/70 hover:text-destructive p-0.5" title="关闭">
                    ✕
                  </button>
                </div>

                <!-- 演示环境提示 (仅普通登录模式下展示) -->
                <div v-if="isDemo && activeMode === 'login'" class="p-2.5 px-3 rounded-xl border border-primary/20 bg-primary/5 text-xs flex items-center justify-between gap-3">
                  <div class="flex items-center gap-2.5 min-w-0">
                    <span class="relative flex h-2 w-2 shrink-0">
                      <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"></span>
                      <span class="relative inline-flex rounded-full h-2 w-2 bg-primary"></span>
                    </span>
                    <div class="min-w-0 space-y-0.5">
                      <div class="text-xs font-semibold text-primary leading-none">演示环境</div>
                      <div class="text-[11px] text-muted-foreground font-mono tracking-tight leading-none">账号 admin · 密码 123456</div>
                    </div>
                  </div>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    class="h-7 px-2.5 text-xs font-medium shrink-0 whitespace-nowrap border-primary/30 text-primary hover:bg-primary/10 hover:text-primary transition-all cursor-pointer select-none"
                    @click.prevent="demoSiteSet(true, true)"
                  >
                    自动填入
                  </Button>
                </div>

                <!-- SSO 引导绑定模式表单 -->
                <template v-if="activeMode === 'sso-bind'">
                  <!-- 提示横幅 -->
                  <div class="p-3 rounded-xl border border-primary/20 bg-primary/5 text-xs flex items-start gap-2.5">
                    <Fingerprint class="w-4 h-4 text-primary shrink-0 mt-0.5" />
                    <div class="flex-1 space-y-1">
                      <div class="font-semibold text-primary">已通过第三方 {{ ssoProvider }} 授权验证</div>
                      <p class="text-muted-foreground text-[11px] leading-relaxed">
                        系统要求必须拥有内部账号方可单点登录。请选择关联已有账号或创建新的内部账号。
                      </p>
                    </div>
                  </div>

                  <!-- 关联子选项卡 -->
                  <div class="flex p-1 bg-muted rounded-lg border border-border">
                    <button
                      type="button"
                      class="flex-1 py-1.5 text-xs font-medium rounded-md transition-all flex items-center justify-center gap-1.5"
                      :class="ssoSubMode === 'link' ? 'bg-card text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'"
                      @click="ssoSubMode = 'link'"
                    >
                      <LogIn class="w-3.5 h-3.5" />
                      <span>关联已有账号</span>
                    </button>
                    <button
                      type="button"
                      class="flex-1 py-1.5 text-xs font-medium rounded-md transition-all flex items-center justify-center gap-1.5"
                      :class="ssoSubMode === 'create' ? 'bg-card text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'"
                      @click="ssoSubMode = 'create'"
                    >
                      <UserPlus class="w-3.5 h-3.5" />
                      <span>创建内部账号</span>
                    </button>
                  </div>

                  <!-- 子表单 1: 关联已有账号 -->
                  <template v-if="ssoSubMode === 'link'">
                    <div class="grid gap-2">
                      <Label for="link-account" class="text-xs font-medium">已有系统用户名</Label>
                      <Input id="link-account" type="text" placeholder="请输入您的内部系统账号" v-model="linkAccount" required />
                    </div>
                    <div class="grid gap-2">
                      <Label for="link-password" class="text-xs font-medium">密码</Label>
                      <Input id="link-password" type="password" placeholder="请输入密码验证身份" v-model="linkPassword" required />
                    </div>
                    <Button
                      type="button"
                      class="w-full active:scale-95 transition duration-150"
                      @click.prevent="clickSSOLinkAccount"
                      :disabled="loading"
                    >
                      <svg v-if="loading" class="mr-2 h-4 w-4 animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      {{ loading ? '验证中...' : '验证密码并绑定登录' }}
                    </Button>
                  </template>

                  <!-- 子表单 2: 创建新内部账号并关联 -->
                  <template v-else>
                    <div v-if="!allowRegistration" class="p-3 rounded-lg border border-amber-500/30 bg-amber-500/10 text-xs text-amber-700 dark:text-amber-400">
                      当前系统未开放自主创建账号，请切换至【关联已有账号】或联系管理员分配账号。
                    </div>
                    <template v-else>
                      <div class="grid gap-2">
                        <Label for="bind-new-account" class="text-xs font-medium">内部账号用户名</Label>
                        <Input
                          id="bind-new-account"
                          type="text"
                          placeholder="2-30位字母/数字"
                          v-model="bindNewAccount"
                          :class="{ 'border-destructive focus-visible:ring-destructive': !!bindAccountError }"
                          @blur="checkBindAccount"
                          @input="bindAccountError = ''"
                          required
                        />
                        <p v-if="bindAccountError" class="text-xs text-destructive -mt-1">{{ bindAccountError }}</p>
                      </div>
                      <div class="grid gap-2">
                        <Label for="bind-new-email" class="text-xs font-medium">邮箱地址 (选填)</Label>
                        <Input id="bind-new-email" type="email" placeholder="第三方提供或留空" v-model="bindNewEmail" />
                      </div>
                      <div class="grid gap-2">
                        <Label for="bind-new-pwd" class="text-xs font-medium">设置登录密码</Label>
                        <Input id="bind-new-pwd" type="password" placeholder="至少 4 位字符" v-model="bindNewPassword" required />
                      </div>
                      <div class="grid gap-2">
                        <Label for="bind-new-cpwd" class="text-xs font-medium">确认密码</Label>
                        <Input id="bind-new-cpwd" type="password" placeholder="再次输入密码" v-model="bindNewConfirmPassword" required />
                      </div>
                      <Button
                        type="button"
                        class="w-full active:scale-95 transition duration-150"
                        @click.prevent="clickSSOCreateAccount"
                        :disabled="loading"
                      >
                        <svg v-if="loading" class="mr-2 h-4 w-4 animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                          <path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        {{ loading ? '创建中...' : '创建账号并完成绑定' }}
                      </Button>
                    </template>
                  </template>

                  <div class="text-center text-xs text-muted-foreground pt-1">
                    <a href="#" class="text-muted-foreground hover:text-foreground underline underline-offset-4" @click.prevent="cancelSSOBind">
                      放弃本次授权并返回常规登录
                    </a>
                  </div>
                </template>

                <!-- 登录模式表单 -->
                <template v-else-if="activeMode === 'login'">
                  <div class="grid gap-2">
                    <Label for="account" class="text-xs font-medium">账号</Label>
                    <Input id="account" type="text" placeholder="请输入登录账号" v-model="account" required />
                  </div>
                  <div class="grid gap-2">
                    <Label for="password" class="text-xs font-medium">密码</Label>
                    <Input id="password" type="password" placeholder="请输入密码" v-model="password" required />
                  </div>
                  <Button
                    type="submit"
                    class="w-full active:scale-95 transition duration-150"
                    @click.prevent="clickLogin"
                    :disabled="loading"
                  >
                    <svg v-if="loading" class="mr-2 h-4 w-4 animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    {{ loading ? '登录中...' : '立即登录' }}
                  </Button>

                  <!-- 单点登录快捷入口 (非演示环境且配置了可用身份源) -->
                  <div v-if="!isDemo && ssoProviders && ssoProviders.length > 0" class="space-y-3 pt-1">
                    <div class="relative text-center text-xs after:absolute after:inset-0 after:top-1/2 after:z-0 after:flex after:items-center after:border-t after:border-border">
                      <span class="relative z-10 bg-card px-2 text-muted-foreground select-none">
                        或使用单点登录
                      </span>
                    </div>
                    <div class="flex flex-col gap-2">
                      <Button
                        v-for="p in ssoProviders"
                        :key="p.id"
                        type="button"
                        variant="outline"
                        class="w-full h-9 text-xs font-medium flex items-center justify-center gap-2 hover:bg-muted transition-all"
                        @click="handleSSOLogin(p)"
                        :disabled="loading"
                      >
                        <Zap v-if="p.type === 'oauth2'" class="w-4 h-4 text-amber-500 shrink-0" />
                        <Fingerprint v-else class="w-4 h-4 text-primary shrink-0" />
                        <span>使用 {{ p.name }} 登录</span>
                      </Button>
                    </div>
                  </div>

                  <div v-if="!isDemo && allowRegistration" class="text-center text-xs text-muted-foreground pt-1">
                    <span>还没有账号？</span>
                    <a
                      href="#"
                      class="text-primary font-medium hover:underline ml-1"
                      @click.prevent="switchToRegister"
                    >
                      立即注册
                    </a>
                  </div>
                </template>

                <!-- 注册模式表单 -->
                <template v-else>
                  <div class="grid gap-2">
                    <Label for="reg-account" class="text-xs font-medium">账号名称</Label>
                    <Input
                      id="reg-account"
                      type="text"
                      placeholder="2-30位字母/数字"
                      v-model="regAccount"
                      :class="{ 'border-destructive focus-visible:ring-destructive': !!regAccountError }"
                      @blur="checkRegAccount"
                      @input="regAccountError = ''"
                      required
                    />
                    <p v-if="regAccountError" class="text-xs text-destructive -mt-1">{{ regAccountError }}</p>
                  </div>
                  <div class="grid gap-2">
                    <Label for="reg-email" class="text-xs font-medium">邮箱地址 (选填)</Label>
                    <Input id="reg-email" type="email" placeholder="用于接收通知与密码找回" v-model="regEmail" />
                  </div>
                  <div class="grid gap-2">
                    <Label for="reg-pwd" class="text-xs font-medium">设置密码</Label>
                    <Input id="reg-pwd" type="password" placeholder="至少 4 位字符" v-model="regPassword" required />
                  </div>
                  <div class="grid gap-2">
                    <Label for="reg-confirm-pwd" class="text-xs font-medium">确认密码</Label>
                    <Input id="reg-confirm-pwd" type="password" placeholder="再次输入密码" v-model="regConfirmPassword" required />
                  </div>

                  <Button
                    type="button"
                    class="w-full active:scale-95 transition duration-150"
                    @click.prevent="clickRegister"
                    :disabled="loading"
                  >
                    <svg v-if="loading" class="mr-2 h-4 w-4 animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    {{ loading ? '注册中...' : '确认注册并进入系统' }}
                  </Button>

                  <div class="text-center text-xs text-muted-foreground pt-1">
                    <span>已有账号？</span>
                    <a
                      href="#"
                      class="text-primary font-medium hover:underline ml-1"
                      @click.prevent="switchToLogin"
                    >
                      返回登录
                    </a>
                  </div>
                </template>

              </div>
            </form>
            <div class="relative hidden md:flex items-center justify-center bg-card border-l border-border/50 p-8">
              <div class="w-full h-full flex flex-col items-center justify-center">
                <img :src="logo" alt="logo" class="max-w-[160px] max-h-[160px] object-contain drop-shadow-sm" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>
