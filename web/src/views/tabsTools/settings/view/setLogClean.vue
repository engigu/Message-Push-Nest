<template>
    <div class="setting-container">
        <div>
            <el-input v-model="cron" size="small" placeholder="请输入定时日志清除的Cron表达式">
                <template size="small" #prepend>cron://</template>
            </el-input>
            <el-input v-model.number="keepNum" size="small" placeholder="请输入要保留的最近的日志条数" style="margin-top: 15px;">
                <template size="small" #prepend>保留数</template>
            </el-input>
        </div>
        <div class="buttom">
            <div class="tips">
                <el-text size="small">说明</el-text>
                <el-tooltip placement="top">
                    <template #content>
                        cron如果不设置，默认是在每天的0点1分进行清理
                        <br />
                        保留数目如果不设置，默认保留最近1000条
                    </template>
                    <el-icon>
                        <QuestionFilled />
                    </el-icon>
                </el-tooltip>
            </div>
            <el-button @click="handleView" size="small">查看日志</el-button>
            <el-button @click="handleSubmit" size="small" type="primary">确定</el-button>
        </div>
    </div>
</template>
  
<script>
import { defineComponent, reactive, toRefs } from 'vue';
import { ElMessage } from 'element-plus'
import { QuestionFilled } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router';
import { request } from '@/api/api'
import { CONSTANT } from '@/constant'

export default defineComponent({
    components: {
        QuestionFilled,
    },
    props: {
    },
    methods: {
    },
    setup() {
        const router = useRouter();
        const state = reactive({
            cron: '',
            keepNum: 1000,
        });

        const handleSubmit = async () => {
            let postData = { old_passwd: state.oldPasswd, new_passwd: state.newPasswd };
            const rsp = await request.post('/settings/setpasswd', postData);
            if (await rsp.data.code == 200) {
                let msg = await rsp.data.msg;
                ElMessage({ message: msg, type: 'success' })
            }
        }
        const handleView = async () => {
            router.push('/sendlogs?taskid=' + CONSTANT.LOG_TASK_ID, { replace: true });
        }
        return {
            ...toRefs(state), handleSubmit, handleView
        }
    }

});
</script>
  
<style scoped>
/* :deep(.el-input .el-input__wrapper) {
    margin-top: 10px;
} */

:deep(.el-button) {
    float: right !important;
    margin-top: 10px;
}

.buttom {
    display: flex;
    width: 300px;
}

.tips {
    width: 300px;
    margin-top: 10px;
}

.setting-container {
    width: 300px;
    margin: 40px auto;
}
</style>