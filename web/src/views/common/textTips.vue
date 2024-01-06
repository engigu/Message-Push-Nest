<template>
    <div class="tips">
        <el-text size="small">{{ text }}</el-text>
        <el-tooltip v-if="desc" placement="top">
            <template #content>
                <div v-html="desc"></div>
            </template>
            <el-icon>
                <QuestionFilled />
            </el-icon>
        </el-tooltip>
    </div>
</template>
  
<script>
import { defineComponent } from 'vue';
import { QuestionFilled } from '@element-plus/icons-vue'

export default defineComponent({
    components: {
        QuestionFilled,
    },
    props: {
        text: String,
        desc: String,
    },
    methods: {
    },
});
</script>
  
<style scoped></style>