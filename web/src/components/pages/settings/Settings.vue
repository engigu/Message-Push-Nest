<script setup lang="ts">
import { ref } from 'vue'
import SettingsSidebar from './SettingsSidebar.vue'
import PasswordSettings from './PasswordSettings.vue'
import LogsSettings from './LogsSettings.vue'
import SiteSettings from './SiteSettings.vue'
import AboutSettings from './AboutSettings.vue'
import LoginLogs from './LoginLogs.vue'

// 当前选中的设置项
const activeTab = ref('password')
</script>

<template>
  <div class="p-6 w-full max-w-6xl mx-auto">
    <div class="flex flex-col lg:flex-row gap-6">
      <!-- 侧边栏 -->
      <SettingsSidebar 
        :active-tab="activeTab" 
        @update:active-tab="activeTab = $event" 
      />

      <!-- 主内容区域 -->
      <div class="flex-1">
        <!-- 重置密码 -->
        <PasswordSettings v-if="activeTab === 'password'" />

        <!-- 日志清理 -->
        <LogsSettings v-if="activeTab === 'logs'" />

        <!-- 站点设置 -->
        <SiteSettings v-if="activeTab === 'site'" />

        <!-- 登录日志 -->
        <LoginLogs v-if="activeTab === 'loginlogs'" />

        <!-- 站点关于 -->
        <AboutSettings v-if="activeTab === 'about'" />
      </div>
    </div>
  </div>
</template>