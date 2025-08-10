<script setup lang="ts">
import { ref } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from 'vue-sonner'
import { request } from '@/api/api'

// 重置密码相关
const passwordForm = ref({
  newPassword: '',
  currentPassword: ''
})

// 重置密码
const resetPassword = async () => {

  
  // 检查演示模式
    // @ts-ignore
  const isDemoMode = (import.meta as any).env.VITE_RUN_MODE === 'demo'
  
  if (isDemoMode) {
    toast.error('演示模式下无法重置密码')
    return
  }

  try {
    let postData = { new_passwd: passwordForm.value.newPassword , old_passwd: passwordForm.value.currentPassword}
    const rsp = await request.post('/settings/setpasswd', postData)
    if (rsp.data.code == 200) {
      let msg = rsp.data.msg
      toast.success(msg)
    } else {
      toast.error(rsp.data.msg || '密码重置失败')
    }
  } catch (error) {
    toast.error('密码重置失败，请稍后重试')
  }
}
</script>

<script lang="ts">
export default {
  name: 'PasswordSettings'
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>重置密码</CardTitle>
      <CardDescription>更改您的登录密码</CardDescription>
    </CardHeader>
    <CardContent class="space-y-4">

      <div class="space-y-2">
        <Label for="current-password">旧密码</Label>
        <Input
          id="current-password"
          type="password"
          v-model="passwordForm.currentPassword"
          placeholder="请输入旧密码"
        />
      </div>  
        <div class="space-y-2">
        <Label for="new-password">新密码</Label>
        <Input
          id="new-password"
          type="password"
          v-model="passwordForm.newPassword"
          placeholder="请输入新密码"
        />
      </div>
 
      <Button @click="resetPassword" class="w-full sm:w-auto">
        重置密码
      </Button>
    </CardContent>
  </Card>
</template>