<script lang="ts">
export const description = "A login page with form and image."
</script>

<script setup lang="ts">
import { Toaster } from '@/components/ui/sonner'
import 'vue-sonner/style.css' 
// 不再直接导入Index组件，而是使用路由视图
</script>

<template>
    <!-- shacn-vue的toast -->
    <Toaster />
    <!-- 使用路由视图代替直接渲染Index组件 -->
    <router-view />
</template>
