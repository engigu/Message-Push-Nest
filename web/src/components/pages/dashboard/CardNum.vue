<template>
  <Card class="w-full cursor-pointer hover:shadow-lg hover:-translate-y-1 transition-all duration-300 group animate-fade-in" @click="handleClick">
    <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
      <CardTitle class="text-sm font-medium text-muted-foreground group-hover:text-primary transition-colors duration-300">
        {{ title }}
      </CardTitle>
      <div class="p-1.5 rounded-lg bg-secondary/30 group-hover:bg-primary/10 transition-colors duration-300">
        <component :is="icon" class="h-5 w-5 text-muted-foreground group-hover:text-primary transition-all duration-300 icon-animate" />
      </div>
    </CardHeader>
    <CardContent>
      <div class="text-2xl font-bold tracking-tight">{{ value }}</div>
      <p v-if="description" class="text-xs text-muted-foreground mt-1">{{ description }}</p>
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

<style scoped>
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fade-in {
  animation: fadeInUp 0.4s ease-out both;
}

@keyframes icon-bounce {
  0%, 100% {
    transform: translateY(0) scale(1.1) rotate(6deg);
  }
  50% {
    transform: translateY(-3px) scale(1.1) rotate(6deg);
  }
}

.group:hover .icon-animate {
  animation: icon-bounce 1s ease-in-out infinite;
}
</style>