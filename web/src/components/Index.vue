<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
// 使用 defineAsyncComponent 导入没有默认导出的组件
// const Dashboard = defineAsyncComponent(() => import("@/components/pages/dashboard/Dashboard.vue"))
// const SendLogs = defineAsyncComponent(() => import("@/components/pages/sendLogs/SendLogs.vue"))
// 导入 constant 模块，使用命名导入匹配 constant.js 的导出方式
import { CONSTANT } from '../constant.js'
import { LocalStieConfigUtils } from '@/util/localSiteConfig'
import { usePageState } from '@/store/page_sate.js'

import { useRoute, useRouter } from 'vue-router'

// 定义标签接口
interface TabRoute {
  name: string;
  path: string;
}

const route = useRoute()
const router = useRouter()
const pageState = usePageState()
const isAuthenticated = ref(Boolean(localStorage.getItem(CONSTANT.STORE_TOKEN_NAME)));
const isMobileMenuOpen = ref(false)
const isUserMenuOpen = ref(false)
const userAccount = ref('管理员')
const siteConfig = ref<any>({})

// 主题：明暗模式与跟随系统
type ThemePreference = 'light' | 'dark' | 'system'

const getInitialThemePreference = (): ThemePreference => {
  try {
    const storedPref = localStorage.getItem('themePreference') as ThemePreference | null
    if (storedPref === 'light' || storedPref === 'dark' || storedPref === 'system') return storedPref
    const legacy = localStorage.getItem('theme') as 'light' | 'dark' | null
    if (legacy === 'light' || legacy === 'dark') return legacy
    return 'system'
  } catch {
    return 'system'
  }
}

const themePreference = ref<ThemePreference>(getInitialThemePreference())
const theme = ref<'light' | 'dark'>('light')

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
  try { localStorage.setItem('themePreference', themePreference.value) } catch {}
}

const toggleTheme = () => {
  themePreference.value = themePreference.value === 'light' ? 'dark' : (themePreference.value === 'dark' ? 'system' : 'light')
  applyThemeFromPreference()
}

const themeLabel = computed(() => {
  if (themePreference.value === 'system') return '跟随系统'
  return theme.value === 'dark' ? '深色' : '浅色'
})

// 从JWT中解析用户名
const parseJwtUsername = (token: string): string => {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.username || payload.user || payload.name || '管理员'
  } catch (error) {
    console.error('解析JWT失败:', error)
    return '管理员'
  }
}

// 更新用户账号信息
const updateUserAccount = () => {
  const token = localStorage.getItem(CONSTANT.STORE_TOKEN_NAME)
  if (token) {
    userAccount.value = parseJwtUsername(token)
  } else {
    userAccount.value = '管理员'
  }
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

// 获取本地配置
const getLocalConfig = () => {
  try {
    const localConfig = LocalStieConfigUtils.getLocalConfig()
    if (localConfig) {
      siteConfig.value = localConfig
      // 更新页面状态中的配置数据
      if (pageState.setSiteConfigData) {
        pageState.setSiteConfigData(localConfig)
      }
      // 更新网站标题
      if (localConfig.title) {
        document.title = localConfig.title
      }
      // 更新favicon
      if (localConfig.logo) {
        updateFavicon(localConfig.logo)
      }
      
    }
  } catch (error) {
    console.error('获取本地配置失败:', error)
  }
}

// 获取最新配置并更新
const getLatestConfig = async () => {
  try {
    const latestConfig = await LocalStieConfigUtils.getLatestLocalConfig()
    if (latestConfig) {
      siteConfig.value = latestConfig
      // 更新页面状态中的配置数据
      if (pageState.setSiteConfigData) {
        pageState.setSiteConfigData(latestConfig)
      }
      // 更新网站标题
      if (latestConfig.title) {
        document.title = latestConfig.title
      }
      // 更新favicon
      if (latestConfig.logo) {
        updateFavicon(latestConfig.logo)
      }
      
    }
  } catch (error) {
    console.error('获取最新配置失败:', error)
    // 如果获取最新配置失败，尝试使用本地配置
    getLocalConfig()
  }
}

// 切换移动端菜单
const toggleMobileMenu = () => {
  isMobileMenuOpen.value = !isMobileMenuOpen.value
}

// 切换用户菜单
const toggleUserMenu = () => {
  isUserMenuOpen.value = !isUserMenuOpen.value
}

// 退出登录
const logout = () => {
  localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME)
  isAuthenticated.value = false
  isUserMenuOpen.value = false
  router.push('/login')
}

// 监听localStorage变化
onMounted(() => {
  // 初始化主题并监听系统主题变化
  applyThemeFromPreference()
  try {
    const media = window.matchMedia('(prefers-color-scheme: dark)')
    const handleSystemChange = () => {
      if (themePreference.value === 'system') applyThemeFromPreference()
    }
    // 新浏览器
    if (media.addEventListener) {
      media.addEventListener('change', handleSystemChange)
    } else if ((media as any).addListener) {
      // 兼容旧浏览器
      ;(media as any).addListener(handleSystemChange)
    }
  } catch {}

  // 初始化用户账号信息
  updateUserAccount();
  
  // 初始化配置信息
  getLocalConfig();
  
  // 如果已认证，获取最新配置
  if (isAuthenticated.value) {
    getLatestConfig();
  }
  
  // 定期检查token状态
  const checkAuth = () => {
    const wasAuthenticated = isAuthenticated.value;
    isAuthenticated.value = Boolean(localStorage.getItem(CONSTANT.STORE_TOKEN_NAME));
    // 如果认证状态发生变化，更新用户账号信息和配置
    if (wasAuthenticated !== isAuthenticated.value) {
      updateUserAccount();
      if (isAuthenticated.value) {
        // 用户刚登录，获取最新配置
        getLatestConfig();
      } else {
        // 用户退出登录，使用本地配置
        getLocalConfig();
      }
    }
  };
  // 监听storage事件
  window.addEventListener('storage', checkAuth);
  // 定期检查（处理同一页面内的变化）
  const interval = setInterval(checkAuth, 1000);
  
  // 点击外部关闭用户菜单
  const handleClickOutside = (event: Event) => {
    const target = event.target as Element;
    if (!target.closest('.user-menu-container')) {
      isUserMenuOpen.value = false;
    }
  };
  document.addEventListener('click', handleClickOutside);
  
  // 清理函数
  return () => {
    window.removeEventListener('storage', checkAuth);
    clearInterval(interval);
    document.removeEventListener('click', handleClickOutside);
  };
});

// 定义标签及其对应的路由路径
const tabRoutes: TabRoute[] = [
  { name: '数据统计', path: '/' },
  { name: '发信日志', path: '/sendlogs' },
  { name: '托管消息', path: '/hostedmessage' },
  { name: '定时消息', path: '/cronmessages' },
  { name: '发信任务', path: '/sendtasks' },
  { name: '发信渠道', path: '/sendways' },
  { name: '设置偏好', path: '/settings' }
];
const activeTab = ref('Dashboard');

// 处理标签点击事件，跳转到对应路由
const handleTabClick = (tab: TabRoute) => {
  activeTab.value = tab.name;
  

  // 使用 Vue Router 进行导航，保持单页应用状态
  // 对于根路径，直接导航；对于其他路径，使用相对路径
  if (tab.path === '/') {
    router.push(tab.path);
  } else {
    // 去掉开头的斜杠，使用相对路径导航
    const relativePath = tab.path.startsWith('/') ? tab.path.substring(1) : tab.path;
    router.push(relativePath);
  }
};


// 根据当前路由设置活动标签
const updateActiveTab = () => {
  // 获取当前路径（去掉开头的斜杠）
  const currentPath = route.path.startsWith('/') ? route.path.substring(1) : route.path;

  // 查找当前路由对应的标签
  const currentTab = tabRoutes.find(tab => {
    // 对于根路径，需要精确匹配
    if (tab.path === '/') {
      return route.path === '/' || route.path === '';
    }

    // 对于其他路径，检查当前路由是否匹配
    // 去掉开头的斜杠进行比较
    const tabPath = tab.path.startsWith('/') ? tab.path.substring(1) : tab.path;
    return currentPath === tabPath || route.path === tab.path;
  });

  // 如果找到对应标签，更新活动标签
  if (currentTab) {
    activeTab.value = currentTab.name;
  }
};

// 初始化时设置活动标签
updateActiveTab();

// 监听路由变化并更新活动标签
watch(() => route.path, updateActiveTab);

// 计算属性：站点标题
const siteTitle = computed(() => {
  return siteConfig.value?.title || '消息管理系统'
})
</script>


<template>
  <router-view v-if="!isAuthenticated || route.path == '/login' || route.path == 'login'"></router-view>

  <div class="min-h-screen bg-background" v-else>
    <!-- 顶部导航栏 -->
    <nav class="bg-background shadow-sm border-b border-border">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <!-- Logo/品牌名称 -->
          <div class="flex items-center">
            <h1 class="text-lg sm:text-xl font-bold text-foreground">{{ siteTitle }}</h1>
          </div>

          <!-- 桌面端导航 -->
          <div class="hidden md:flex space-x-6 lg:space-x-8">
            <button v-for="tab in tabRoutes" :key="tab.name" @click="handleTabClick(tab)" :class="[
              'relative py-2 px-3 text-sm font-medium transition-all duration-200 rounded-md whitespace-nowrap',
              activeTab === tab.name
                ? 'text-blue-600 bg-blue-50 dark:text-blue-400 dark:bg-blue-400/10'
                : 'text-gray-600 hover:text-blue-600 hover:bg-gray-50 dark:text-gray-300 dark:hover:text-blue-400 dark:hover:bg-white/5'
            ]">
              {{ tab.name }}
              <!-- 活动状态指示器 -->
              <span v-if="activeTab === tab.name"
                class="absolute bottom-0 left-1/2 transform -translate-x-1/2 w-1 h-1 bg-blue-600 rounded-full"></span>
            </button>
          </div>

          <!-- 右侧区域 -->
          <div class="flex items-center space-x-4">
            <!-- 主题切换（仅桌面显示） -->
            <button @click="toggleTheme" class="hidden md:inline-flex p-2 rounded-md text-gray-600 hover:text-blue-600 hover:bg-gray-50 transition-colors dark:text-gray-300 dark:hover:text-blue-400 dark:hover:bg-white/5" :title="theme === 'dark' ? '切换到浅色' : '切换到深色'">
              <svg v-if="theme === 'dark'" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
              </svg>
              <svg v-else class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="5"/>
                <line x1="12" y1="1" x2="12" y2="3"/>
                <line x1="12" y1="21" x2="12" y2="23"/>
                <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
                <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
                <line x1="1" y1="12" x2="3" y2="12"/>
                <line x1="21" y1="12" x2="23" y2="12"/>
                <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
                <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
              </svg>
            </button>
            <!-- 用户账号下拉菜单 -->
            <div class="relative user-menu-container">
              <button @click="toggleUserMenu" class="flex items-center space-x-2 p-2 rounded-md hover:bg-muted transition-colors dark:hover:bg-white/5">
                <div class="w-8 h-8 bg-blue-100 dark:bg-muted border border-border rounded-full flex items-center justify-center">
                  <svg class="w-4 h-4 text-blue-600 dark:text-blue-300" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd" />
                  </svg>
                </div>
                <span class="hidden sm:block text-sm font-medium text-foreground">{{ userAccount }}</span>
                <svg class="w-4 h-4 text-gray-400 dark:text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              
              <!-- 下拉菜单 -->
              <div v-if="isUserMenuOpen" class="absolute right-0 mt-2 w-52 bg-card text-foreground rounded-md shadow-lg border border-border z-50">
                <div class="py-1">
                  <div class="px-4 py-2 text-sm text-muted-foreground border-b border-border">
                    <div class="font-medium text-foreground">{{ userAccount }}</div>
                    <!-- <div class="text-xs">当前登录账号</div> -->
                  </div>
                  <!-- 移动端主题切换入口 -->
                  <button @click="toggleTheme" class="md:hidden w-full text-left px-4 py-2 text-sm hover:bg-muted dark:hover:bg-white/5 flex items-center justify-between">
                    <span class="truncate">外观</span>
                    <span class="text-xs text-muted-foreground flex-shrink-0">{{ themeLabel }}</span>
                  </button>
                  <button @click="logout" class="w-full text-left px-4 py-2 text-sm text-destructive hover:bg-muted dark:hover:bg-white/5 transition-colors">
                    <div class="flex items-center space-x-2">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                      </svg>
                      <span>退出登录</span>
                    </div>
                  </button>
                </div>
              </div>
            </div>

            <!-- 移动端菜单按钮 -->
            <button @click="toggleMobileMenu"
              class="md:hidden p-2 rounded-md text-gray-600 hover:text-blue-600 hover:bg-gray-50 transition-colors">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path v-if="!isMobileMenuOpen" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M4 6h16M4 12h16M4 18h16" />
                <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <!-- 移动端菜单 -->
        <div v-if="isMobileMenuOpen" class="md:hidden border-t border-border py-2">
          <div class="flex flex-col space-y-1">
            <button v-for="tab in tabRoutes" :key="tab.name" @click="handleTabClick(tab); isMobileMenuOpen = false"
              :class="[
                'text-left py-3 px-4 text-sm font-medium transition-all duration-200 rounded-md',
                activeTab === tab.name
                  ? 'text-blue-600 bg-blue-50 dark:text-blue-400 dark:bg-blue-400/10'
                  : 'text-muted-foreground hover:text-blue-600 hover:bg-muted dark:hover:bg-white/5'
              ]">
              {{ tab.name }}
            </button>
          </div>
        </div>
      </div>
    </nav>

    <!-- 主要内容区域 -->
    <main class="max-w-7xl mx-auto py-2 px-4 sm:px-2 lg:px-8">
      <!-- <div class="bg-white rounded-lg shadow-sm border border-gray-200"> -->
      <div>
        <!-- 使用 router-view 显示嵌套路由的内容 -->
        <router-view></router-view>
      </div>
    </main>
  </div>

</template>
