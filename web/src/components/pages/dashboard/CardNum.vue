<template>
  <Card class="w-full cursor-pointer hover:shadow-md transition-shadow duration-200" @click="handleClick">
    <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
      <CardTitle class="text-sm font-medium text-muted-foreground">
        {{ title }}
      </CardTitle>
      <component :is="icon" class="h-5 w-5 text-muted-foreground" />
    </CardHeader>
    <CardContent>
      <div class="text-2xl font-bold">{{ value }}</div>
      <p class="text-xs text-muted-foreground">{{ description }}</p>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card"
import type { Component } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const props = defineProps<{
  title: string
  value: string | number
  description?: string
  icon?: Component
  routePath?: string
}>()

const handleClick = () => {
  if (props.routePath) {
    router.push(props.routePath)
  }
}
</script>

<script lang="ts">
export default {
  name: 'CardNum'
}
</script>