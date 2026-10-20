<script setup lang="ts">
import { onMounted, ref, reactive } from 'vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { PlusIcon, SearchIcon, RefreshCwIcon, KeyRoundIcon, PencilIcon, Trash2Icon, ShieldAlertIcon, FingerprintIcon } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { usePageState } from '@/store/page_sate'

// @ts-ignore
import { request } from '@/api/api'

interface UserItem {
  id: string
  username: string
  email: string
  role: string
  status: number
  oidc_subject?: string
  auth_provider?: string
  is_oidc?: boolean
  login_type?: string
  created_by: string
  created_on: string
}

const pageState = usePageState()
const loading = ref(false)
const users = ref<UserItem[]>([])
const total = ref(0)
const search = ref('')
const page = ref(1)
const size = ref(10)

// 格式化时间，兼容 0001-01-01 零值
const formatTime = (timeStr?: string) => {
  if (!timeStr || timeStr.startsWith('0001-01-01')) {
    return '-'
  }
  return timeStr
}

// 格式化外部身份源名称
const formatProviderName = (provider?: string) => {
  if (!provider) return 'OIDC'
  const p = provider.toLowerCase()
  if (p === 'oidc') return 'OIDC'
  if (p === 'oauth2') return 'OAuth 2.0'
  return provider
}

// 格式化集成登录徽章文本
const formatSsoBadge = (provider?: string) => {
  const name = formatProviderName(provider)
  return `集成登录 · ${name}`
}

// 模态框控制
const addDialogOpen = ref(false)
const editDialogOpen = ref(false)
const resetPasswdDialogOpen = ref(false)
const deleteDialogOpen = ref(false)
const actionLoading = ref(false)

// 新增表单
const addForm = reactive({
  username: '',
  password: '',
  email: '',
  role: 'user'
})
const addUsernameError = ref('')

const checkAddUsername = async () => {
  const name = addForm.username.trim()
  if (!name) {
    addUsernameError.value = ''
    return
  }
  if (name.length < 2 || name.length > 30) {
    addUsernameError.value = '用户名长度需在 2 到 30 个字符之间'
    return
  }
  try {
    const res = await request.get('/api/v1/auth/check-username', { params: { username: name } })
    if (res.data && res.data.code === 200 && res.data.data) {
      if (!res.data.data.available) {
        addUsernameError.value = res.data.data.reason || '该用户名已被使用'
      } else {
        addUsernameError.value = ''
      }
    }
  } catch (e) {}
}

// 编辑表单
const editForm = reactive({
  id: '',
  username: '',
  email: '',
  role: 'user',
  status: 1,
  is_oidc: false,
  auth_provider: ''
})

// 重置密码表单
const resetPasswdForm = reactive({
  id: '',
  username: '',
  password: ''
})
const targetResetUserIsOidc = ref(false)
const targetResetUserProvider = ref('')

// 待删除用户
const targetDeleteUser = ref<UserItem | null>(null)

// 获取用户列表
const fetchUsers = async () => {
  loading.value = true
  try {
    const rsp = await request.get('/api/v1/users', {
      params: {
        search: search.value,
        page: page.value,
        size: size.value
      }
    })
    const data = rsp.data
    if (data && data.code === 200 && data.data) {
      users.value = data.data.lists || []
      total.value = data.data.total || 0
    } else {
      toast.error(data?.msg || '获取用户列表失败')
    }
  } catch (e: any) {
    toast.error('网络错误，获取用户列表失败')
  } finally {
    loading.value = false
  }
}

onMounted(fetchUsers)

const onSearch = () => {
  page.value = 1
  fetchUsers()
}

// 打开新增
const openAddModal = () => {
  addForm.username = ''
  addForm.password = ''
  addForm.email = ''
  addForm.role = 'user'
  addUsernameError.value = ''
  addDialogOpen.value = true
  setTimeout(() => {
    addForm.username = ''
    addForm.password = ''
  }, 60)
}

const handleAddUser = async () => {
  if (!addForm.username.trim()) {
    toast.error('请输入用户名')
    return
  }
  if (addUsernameError.value) {
    toast.error(addUsernameError.value)
    return
  }
  if (!addForm.password.trim() || addForm.password.length < 4) {
    toast.error('密码长度至少 4 位')
    return
  }

  actionLoading.value = true
  try {
    const rsp = await request.post('/api/v1/users', addForm)
    if (rsp.data && rsp.data.code === 200) {
      toast.success('添加用户成功')
      addDialogOpen.value = false
      fetchUsers()
    } else {
      toast.error(rsp.data?.msg || '添加用户失败')
    }
  } catch (e: any) {
    toast.error('请求失败: ' + (e?.message || '未知错误'))
  } finally {
    actionLoading.value = false
  }
}

// 打开编辑
const openEditModal = (user: UserItem) => {
  editForm.id = user.id
  editForm.username = user.username
  editForm.email = user.email || ''
  editForm.role = user.role || 'user'
  editForm.status = user.status
  editForm.is_oidc = !!user.is_oidc
  editForm.auth_provider = user.auth_provider || ''
  editDialogOpen.value = true
}

const handleEditUser = async () => {
  actionLoading.value = true
  try {
    const rsp = await request.put('/api/v1/users', editForm)
    if (rsp.data && rsp.data.code === 200) {
      toast.success('修改用户信息成功')
      editDialogOpen.value = false
      fetchUsers()
    } else {
      toast.error(rsp.data?.msg || '修改失败')
    }
  } catch (e: any) {
    toast.error('请求失败: ' + (e?.message || '未知错误'))
  } finally {
    actionLoading.value = false
  }
}

const handleAdminUnbindSSO = async () => {
  if (!confirm(`确定要解除用户 [${editForm.username}] 的单点登录绑定吗？解除后该用户将转为纯本地账号。`)) {
    return
  }
  actionLoading.value = true
  try {
    const rsp = await request.post('/api/v1/users/unbind-sso', {
      id: editForm.id,
      provider: editForm.auth_provider
    })
    if (rsp.data && rsp.data.code === 200) {
      toast.success('已成功解除该用户的单点登录绑定')
      editForm.is_oidc = false
      editForm.auth_provider = ''
      fetchUsers()
    } else {
      toast.error(rsp.data?.msg || '解除绑定失败')
    }
  } catch (e: any) {
    toast.error('请求失败: ' + (e?.message || '未知错误'))
  } finally {
    actionLoading.value = false
  }
}

// 打开重置密码
const openResetPasswdModal = (user: UserItem) => {
  resetPasswdForm.id = user.id
  resetPasswdForm.username = user.username
  resetPasswdForm.password = ''
  targetResetUserIsOidc.value = !!user.is_oidc
  targetResetUserProvider.value = user.auth_provider || 'OIDC'
  resetPasswdDialogOpen.value = true
  setTimeout(() => {
    resetPasswdForm.password = ''
  }, 60)
}

const handleResetPassword = async () => {
  if (!resetPasswdForm.password.trim() || resetPasswdForm.password.length < 4) {
    toast.error('新密码长度至少 4 位')
    return
  }
  actionLoading.value = true
  try {
    const rsp = await request.post('/api/v1/users/reset-passwd', {
      id: resetPasswdForm.id,
      password: resetPasswdForm.password
    })
    if (rsp.data && rsp.data.code === 200) {
      toast.success(`已成功重置 [${resetPasswdForm.username}] 的密码`)
      resetPasswdDialogOpen.value = false
    } else {
      toast.error(rsp.data?.msg || '重置密码失败')
    }
  } catch (e: any) {
    toast.error('请求失败: ' + (e?.message || '未知错误'))
  } finally {
    actionLoading.value = false
  }
}

// 打开删除确认
const openDeleteModal = (user: UserItem) => {
  targetDeleteUser.value = user
  deleteDialogOpen.value = true
}

const handleDeleteUser = async () => {
  if (!targetDeleteUser.value) return
  actionLoading.value = true
  try {
    const rsp = await request.delete('/api/v1/users', {
      data: { id: targetDeleteUser.value.id }
    })
    if (rsp.data && rsp.data.code === 200) {
      toast.success('用户已删除')
      deleteDialogOpen.value = false
      fetchUsers()
    } else {
      toast.error(rsp.data?.msg || '删除用户失败')
    }
  } catch (e: any) {
    toast.error('请求失败: ' + (e?.message || '未知错误'))
  } finally {
    actionLoading.value = false
  }
}
</script>

<template>
  <Card class="border-border shadow-sm">
    <CardHeader class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4">
      <div>
        <CardTitle class="text-xl font-bold flex items-center gap-2">
          <span>用户管理</span>
          <Badge variant="outline" class="text-xs font-mono font-normal">共 {{ total }} 位用户</Badge>
        </CardTitle>
        <CardDescription class="mt-1">
          管理系统成员账号、角色权限与登录状态。系统数据按用户严格隔离。
        </CardDescription>
      </div>

      <div class="flex items-center gap-2">
        <Button variant="outline" size="sm" @click="fetchUsers" :disabled="loading" class="gap-1.5">
          <RefreshCwIcon class="w-4 h-4" :class="{ 'animate-spin': loading }" />
          <span>刷新</span>
        </Button>
        <Button size="sm" @click="openAddModal" class="gap-1.5">
          <PlusIcon class="w-4 h-4" />
          <span>新增用户</span>
        </Button>
      </div>
    </CardHeader>

    <CardContent class="space-y-4">
      <!-- 搜索过滤条 -->
      <div class="flex items-center gap-3">
        <div class="relative flex-1 max-w-sm">
          <SearchIcon class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            v-model="search"
            placeholder="按用户名或邮箱搜索..."
            class="pl-9 h-9"
            @keyup.enter="onSearch"
          />
        </div>
        <Button variant="secondary" size="sm" class="h-9" @click="onSearch">
          查询
        </Button>
      </div>

      <!-- 用户列表表格 -->
      <div class="overflow-x-auto rounded-lg border border-border">
        <table class="w-full text-sm border-collapse">
          <thead>
            <tr class="bg-muted/50 border-b border-border text-muted-foreground">
              <th class="px-4 py-3 text-left font-medium">用户</th>
              <th class="px-4 py-3 text-left font-medium">用户 ID</th>
              <th class="px-4 py-3 text-left font-medium">角色</th>
              <th class="px-4 py-3 text-left font-medium">状态</th>
              <th class="px-4 py-3 text-left font-medium">创建时间</th>
              <th class="px-4 py-3 text-right font-medium">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border">
            <tr v-if="!loading && users.length === 0">
              <td colspan="6" class="px-4 py-12 text-center text-muted-foreground">
                没有找到匹配的用户账号
              </td>
            </tr>
            <tr v-if="loading && users.length === 0">
              <td colspan="6" class="px-4 py-12 text-center text-muted-foreground">
                <RefreshCwIcon class="w-5 h-5 animate-spin mx-auto mb-2 text-primary" />
                正在加载用户列表...
              </td>
            </tr>
            <tr
              v-for="user in users"
              :key="user.id"
              class="hover:bg-muted/30 transition-colors"
            >
              <td class="px-4 py-3">
                <div class="font-medium text-foreground flex items-center gap-1.5 flex-wrap">
                  <span>{{ user.username }}</span>
                  <Badge
                    v-if="user.is_oidc || user.login_type === 'sso'"
                    variant="secondary"
                    class="text-[10px] py-0.5 px-1.5 font-normal bg-sky-500/10 text-sky-600 dark:text-sky-400 border border-sky-500/25 inline-flex items-center gap-1"
                    :title="`该账号通过 ${formatProviderName(user.auth_provider)} 单点登录集成创建`"
                  >
                    <FingerprintIcon class="w-3 h-3" />
                    <span>{{ formatSsoBadge(user.auth_provider) }}</span>
                  </Badge>
                  <Badge
                    v-if="user.id === pageState.userId"
                    variant="secondary"
                    class="text-[10px] py-0.5 px-1.5 font-normal"
                  >
                    当前账号
                  </Badge>
                </div>
                <div class="text-xs text-muted-foreground mt-0.5">{{ user.email || '未绑定邮箱' }}</div>
              </td>
              <td class="px-4 py-3 font-mono text-xs text-muted-foreground">
                {{ user.id }}
              </td>
              <td class="px-4 py-3">
                <Badge
                  :variant="user.role === 'admin' ? 'default' : 'outline'"
                  class="font-normal text-xs"
                >
                  {{ user.role === 'admin' ? '系统管理员' : '普通用户' }}
                </Badge>
              </td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex items-center gap-1 text-xs font-medium"
                  :class="user.status === 1 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-500'"
                >
                  <span class="w-1.5 h-1.5 rounded-full" :class="user.status === 1 ? 'bg-emerald-500' : 'bg-rose-500'"></span>
                  {{ user.status === 1 ? '正常' : '已禁用' }}
                </span>
              </td>
              <td class="px-4 py-3 text-xs text-muted-foreground whitespace-nowrap">
                {{ formatTime(user.created_on) }}
              </td>
              <td class="px-4 py-3 text-right">
                <div class="flex items-center justify-end gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    class="h-8 px-2 text-xs"
                    title="修改信息"
                    @click="openEditModal(user)"
                  >
                    <PencilIcon class="w-3.5 h-3.5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    class="h-8 px-2 text-xs"
                    title="重置密码"
                    @click="openResetPasswdModal(user)"
                  >
                    <KeyRoundIcon class="w-3.5 h-3.5" />
                  </Button>
                  <Button
                    v-if="user.username !== 'admin' && user.id !== pageState.userId"
                    variant="ghost"
                    size="sm"
                    class="h-8 px-2 text-xs text-rose-500 hover:text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/30"
                    title="删除用户"
                    @click="openDeleteModal(user)"
                  >
                    <Trash2Icon class="w-3.5 h-3.5" />
                  </Button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </CardContent>
  </Card>

  <!-- 新增用户 Dialog -->
  <Dialog :open="addDialogOpen" @update:open="val => (addDialogOpen = val)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>新增系统用户</DialogTitle>
        <DialogDescription>为系统分配新成员账号，用户拥有独立的信道与任务隔离空间。</DialogDescription>
      </DialogHeader>

      <!-- 隐形诱饵输入框，用于拦截并吸收浏览器针对当前登录账号的强行回填 -->
      <div style="position: absolute; width: 0; height: 0; overflow: hidden; opacity: 0; pointer-events: none;">
        <input type="text" name="fake_username_honeypot" tabindex="-1" autocomplete="off" />
        <input type="password" name="fake_password_honeypot" tabindex="-1" autocomplete="new-password" />
      </div>

      <div class="space-y-4 py-2">
        <div class="space-y-1.5">
          <Label for="add-user-account">用户名</Label>
          <Input
            id="add-user-account"
            name="new_member_account_prevent_autofill"
            v-model="addForm.username"
            placeholder="请输入用户名"
            autocomplete="off"
            autocapitalize="off"
            spellcheck="false"
            :class="{ 'border-destructive focus-visible:ring-destructive': !!addUsernameError }"
            @blur="checkAddUsername"
            @input="addUsernameError = ''"
          />
          <p v-if="addUsernameError" class="text-xs text-destructive mt-0.5">{{ addUsernameError }}</p>
        </div>
        <div class="space-y-1.5">
          <Label for="add-user-pwd">初始登录密码</Label>
          <Input
            id="add-user-pwd"
            name="new_member_pwd_prevent_autofill"
            type="password"
            v-model="addForm.password"
            placeholder="至少 4 位字符"
            autocomplete="new-password"
          />
        </div>
        <div class="space-y-1.5">
          <Label for="email">电子邮箱（选填）</Label>
          <Input
            id="email"
            name="new_member_email_prevent_autofill"
            type="email"
            v-model="addForm.email"
            placeholder="user@example.com"
            autocomplete="off"
          />
        </div>
        <div class="space-y-1.5">
          <Label for="role">用户角色</Label>
          <select
            id="role"
            v-model="addForm.role"
            class="w-full h-9 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring"
          >
            <option value="user">普通用户 (数据与统计严格隔离)</option>
            <option value="admin">系统管理员 (可查看全局与配置)</option>
          </select>
        </div>
      </div>
      <DialogFooter>
        <Button variant="outline" @click="addDialogOpen = false" :disabled="actionLoading">取消</Button>
        <Button @click="handleAddUser" :disabled="actionLoading">
          {{ actionLoading ? '创建中...' : '确认创建' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <!-- 编辑用户 Dialog -->
  <Dialog :open="editDialogOpen" @update:open="val => (editDialogOpen = val)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>编辑用户 - {{ editForm.username }}</DialogTitle>
        <DialogDescription>修改该账号的角色及可用状态。</DialogDescription>
      </DialogHeader>
      <div class="space-y-4 py-2">
        <div v-if="editForm.is_oidc" class="p-3 rounded-lg bg-sky-500/10 border border-sky-500/20 text-xs flex items-center justify-between gap-2">
          <div class="text-sky-700 dark:text-sky-300">
            该账号已绑定 {{ formatProviderName(editForm.auth_provider) }} 单点登录。
          </div>
          <Button
            type="button"
            variant="outline"
            size="sm"
            class="h-7 px-2 text-xs text-rose-600 hover:text-rose-700 hover:bg-rose-50 dark:hover:bg-rose-950/30 border-rose-200 dark:border-rose-900/40 shrink-0"
            @click="handleAdminUnbindSSO"
            :disabled="actionLoading"
          >
            解绑 SSO
          </Button>
        </div>
        <div class="space-y-1.5">
          <Label for="edit-email">电子邮箱</Label>
          <Input id="edit-email" type="email" v-model="editForm.email" placeholder="user@example.com" />
        </div>
        <div class="space-y-1.5">
          <Label for="edit-role">用户角色</Label>
          <select
            id="edit-role"
            v-model="editForm.role"
            :disabled="editForm.username === 'admin'"
            class="w-full h-9 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring disabled:opacity-60"
          >
            <option value="user">普通用户</option>
            <option value="admin">系统管理员</option>
          </select>
          <p v-if="editForm.username === 'admin'" class="text-[11px] text-muted-foreground">内置 admin 账号必须保持系统管理员角色</p>
        </div>
        <div class="space-y-1.5">
          <Label for="edit-status">账号状态</Label>
          <select
            id="edit-status"
            v-model.number="editForm.status"
            :disabled="editForm.id === pageState.userId"
            class="w-full h-9 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring disabled:opacity-60"
          >
            <option :value="1">正常 (允许登录使用)</option>
            <option :value="0">已禁用 (禁止登录)</option>
          </select>
          <p v-if="editForm.id === pageState.userId" class="text-[11px] text-muted-foreground">无法禁用当前登录的账号自身</p>
        </div>
      </div>
      <DialogFooter>
        <Button variant="outline" @click="editDialogOpen = false" :disabled="actionLoading">取消</Button>
        <Button @click="handleEditUser" :disabled="actionLoading">
          {{ actionLoading ? '保存中...' : '保存修改' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <!-- 重置密码 Dialog -->
  <Dialog :open="resetPasswdDialogOpen" @update:open="val => (resetPasswdDialogOpen = val)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>重置密码 - {{ resetPasswdForm.username }}</DialogTitle>
        <DialogDescription>管理员可直接为用户指定新密码。</DialogDescription>
      </DialogHeader>

      <!-- 隐形诱饵输入框拦截重置密码回填 -->
      <div style="position: absolute; width: 0; height: 0; overflow: hidden; opacity: 0; pointer-events: none;">
        <input type="text" name="fake_reset_user_honeypot" tabindex="-1" autocomplete="off" />
        <input type="password" name="fake_reset_pwd_honeypot" tabindex="-1" autocomplete="new-password" />
      </div>

      <div class="space-y-4 py-2">
        <div v-if="targetResetUserIsOidc" class="p-2.5 rounded-md bg-amber-500/10 border border-amber-500/20 text-xs text-amber-700 dark:text-amber-400 space-y-1">
          <div class="font-medium">⚠️ {{ formatProviderName(targetResetUserProvider) }} 单点登录账号提示</div>
          <div>该用户通过 {{ formatProviderName(targetResetUserProvider) }} 单点登录集成，通常由外部身份源验证身份。在此设置密码将作为系统的本地兜底登录凭据。</div>
        </div>
        <div class="space-y-1.5">
          <Label for="reset-target-password">新密码</Label>
          <Input
            id="reset-target-password"
            name="reset_member_pwd_prevent_autofill"
            type="password"
            v-model="resetPasswdForm.password"
            placeholder="请输入新密码（至少 4 位）"
            autocomplete="new-password"
          />
        </div>
      </div>
      <DialogFooter>
        <Button variant="outline" @click="resetPasswdDialogOpen = false" :disabled="actionLoading">取消</Button>
        <Button @click="handleResetPassword" :disabled="actionLoading">
          {{ actionLoading ? '重置中...' : '确认重置' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <!-- 删除确认 Dialog -->
  <Dialog :open="deleteDialogOpen" @update:open="val => (deleteDialogOpen = val)">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2 text-rose-600">
          <ShieldAlertIcon class="w-5 h-5" />
          <span>确认删除用户账号？</span>
        </DialogTitle>
        <DialogDescription class="pt-2">
          此操作将彻底删除用户 <span class="font-bold text-foreground">[{{ targetDeleteUser?.username }}]</span>（ID: {{ targetDeleteUser?.id }}）。
          该操作不可逆，请谨慎确认。
        </DialogDescription>
      </DialogHeader>
      <DialogFooter class="mt-4">
        <Button variant="outline" @click="deleteDialogOpen = false" :disabled="actionLoading">取消</Button>
        <Button variant="destructive" @click="handleDeleteUser" :disabled="actionLoading">
          {{ actionLoading ? '删除中...' : '确认删除' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
:deep(input:-webkit-autofill),
:deep(input:-webkit-autofill:hover),
:deep(input:-webkit-autofill:focus) {
  -webkit-box-shadow: 0 0 0 1000px hsl(var(--background)) inset !important;
  -webkit-text-fill-color: hsl(var(--foreground)) !important;
  transition: background-color 5000s ease-in-out 0s;
}
</style>
