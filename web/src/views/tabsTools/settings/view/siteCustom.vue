<template>
    <div class="setting-container">
        <div>
            <el-input v-model="title" size="small" placeholder="请输入自定义的网站标题">
                <template size="small" #prepend>站点标题</template>
            </el-input>
            <el-input v-model.number="slogan" size="small" placeholder="请输入自定义的网站slogan" style="margin-top: 15px;">
                <template size="small" #prepend>站点标语</template>
            </el-input>
            <el-input v-model.number="logo" size="small" placeholder="请输入自定义的网站logo（svg文本）" style="margin-top: 15px;">
                <template size="small" #prepend>站点图标</template>
            </el-input>
        </div>
        <div class="buttom">
            <div class="tips">
                <el-text size="small">说明</el-text>
                <el-tooltip placement="top">
                    <template #content>
                        1. logo请输入svg文本，替换后登录页面，ico，导航栏logo将全部一起更换
                        <br />
                        2. slogan将在登录页面展示
                        <br />
                        *将在下一次登录的时候生效
                    </template>
                    <el-icon>
                        <QuestionFilled />
                    </el-icon>
                </el-tooltip>
            </div>
            <el-button @click="handleView" size="small">恢复默认</el-button>
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
    props: {
    },
    methods: {
    },
    setup() {
        const router = useRouter();
        const state = reactive({
            title: '',
            slogan: '',
            logo: '',
            section: 'site_config',
        });

        const handleSubmit = async () => {
            let postData = {
                section: state.section,
                data: {
                    title: state.title.trim(),
                    slogan: state.slogan.trim(),
                    logo: state.logo.trim(),
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
            let params = { params: { section: "site_config" } };
            const rsp = await request.get('/settings/getsetting', params);
            if (await rsp.data.code == 200) {
                let data = await rsp.data.data;
                state.title = data.title;
                state.logo = data.logo;
                state.slogan = data.slogan;
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