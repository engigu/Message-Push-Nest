<template>
    <div class="setting-container">
        <div>
            <el-input v-model="title" size="small" placeholder="请输入自定义的网站标题">
                <template size="small" #prepend>站点标题</template>
            </el-input>
            <el-input v-model="slogan" size="small" placeholder="请输入自定义的网站slogan">
                <template size="small" #prepend>站点标语</template>
            </el-input>
            <el-input v-model="logo" size="small" placeholder="请输入自定义的网站logo（svg文本）">
                <template size="small" #prepend>站点图标</template>
            </el-input>
            <el-input v-model="pagesize" size="small" placeholder="页面分页大小">
                <template size="small" #prepend>分页大小</template>
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
                        ** 将在下一次登录的时候生效，如果不生效请在登录页面Ctrl+F5强制刷新   
                        <br />
                        ** logo将替换网页ico，登录页面logo，导航栏logo
                    </template>
                    <el-icon>
                        <QuestionFilled />
                    </el-icon>
                </el-tooltip>
            </div>
            <el-button @click="handleSubmitReset" size="small">恢复默认</el-button>
            <el-button @click="handleSubmit" size="small" type="primary">确定</el-button>
        </div>
    </div>
</template>
  
<script>
import { defineComponent, reactive, toRefs, onMounted } from 'vue';
import { ElMessage } from 'element-plus'
import { QuestionFilled } from '@element-plus/icons-vue'
import { request } from '@/api/api'
import { LocalStieConfigUtils } from '@/util/localSiteConfig'


export default defineComponent({
    components: {
        QuestionFilled,
    },
    setup() {
        const state = reactive({
            title: '',
            slogan: '',
            logo: '',
            pagesize: '',
            section: 'site_config',
        });

        const handleSubmit = async () => {
            let postData = {
                section: state.section,
                data: {
                    title: state.title.trim(),
                    slogan: state.slogan.trim(),
                    logo: state.logo.trim(),
                    pagesize: state.pagesize,
                },
            };
            const rsp = await request.post('/settings/set', postData);
            if (await rsp.data.code == 200) {
                let msg = await rsp.data.msg;
                ElMessage({ message: msg, type: 'success' })
            }
        }

        const handleSubmitReset = async () => {
            const rsp = await request.post('/settings/reset', {});
            if (await rsp.data.code == 200) {
                let msg = await rsp.data.msg;
                ElMessage({ message: msg, type: 'success' });
                // 重新获取设置
                await getSiteConfig();
            }
        }

        const getSiteConfig = async () => {
            let params = { params: { section: "site_config" } };
            const rsp = await request.get('/settings/getsetting', params);
            if (await rsp.data.code == 200) {
                let data = await rsp.data.data;
                state.title = data.title;
                state.logo = data.logo;
                state.slogan = data.slogan;
                state.pagesize = data.pagesize;

                LocalStieConfigUtils.updateLocalConfig(data);
            }
        }

        onMounted(() => {
            getSiteConfig();
        })

        return {
            ...toRefs(state), handleSubmit, handleSubmitReset
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