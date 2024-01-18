<template>
    <div class="setting-container">
        <div>
            <el-input v-model="cron" size="small" placeholder="请输入定时日志清除的Cron表达式">
                <template size="small" #prepend>cron://</template>
            </el-input>
            <el-input v-model="keepNum" size="small" placeholder="请输入要保留的最近的日志条数" style="margin-top: 15px;">
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
import { defineComponent, reactive, toRefs, onMounted } from 'vue';
import { ElMessage } from 'element-plus'
import { QuestionFilled } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router';
import { request } from '@/api/api'
import { CONSTANT } from '@/constant'

export default defineComponent({
    components: {
        QuestionFilled,
    },
    setup() {
        const router = useRouter();
        const state = reactive({
            section: 'log_config',
            cron: '',
            keepNum: '1000',
        });

        const handleSubmit = async () => {
            let postData = {
                section: state.section,
                data: {
                    cron: state.cron.trim(),
                    keep_num: state.keepNum.trim(),
                },
            };
            const rsp = await request.post('/settings/set', postData);
            if (await rsp.data.code == 200) {
                let msg = await rsp.data.msg;
                ElMessage({ message: msg, type: 'success' })
            }
        }

        const handleView = async () => {
            router.push('/sendlogs?taskid=' + CONSTANT.LOG_TASK_ID, { replace: true });
        }

        const getSiteConfig = async () => {
            let params = { params: { section: "log_config" } };
            const rsp = await request.get('/settings/getsetting', params);
            if (await rsp.data.code == 200) {
                let data = await rsp.data.data;
                state.cron = data.cron;
                state.keepNum = data.keep_num;
            }
        }

        onMounted(() => {
            getSiteConfig();
        })

        return {
            ...toRefs(state), handleSubmit, handleView
        }
    }

});
</script>
  
<style scoped>

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