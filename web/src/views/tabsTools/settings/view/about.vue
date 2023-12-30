<template>
    <div class="setting-container">
        <el-text size="small">当前版本：v1.0.0</el-text>
        <div class="buttom">
            <div class="tips">
                <el-text size="small">版本功能说明</el-text>
                <el-tooltip placement="top">
                    <template #content>
                        logo请输入svg文本，替换后登录页面、ico、导航栏logo将全部一起更换
                        <br />
                        *将在下一次登录的时候生效
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
import { defineComponent, reactive, toRefs } from 'vue';
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
            oldPasswd: '',
            newPasswd: '',
        });

        const handleChange = async () => {
            let postData = { old_passwd: state.oldPasswd, new_passwd: state.newPasswd };
            const rsp = await request.post('/settings/setpasswd', postData);
            if (await rsp.data.code == 200) {
                let msg = await rsp.data.msg;
                ElMessage({ message: msg, type: 'success' })
            }
        }
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