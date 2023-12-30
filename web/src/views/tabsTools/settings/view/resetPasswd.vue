<template>
    <div class="setting-container">
        <el-input v-model="oldPasswd" size="small" placeholder="请输入旧密码" type="password" show-password>
            <template #prepend>旧密码</template>
        </el-input>
        <el-input v-model="newPasswd" size="small" placeholder="请输入新密码" type="password" show-password
            style=" margin-top: 15px;">
            <template #prepend>新密码</template>
        </el-input>
        <el-button @click="handleChange" size="small" type="primary">确定</el-button>
    </div>
</template>
  
<script>
import { defineComponent, reactive, toRefs } from 'vue';
import { ElMessage } from 'element-plus'
import { request } from '@/api/api'

export default defineComponent({
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
/* :deep(.el-input .el-input__wrapper) {
    margin-top: 10px;
} */

:deep(.el-button) {
    float: right !important;
    margin-top: 10px;
}

.setting-container {
    width: 300px;
    margin: 50px auto;
}
</style>