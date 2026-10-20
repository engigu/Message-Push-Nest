<script setup lang="ts">
import { computed } from 'vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { KeyIcon, TrashIcon, SettingsIcon, InfoIcon, HistoryIcon, FingerprintIcon, UsersIcon } from 'lucide-vue-next'
import { usePageState } from '@/store/page_sate'

// 定义props
interface Props {
  activeTab: string
}

// 定义emits
interface Emits {
  (e: 'update:activeTab', value: string): void
}

defineProps<Props>()
defineEmits<Emits>()

const pageState = usePageState()

// 过滤后的设置菜单项
const settingsMenu = computed(() => {
  const baseMenu = [
    { id: 'password', name: '账号安全', icon: KeyIcon, adminOnly: false },
    { id: 'users', name: '用户管理', icon: UsersIcon, adminOnly: true },
    { id: 'oidc', name: '单点登录', icon: FingerprintIcon, adminOnly: true },
    { id: 'clean', name: '数据清理', icon: TrashIcon, adminOnly: false },
    { id: 'loginlogs', name: '登录日志', icon: HistoryIcon, adminOnly: true },
    { id: 'site', name: '站点设置', icon: SettingsIcon, adminOnly: true },
    { id: 'about', name: '站点关于', icon: InfoIcon, adminOnly: false }
  ]

  if (pageState.isAdmin) {
    return baseMenu
  }
  return baseMenu.filter(item => !item.adminOnly)
})
</script>

<script lang="ts">
export default {
  name: 'SettingsSidebar'
}
</script>

<template>
  <div class="lg:w-64 flex-shrink-0">
    <Card>
      <CardHeader>
        <CardTitle class="text-lg">设置</CardTitle>
        <CardDescription>管理系统配置和偏好</CardDescription>
      </CardHeader>
      <CardContent class="p-0">
        <nav class="space-y-1">
          <button
            v-for="item in settingsMenu"
            :key="item.id"
            @click="$emit('update:activeTab', item.id)"
            :class="[
              'w-full flex items-center px-4 py-3 text-left text-sm font-medium transition-colors rounded-md',
              activeTab === item.id
                ? 'bg-accent text-accent-foreground border-r-2 border-primary dark:bg-accent/40'
                : 'text-muted-foreground hover:bg-muted hover:text-foreground dark:hover:bg-white/5'
            ]"
          >
            <component :is="item.icon" class="mr-3 w-5 h-5" />
            {{ item.name }}
          </button>
        </nav>
      </CardContent>
    </Card>
  </div>
</template>