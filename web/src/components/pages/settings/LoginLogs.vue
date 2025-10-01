<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogHeader } from '@/components/ui/dialog'
import { toast } from 'vue-sonner'

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
const ipDialogOpen = ref(false)
const ipLoading = ref(false)
const selectedIp = ref('')
const ipInfo = ref<any>(null)

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

const openIpInfo = async (ip: string) => {
  selectedIp.value = ip
  ipDialogOpen.value = true
  ipLoading.value = true
  ipInfo.value = null
  try {
    const rsp = await fetch(`https://api.ip.sb/geoip/${encodeURIComponent(ip)}`)
    if (!rsp.ok) throw new Error('请求失败')
    const data = await rsp.json()
    ipInfo.value = data
  } catch (e) {
    toast.error('获取IP信息失败')
  } finally {
    ipLoading.value = false
  }
}

const formatUtcOffset = (offset: any) => {
  const num = typeof offset === 'number' ? offset : parseInt(offset, 10)
  if (Number.isNaN(num)) return '-'
  const hours = Math.floor(num / 3600)
  const sign = hours >= 0 ? '+' : '-'
  const abs = Math.abs(hours)
  return `UTC${sign}${abs}`
}

const ipDisplayRows = computed(() => {
  if (!ipInfo.value) return [] as Array<{ label: string, value: string }>
  const info = ipInfo.value
  const rows: Array<{ label: string, value: string }> = []

  const ip = info.ip || selectedIp.value
  if (ip) rows.push({ label: 'IP', value: String(ip) })

  const country = [info.country, info.country_code ? `(${info.country_code})` : ''].filter(Boolean).join(' ')
  if (country) rows.push({ label: '国家/地区', value: country })

  const tz = [info.timezone, info.offset != null ? formatUtcOffset(info.offset) : ''].filter(Boolean).join(' · ')
  if (tz) rows.push({ label: '时区', value: tz })

  if (info.isp) rows.push({ label: 'ISP', value: String(info.isp) })
  if (info.organization) rows.push({ label: '组织', value: String(info.organization) })

  const asn = [info.asn, info.asn_organization].filter(Boolean).join(' - ')
  if (asn) rows.push({ label: 'ASN', value: asn })

  const continent = info.continent_code
  if (continent) rows.push({ label: '大洲', value: String(continent) })

  const coordParts = [] as string[]
  if (info.latitude != null) coordParts.push(String(info.latitude))
  if (info.longitude != null) coordParts.push(String(info.longitude))
  const coord = coordParts.join(', ')
  if (coord) rows.push({ label: '坐标', value: coord })

  return rows
})
</script>

<template>
  <Card>
    <CardHeader class="flex items-start sm:items-center justify-between gap-2">
      <div class="min-w-0">
        <CardTitle>最近登录日志</CardTitle>
        <CardDescription class="mt-1">展示最近 8 条登录记录</CardDescription>
      </div>
      <button class="text-sm text-muted-foreground hover:text-foreground flex-shrink-0" @click="fetchLogs" :disabled="loading">
        {{ loading ? '刷新中...' : '刷新' }}
      </button>
    </CardHeader>
    <CardContent>
      <div class="overflow-x-auto" style="will-change: transform;">
        <div class="min-w-full">
          <table class="w-full text-sm border-separate border-spacing-0">
          <thead>
            <tr>
              <th class="px-3 py-2 text-left border-b border-border">用户名</th>
              <th class="px-3 py-2 text-left border-b border-border">IP</th>
              <th class="px-3 py-2 text-left border-b border-border">UA</th>
              <th class="px-3 py-2 text-left border-b border-border">登录时间</th>
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
            <tr v-for="item in logs" :key="item.id" class="border-b border-border">
              <td class="px-3 py-2">{{ item.username }}</td>
              <td class="px-3 py-2">
                <button class="text-blue-600 dark:text-blue-400 hover:underline" @click="openIpInfo(item.ip)">{{ item.ip }}</button>
              </td>
              <td class="px-3 py-2 truncate max-w-[220px] sm:max-w-[420px]" :title="item.ua">{{ item.ua }}</td>
              <td class="px-3 py-2">{{ item.created_on }}</td>
            </tr>
          </tbody>
        </table>
        </div>
      </div>
    </CardContent>
  </Card>

  <Dialog :open="ipDialogOpen" @update:open="val => (ipDialogOpen = val)">
    <DialogContent class="w-[90vw] max-w-[90vw] sm:max-w-lg max-h-[80vh] overflow-y-auto">

        <DialogHeader>
    <!-- <VisuallyHidden> -->
      <!-- <DialogTitle>IP 信息</DialogTitle> -->
    <!-- </VisuallyHidden> -->
  </DialogHeader>
      
      <div v-if="ipLoading" class="text-sm text-muted-foreground">加载中...</div>
      <div v-else class="space-y-3 text-sm">
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-2">
          <template v-for="(row, idx) in ipDisplayRows" :key="idx">
            <div class="text-muted-foreground">{{ row.label }}</div>
            <div class="break-all" :class="{ 'font-mono': row.label === 'IP' || row.label === '坐标' }">{{ row.value }}</div>
          </template>
        </div>
        <div class="text-xs text-muted-foreground mt-2">数据来源：<a :href="'https://api.ip.sb/geoip/' + encodeURIComponent(selectedIp)" target="_blank" rel="noreferrer" class="underline">api.ip.sb</a></div>
      </div>
    </DialogContent>
  </Dialog>
</template>


