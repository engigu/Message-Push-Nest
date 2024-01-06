<template>
    <div class="setting-container">
        <el-text size="small">当前版本：{{ version }}</el-text>
        <div class="buttom">
            <div class="tips">
                <el-text size="small">版本功能说明</el-text>
                <el-tooltip v-if="desc" placement="top">
                    <template #content>
                      <div v-html="desc"></div>
                    </template>
                    <el-icon>
                        <QuestionFilled />
                    </el-icon>
                </el-tooltip>
            </div>
        </div>
    </div>
</template>
  
<script>
import { defineComponent, reactive, toRefs, onMounted } from 'vue';
import { ElMessage } from 'element-plus'
import { request } from '@/api/api'
import { QuestionFilled } from '@element-plus/icons-vue'

export default defineComponent({
    components: {
        QuestionFilled,
    },
    props: {
    },
    methods: {
    },
    setup() {
        const state = reactive({
            version: '',
            desc: '',
        });

        const handleChange = async () => {
            let postData = { old_passwd: state.oldPasswd, new_passwd: state.newPasswd };
            const rsp = await request.post('/settings/setpasswd', postData);
            if (await rsp.data.code == 200) {
                let msg = await rsp.data.msg;
                ElMessage({ message: msg, type: 'success' })
            }
        }

        const getAbout = async () => {
            let params = { params: { section: "about" } };
            const rsp = await request.get('/settings/getsetting', params);
            if (await rsp.data.code == 200) {
                let data = await rsp.data.data;
                state.version = data.version;
                state.desc = data.desc.replace(/\n/g, '<br />');
            }
        }

        onMounted(() => {
            getAbout();
        })

        return {
            ...toRefs(state), handleChange
        }
    }

});
</script>
  
<style scoped>
:deep(.el-input .el-input__wrapper) {
    margin-top: 10px;
}

:deep(.el-button) {
    float: right !important;
    margin-top: 10px;
}

.setting-container {
    width: 200px;
    margin: 50px auto;
}

.buttom {
    margin-top: 30px;
}
</style>