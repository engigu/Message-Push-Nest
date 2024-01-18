
<template>
    <el-popconfirm width="70px" confirm-button-text="确定" cancel-button-text="算了" icon-color="#626AEF" title="确定删除吗?"
        @confirm="handleDelete()">
        <template #reference>
            <el-button link size="small" type="danger">删除</el-button>
        </template>
    </el-popconfirm>
</template>
  
<script>
import { defineComponent } from 'vue';

export default defineComponent({
    methods: {
        handleDelete() {
            this.$emit('customHandleDelete');
        }
    }
});
</script>
  