<script setup lang="ts">
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { toast } from "vue-sonner"
import { ref, onMounted } from "vue"
import { Card, CardContent } from '@/components/ui/card'
import { Fingerprint, Zap, AlertCircle } from 'lucide-vue-next'
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
// 设置默认网站标题
if (siteConfigData.title) {
  document.title = siteConfigData.title;
}
let pageState = usePageState();
let router = useRouter();

// 演示模式智能判定（环境变量 / demo 域名 / URL 参数）
let isDemo = ref(checkIsDemoMode());
let account = ref(isDemo.value ? "admin" : "");
let password = ref(isDemo.value ? "123456" : "");
let loading = ref(false);
let ssoProviders = ref<any[]>([]);
let loginErrorMessage = ref('');

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

// 站点演示模式填充
const demoSiteSet = (force = false, showToast = false) => {
  if (force || isDemo.value) {
    account.value = 'admin';
    password.value = '123456';
    if (showToast) {
      toast.success('已填入演示账号密码');
    }
  }
};

let handleSubmit = async function () {
  toast('注册功能还在开发的路上...');
}

// 更新favicon
const updateFavicon = (logoSvg: string) => {
  if (logoSvg) {
    // 查找现有的favicon链接
    let link = document.querySelector("link[rel*='icon']") as HTMLLinkElement
    if (!link) {
      // 如果不存在，创建新的favicon链接
      link = document.createElement('link')
      link.rel = 'icon'
      document.head.appendChild(link)
    }
    // 将SVG转换为data URL
    const svgBlob = new Blob([logoSvg], { type: 'image/svg+xml' })
    const svgUrl = URL.createObjectURL(svgBlob)
    link.href = svgUrl
    link.type = 'image/svg+xml'
  }
}

// 登录页面加载时获取最新配置与 SSO 状态
onMounted(async () => {
  if (isDemo.value) {
    demoSiteSet(true);
  }

  // 检查 SSO 回调参数 (token 或 error)
  const queryParams = new URLSearchParams(window.location.search);
  const hash = window.location.hash || '';
  let token = queryParams.get('token');
  let rawError = queryParams.get('error') || queryParams.get('message');

  if (!token && hash.includes('?')) {
    const hashParams = new URLSearchParams(hash.substring(hash.indexOf('?') + 1));
    token = hashParams.get('token');
    if (!rawError) {
      rawError = hashParams.get('error') || hashParams.get('message');
    }
  }

  if (rawError) {
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
      if (ssoRes.data.data.enabled && ssoRes.data.data.providers) {
        ssoProviders.value = ssoRes.data.data.providers;
      }
      if (ssoRes.data.data.is_demo) {
        setRuntimeDemoMode(true);
        isDemo.value = true;
        demoSiteSet(true);
      }
    }
  } catch (e) {
    console.warn('获取 SSO 配置失败:', e);
  }

  // 登录页面加载时获取最新配置
  try {
    await LocalStieConfigUtils.getLatestLocalConfig();
    // 更新页面状态中的站点配置
    pageState.setSiteConfigData(LocalStieConfigUtils.getLocalConfig());
    // 更新当前页面的logo、slogan和网站标题
      const siteConfig = LocalStieConfigUtils.getLocalConfig();
      logo.value = 'data:image/svg+xml;base64,' + btoa(siteConfig.logo);
      slogan.value = siteConfig.slogan;
      // 更新网站标题
      if (siteConfig.title) {
        document.title = siteConfig.title;
      }
      // 更新favicon
      if (siteConfig.logo) {
        updateFavicon(siteConfig.logo);
      }
  } catch (error) {
    console.warn('获取站点配置失败:', error);
  }
});

// 登录
let clickLogin = async () => {
  // 防止重复提交
  if (loading.value) {
    return;
  }
  
  // 验证输入
  if (!account.value.trim() || !password.value.trim()) {
    toast.error('请输入账号和密码');
    return;
  }
  
  loading.value = true;
  
  try {
    const rspe = await request.post('/api/v1/auth', { username: account.value, passwd: password.value });
    const rsp = rspe.data;
    if (rsp.code != 200) {
        toast.error(rsp.msg);
    } else {
        pageState.setToken(rsp.data.token);
        toast.success('登录成功');
        // 延迟一下再跳转，让用户看到成功提示
        setTimeout(() => {
          router.push('/');
        }, 500);
    }
  } catch (error) {
    toast.error('登录失败，请检查网络连接');
  } finally {
    loading.value = false;
  }
};

</script>

<template>
  <div class="flex min-h-svh flex-col items-center justify-center bg-muted p-6 md:p-10">
    <div class="w-full max-w-sm md:max-w-3xl">
      <div class="flex flex-col gap-6">
        <Card class="overflow-hidden">
          <CardContent class="grid p-0 md:grid-cols-2">
            <form class="p-6 md:p-8">
              <div class="flex flex-col gap-6">
                <!-- 小屏时显示logo和标题 -->
                <div class="flex flex-col items-center text-center md:hidden">
                  <div class="w-20 h-20 mb-4 flex justify-center items-center bg-card dark:bg-muted border border-border rounded-full shadow-sm">
                    <img :src="logo" alt="logo" class="w-12 h-12 object-contain" />
                  </div>
                  <h1 class="text-2xl font-bold text-foreground">
                    欢迎回来！
                  </h1>
                  <p class="text-sm text-muted-foreground mt-1">
                    {{ slogan }}
                  </p>
                </div>
                
                <!-- 大屏时显示标题 -->
                <div class="hidden md:flex flex-col items-center text-center">
                  <h1 class="text-2xl font-bold">
                    欢迎回来！
                  </h1>
                  <p class="text-balance text-muted-foreground">
                    {{ slogan }}
                  </p>
                </div>
                <!-- 登录失败 / SSO 提示信息展示条 -->
                <div v-if="loginErrorMessage" class="p-3.5 rounded-xl border border-destructive/30 bg-destructive/10 text-destructive text-xs leading-relaxed flex items-start gap-2.5">
                  <AlertCircle class="w-4 h-4 shrink-0 mt-0.5" />
                  <div class="flex-1">
                    <div class="font-semibold text-xs mb-0.5">登录未成功</div>
                    <p class="break-words">{{ loginErrorMessage }}</p>
                  </div>
                  <button type="button" @click="loginErrorMessage = ''" class="text-destructive/70 hover:text-destructive p-0.5" title="关闭">
                    ✕
                  </button>
                </div>

                <!-- 演示模式专属状态与快捷填入条 -->
                <div v-if="isDemo" class="p-2.5 px-3 rounded-xl border border-primary/20 bg-primary/5 text-xs flex items-center justify-between gap-3">
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

                <div class="grid gap-2">
                  <Label for="account">账号</Label>
                  <Input id="account" type="text" placeholder="" v-model="account" required />
                </div>
                <div class="grid gap-2">
                  <div class="flex items-center">
                    <Label for="password">密码</Label>
                    <!-- <a
                  href="#"
                  class="ml-auto text-sm underline-offset-2 hover:underline"
                >
                  Forgot your password?
                </a> -->
                  </div>
                  <Input id="password" type="password" v-model="password" required />
                </div>
                <Button type="submit" class="w-full active:scale-95 active:bg-gray-100 transition duration-150"
                  @click.prevent="clickLogin" :disabled="loading">
                  <svg v-if="loading" class="mr-2 h-4 w-4 animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  {{ loading ? '登录中...' : '登录' }}
                </Button>

                <!-- 单点登录快捷入口 (OIDC / OAuth 2.0) -->
                <div v-if="ssoProviders && ssoProviders.length > 0" class="space-y-3">
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
                      class="w-full h-10 text-xs font-semibold flex items-center justify-center gap-2 hover:bg-muted transition-all"
                      @click="handleSSOLogin(p)"
                      :disabled="loading"
                    >
                      <Zap v-if="p.type === 'oauth2'" class="w-4 h-4 text-amber-500 shrink-0" />
                      <Fingerprint v-else class="w-4 h-4 text-primary shrink-0" />
                      <span>使用 {{ p.name }} 登录</span>
                    </Button>
                  </div>
                </div>

                <div class="text-center text-sm">
                  还没有账号？
                  <a href="#" class="underline underline-offset-4" @click="handleSubmit">
                    注册
                  </a>
                </div>
              </div>
            </form>
            <div class="relative hidden md:block bg-card">
              <div class="w-full h-full flex justify-center items-center bg-card">
                <img :src="logo" alt="logo" class="max-w-full max-h-full" />
              </div>
            </div>
          </CardContent>
        </Card>
        <!-- <div class="text-balance text-center text-xs text-muted-foreground [&_a]:underline [&_a]:underline-offset-4 hover:[&_a]:text-primary">
      By clicking continue, you agree to our <a href="#">Terms of Service</a>
      and <a href="#">Privacy Policy</a>.
    </div> -->
      </div>
    </div>
  </div>
</template>
