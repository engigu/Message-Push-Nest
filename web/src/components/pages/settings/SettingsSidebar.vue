<script setup lang="ts">
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { KeyIcon, TrashIcon, SettingsIcon, InfoIcon } from 'lucide-vue-next'

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

// 设置菜单项
const settingsMenu = [
  { id: 'password', name: '重置密码', icon: KeyIcon },
  { id: 'logs', name: '日志清理', icon: TrashIcon },
  { id: 'site', name: '站点设置', icon: SettingsIcon },
  { id: 'about', name: '站点关于', icon: InfoIcon }
]
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
              'w-full flex items-center px-4 py-3 text-left text-sm font-medium transition-colors',
              activeTab === item.id
                ? 'bg-blue-50 text-blue-700 border-r-2 border-blue-700'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
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