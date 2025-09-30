<script setup lang="ts">
import { onMounted, ref } from 'vue'
// @ts-ignore
import { request } from '@/api/api'

interface LoginLog {
  id: number
  user_id: number
  username: string
  ip: string
  ua: string
  created_on: string
}

const loading = ref(false)
const logs = ref<LoginLog[]>([])

const fetchLogs = async () => {
  loading.value = true
  try {
    const rsp = await request.get('/loginlogs/recent')
    const data = rsp.data
    if (data && data.code === 200 && data.data) {
      logs.value = data.data.lists || []
    }
  } finally {
    loading.value = false
  }
}

onMounted(fetchLogs)
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold">最近登录日志</h2>
      <button class="text-sm text-muted-foreground hover:text-foreground" @click="fetchLogs" :disabled="loading">
        {{ loading ? '刷新中...' : '刷新' }}
      </button>
    </div>

    <div class="overflow-hidden rounded-md border border-border">
      <table class="w-full text-sm">
        <thead>
          <tr class="bg-muted">
            <th class="px-3 py-2 text-left">用户名</th>
            <th class="px-3 py-2 text-left">IP</th>
            <th class="px-3 py-2 text-left">UA</th>
            <th class="px-3 py-2 text-left">登录时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="!logs.length">
            <td colspan="4" class="px-3 py-6">
              <div class="flex items-center justify-center">
                <span class="text-sm text-muted-foreground">暂无数据</span>
              </div>
            </td>
          </tr>
          <tr v-for="item in logs" :key="item.id" class="border-t border-border">
            <td class="px-3 py-2">{{ item.username }}</td>
            <td class="px-3 py-2">{{ item.ip }}</td>
            <td class="px-3 py-2 truncate max-w-[420px]" :title="item.ua">{{ item.ua }}</td>
            <td class="px-3 py-2">{{ item.created_on }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>


