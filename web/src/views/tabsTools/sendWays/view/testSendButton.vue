
<template>
    <el-popconfirm width="400px" confirm-button-text="确定" cancel-button-text="算了" icon-color="#626AEF"
        title="将往添加的渠道发送一条测试消息，请注意查收!" @confirm="handleSubmit()">
        <template #reference>
            <el-button size="small" type="primary">点击测试</el-button>
        </template>
    </el-popconfirm>
</template>
  
<script>
import { defineComponent } from 'vue';

export default defineComponent({
    methods: {
        handleSubmit() {
            this.$emit('customhandleSubmit');
        }
    }
});
</script>
  