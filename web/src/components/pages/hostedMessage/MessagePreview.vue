<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { request } from '@/api/api'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { AlertCircleIcon, ShieldAlertIcon, MonitorIcon, MoonIcon, SunIcon } from 'lucide-vue-next'

import CryptoJS from 'crypto-js'
import { decompress } from 'fzstd'
import { applyTheme } from '@/util/theme'

const route = useRoute()
const loading = ref(true)
const errMsg = ref('')
const errCode = ref<number | null>(null)

// 存储站点配置用于展示 Header / Footer
interface SiteConfig {
  logo: string
  title: string
  slogan: string
  theme_color: string
}
const siteConfig = ref<SiteConfig | null>(null)

interface PreviewMessage {
  title: string
  content: string
  type: string
  created_on: string
}

const message = ref<PreviewMessage | null>(null)

// 3DES 解密函数，接收 Base64 的密文和 Base64 的密钥
const decryptTripleDes = (cipherTextBase64: string, keyBase64: string): Uint8Array | null => {
  try {
    const key = CryptoJS.enc.Base64.parse(keyBase64)
    const iv = CryptoJS.lib.WordArray.create(key.words.slice(0, 2)) 
    
    const decrypted = CryptoJS.TripleDES.decrypt(
      cipherTextBase64,
      key,
      {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
      }
    )
    
    // Convert CryptoJS WordArray to Uint8Array
    const words = decrypted.words
    const sigBytes = decrypted.sigBytes
    const u8 = new Uint8Array(sigBytes)
    for (let i = 0; i < sigBytes; i++) {
      u8[i] = (words[i >>> 2] >>> (24 - (i % 4) * 8)) & 0xff
    }
    return u8
  } catch (e) {
    console.error('解密失败:', e)
    return null
  }
}

// 格式化类型显示
const typeLabel = computed(() => {
  if (!message.value) return ''
  switch (message.value.type.toLowerCase()) {
    case 'markdown': return 'Markdown'
    case 'html': return 'HTML'
    default: return 'Text'
  }
})

// 正则表达式匹配 URL 并生成超链接
const linkifyText = (text: string) => {
  const urlRegex = /(https?:\/\/[^\s]+)/g
  return text.replace(urlRegex, (url) => {
    return `<a href="${url}" target="_blank" rel="noopener noreferrer" class="text-brand hover:underline break-all">${url}</a>`
  })
}

// 解析 Markdown/HTML/Text 格式内容
const renderedContent = computed(() => {
  if (!message.value) return ''
  const type = message.value.type.toLowerCase()
  const rawContent = message.value.content || ''
  
  if (type === 'markdown') {
    try {
      const parsed = marked.parse(rawContent)
      return DOMPurify.sanitize(parsed as string)
    } catch (e) {
      console.error('Markdown 解析失败:', e)
      return DOMPurify.sanitize(rawContent)
    }
  } else if (type === 'html') {
    return DOMPurify.sanitize(rawContent)
  }
  
  // Text 纯文本高亮 URL
  return DOMPurify.sanitize(linkifyText(rawContent))
})

const fetchPreviewData = async () => {
  loading.value = true
  errMsg.value = ''
  errCode.value = null
  const key = route.params.key as string
  
  try {
    const response = await request.get('/hostedmessages/preview', { params: { key } })
    if (response.data.code === 200) {
      const data = response.data.data
      
      // 还原被混淆的秘钥 (对调前后半段 -> 邻近两两交换)
      const deobfuscateKey = (s: string) => {
        if (!s || s.length < 2) return s
        const arr = s.split('')
        const n = arr.length
        
        // 1. 对调前后半段
        const mid = Math.floor(n / 2)
        for (let i = 0; i < mid; i++) {
          const temp = arr[i]
          arr[i] = arr[mid + i]
          arr[mid + i] = temp
        }
        
        // 2. 邻近两两交换
        for (let i = 0; i < n - 1; i += 2) {
          const temp = arr[i]
          arr[i] = arr[i + 1]
          arr[i + 1] = temp
        }
        
        return arr.join('')
      }
      
      const realKey = deobfuscateKey(data.s)
      
      // 前端解密密文
      const decryptedTitleBytes = decryptTripleDes(data.title, realKey)
      const decryptedContentBytes = decryptTripleDes(data.content, realKey)
      
      const decoder = new TextDecoder('utf-8')
      let decryptedTitle = ''
      let decryptedContent = ''

      if (decryptedTitleBytes) {
        try {
          const decompressedTitle = decompress(decryptedTitleBytes)
          decryptedTitle = decoder.decode(decompressedTitle)
        } catch (e) {
          console.error('Title 解压失败:', e)
        }
      }

      if (decryptedContentBytes) {
        try {
          const decompressedContent = decompress(decryptedContentBytes)
          decryptedContent = decoder.decode(decompressedContent)
        } catch (e) {
          console.error('Content 解压失败:', e)
        }
      }
      
      message.value = {
        title: decryptedTitle,
        content: decryptedContent,
        type: data.type,
        created_on: data.created_on
      }
      
      // 1. 公开预览获取站点配置以获取全局配置、LOGO和主题色
      try {
        const siteRsp = await request.get('/settings/getsetting', { params: { section: 'site_config' } })
        if (siteRsp.data.code === 200) {
          siteConfig.value = siteRsp.data.data
          if (siteConfig.value) {
            if (siteConfig.value.title) {
              document.title = `${siteConfig.value.title} - 托管消息`
            }
            if (siteConfig.value.logo) {
              updateFavicon(siteConfig.value.logo)
            }
            if (siteConfig.value.theme_color) {
              applyTheme(siteConfig.value.theme_color)
            }
          }
        }
      } catch (e) {
        console.warn('获取站点配置失败，使用默认主题', e)
      }
    } else {
      errCode.value = response.data.code || 500
      errMsg.value = response.data.msg || '获取消息失败'
    }
  } catch (error: any) {
    console.error('获取托管消息预览错误:', error)
    if (error.response) {
      errCode.value = error.response.status
      errMsg.value = error.response.data?.msg || '网络错误，请稍后再试'
    } else {
      errMsg.value = '网络请求失败，请检查网络链接'
    }
  } finally {
    loading.value = false
  }
}

const themePreference = ref<'light' | 'dark' | 'system'>('system')
const theme = ref<'light' | 'dark'>('light')

// 初始化并应用深色/浅色偏好（与主站保持一致）
const initThemePreference = () => {
  try {
    const storedPref = localStorage.getItem('themePreference') as 'light' | 'dark' | 'system' | null
    if (storedPref) {
      themePreference.value = storedPref
    }
  } catch (e) {
    console.warn('获取主题偏好失败', e)
  }
  applyThemeFromPreference()
}

const applyThemeFromPreference = () => {
  const systemDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
  const effective: 'light' | 'dark' = themePreference.value === 'system'
    ? (systemDark ? 'dark' : 'light')
    : themePreference.value
  
  theme.value = effective
  
  if (effective === 'dark') {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
  
  try { localStorage.setItem('themePreference', themePreference.value) } catch { }
}

// 更新favicon
const updateFavicon = (logoSvg: string) => {
  if (logoSvg) {
    let link = document.querySelector("link[rel*='icon']") as HTMLLinkElement
    if (!link) {
      link = document.createElement('link')
      link.type = 'image/svg+xml'
      link.rel = 'shortcut icon'
      document.getElementsByTagName('head')[0].appendChild(link)
    }
    
    if (logoSvg.startsWith('data:') || logoSvg.startsWith('http')) {
      link.href = logoSvg
    } else {
      const encodedSvg = encodeURIComponent(logoSvg)
      link.href = `data:image/svg+xml,${encodedSvg}`
    }
  }
}

const toggleTheme = () => {
  themePreference.value = themePreference.value === 'light' ? 'dark' : (themePreference.value === 'dark' ? 'system' : 'light')
  applyThemeFromPreference()
}

const themeLabel = computed(() => {
  if (themePreference.value === 'system') return '跟随系统'
  return theme.value === 'dark' ? '深色' : '浅色'
})

onMounted(() => {
  initThemePreference()
  fetchPreviewData()
})
</script>

<template>
  <div class="min-h-screen bg-background flex flex-col items-center justify-between p-4 sm:p-6 md:p-8 relative">
    
    <!-- Premium background gradient mesh -->
    <div class="fixed inset-0 overflow-hidden pointer-events-none z-0">
      <!-- Top Left Brand Glow -->
      <div class="absolute -top-[10%] -left-[10%] w-[60vw] h-[60vw] max-w-[600px] max-h-[600px] rounded-full bg-brand/20 blur-[100px] dark:bg-brand/10"></div>
      <!-- Right Side Subtle Glow -->
      <div class="absolute top-[30%] -right-[10%] w-[50vw] h-[50vw] max-w-[500px] max-h-[500px] rounded-full bg-blue-500/10 blur-[100px] dark:bg-blue-900/10"></div>
    </div>

    <div class="w-full max-w-4xl mt-2 sm:mt-6 flex flex-col space-y-6 relative z-10">
      
      <!-- Premium Header: Site Logo and Title -->
      <div v-if="siteConfig" class="flex items-center justify-between w-full px-1">
        <div class="flex items-center space-x-3">
          <!-- 站点 Logo -->
          <div 
            v-if="siteConfig.logo" 
            class="w-12 h-12 rounded-xl bg-card border border-border shadow-sm flex items-center justify-center overflow-hidden [&_svg]:w-7 [&_svg]:h-7 [&_svg]:fill-brand [&_path]:fill-brand"
            v-html="siteConfig.logo"
          ></div>
          <div class="flex flex-col">
            <span class="text-lg font-bold text-foreground tracking-tight leading-tight">{{ siteConfig.title }}</span>
            <span class="text-xs text-muted-foreground mt-0.5">{{ siteConfig.slogan }}</span>
          </div>
        </div>
        
        <!-- 右侧优雅时间戳与类型展示 -->
        <div class="flex items-center space-x-3">
          <div v-if="message" class="hidden sm:flex items-center space-x-3">
            <Badge variant="secondary" class="font-bold bg-brand/10 text-brand border border-brand/20 px-3 py-1.5 text-xs rounded-full shadow-sm select-none">
              # {{ typeLabel }}
            </Badge>
            <div class="px-4 py-1.5 bg-card border border-border rounded-full shadow-sm text-xs font-semibold text-foreground/80 tracking-wider">
              {{ message.created_on }}
            </div>
          </div>
          
          <!-- 主题切换按钮 -->
          <button @click="toggleTheme" :title="'切换主题（当前：' + themeLabel + '）'"
            class="p-2 rounded-full border border-border bg-card text-muted-foreground hover:text-brand hover:bg-muted/50 transition-colors">
            <MonitorIcon v-if="themePreference === 'system'" class="w-4 h-4" />
            <MoonIcon v-else-if="theme === 'dark'" class="w-4 h-4" />
            <SunIcon v-else class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- 骨架屏加载中 -->
      <Card v-if="loading" class="w-full border-border bg-card/60 backdrop-blur-md shadow-lg">
        <CardHeader class="space-y-4">
          <div class="h-8 w-2/3 bg-slate-200 dark:bg-slate-800 rounded animate-pulse"></div>
          <div class="flex items-center space-x-2">
            <div class="h-4 w-24 bg-slate-200 dark:bg-slate-800 rounded animate-pulse"></div>
            <div class="h-4 w-16 bg-slate-200 dark:bg-slate-800 rounded animate-pulse"></div>
          </div>
        </CardHeader>
        <CardContent class="space-y-3">
          <div class="h-4 w-full bg-slate-200 dark:bg-slate-800 rounded animate-pulse"></div>
          <div class="h-4 w-full bg-slate-200 dark:bg-slate-800 rounded animate-pulse"></div>
          <div class="h-4 w-3/4 bg-slate-200 dark:bg-slate-800 rounded animate-pulse"></div>
        </CardContent>
      </Card>

      <!-- 错误页面 (例如被禁止访问或者消息不存在) -->
      <Card v-else-if="errMsg || !message" class="w-full border-red-200/50 dark:border-red-950/50 bg-white/80 dark:bg-slate-900/80 backdrop-blur-md shadow-2xl overflow-hidden">
        <div class="p-8 text-center flex flex-col items-center justify-center space-y-4">
          <div class="w-16 h-16 rounded-full bg-red-50 dark:bg-red-950/30 flex items-center justify-center border border-red-100 dark:border-red-900">
            <ShieldAlertIcon v-if="errCode === 403" class="w-8 h-8 text-red-600 dark:text-red-400" />
            <AlertCircleIcon v-else class="w-8 h-8 text-red-600 dark:text-red-400" />
          </div>
          <h2 class="text-2xl font-bold text-slate-950 dark:text-white">
            {{ errCode === 403 ? '无访问权限' : '获取消息失败' }}
          </h2>
          <p class="text-slate-600 dark:text-slate-400 text-sm max-w-md">
            {{ errMsg || '该消息不存在或已被主人删除。' }}
          </p>
        </div>
      </Card>

      <!-- 核心设计感主体卡片 -->
      <div v-else class="flex flex-col space-y-4 sm:space-y-6">
        <!-- 移动端额外呈现时间戳与类型 -->
        <div class="flex sm:hidden justify-start px-2 items-center space-x-2">
          <Badge variant="secondary" class="font-bold bg-brand/10 text-brand border border-brand/20 px-3 py-1 text-xs rounded-full shadow-sm select-none">
            # {{ typeLabel }}
          </Badge>
          <span class="text-xs font-semibold text-muted-foreground/80 bg-card border border-border px-3 py-1 rounded-full shadow-sm">{{ message.created_on }}</span>
        </div>

        <Card class="w-full border border-border/80 bg-card/65 backdrop-blur-xl shadow-2xl overflow-hidden rounded-3xl transition-all duration-300 hover:shadow-brand/5">
          <CardContent class="p-6 md:p-8 flex flex-col space-y-4">
            
            <!-- 标题展示（仅在不是无意义默认标题时渲染） -->
            <h2 v-if="message.title && message.title !== message.content" class="text-lg sm:text-xl font-bold text-foreground leading-snug break-all tracking-tight pb-2 border-b border-border/40">
              {{ message.title }}
            </h2>

            <!-- 渲染的真实数据正文 -->
            <div class="text-foreground/90 break-all overflow-x-auto">
              <!-- HTML / Markdown 渲染 -->
              <div 
                v-if="message.type.toLowerCase() === 'html' || message.type.toLowerCase() === 'markdown'" 
                class="prose prose-slate dark:prose-invert max-w-none break-words text-sm"
                v-html="renderedContent"
              ></div>
              
              <!-- Text 纯文本渲染（支持高亮链接） -->
              <div 
                v-else 
                class="whitespace-pre-wrap font-sans text-sm leading-relaxed"
                v-html="renderedContent"
              ></div>
            </div>
          </CardContent>
        </Card>
      </div>

    </div>

    <!-- Premium Footer: Site Info and Copyright -->
    <div v-if="siteConfig" class="text-center py-8 text-xs text-muted-foreground select-none">
      Powered by <span class="text-brand font-medium">{{ siteConfig.title }}</span> &copy; 2026
    </div>
  </div>
</template>

<style scoped>
/* 确保高亮链接文本颜色采用系统设置的动态品牌主题色 */
.text-brand {
  color: var(--brand);
}

/* 针对 prose 样式的微调，确保 Markdown 表格和图片等良好适配 */
:deep(.prose) {
  font-family: inherit;
}
:deep(.prose pre) {
  background-color: var(--color-slate-100, #f1f5f9);
  color: var(--color-slate-900, #0f172a);
  padding: 1rem;
  border-radius: 0.5rem;
  overflow-x: auto;
}
.dark :deep(.prose pre) {
  background-color: var(--color-slate-800, #1e293b);
  color: var(--color-slate-100, #f1f5f9);
}
:deep(.prose img) {
  max-width: 100%;
  height: auto;
  border-radius: 0.375rem;
}
:deep(.prose table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1rem 0;
}
:deep(.prose th), :deep(.prose td) {
  border: 1px solid var(--color-slate-200, #e2e8f0);
  padding: 0.5rem 0.75rem;
}
.dark :deep(.prose th), .dark :deep(.prose td) {
  border-color: var(--color-slate-700, #334155);
}
</style>

