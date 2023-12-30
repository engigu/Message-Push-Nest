<template>
    <div class="login-center-container" v-if="!pageState.isLogin">

        <div class="main-center-body">
            <div class="container">
                <img class="login-logo" src="../../../public/logo.svg" alt="login logo">
                <p class="desc">A Message Way Hosted Site</p>

                <div class="login-block">
                    <p class="login-text">账号：</p>
                    <el-input style="width: 80%;" v-model="account" placeholder="请输入账号" />
                </div>

                <div class="login-block">
                    <p class="login-text">密码：</p>
                    <el-input style="width: 80%;" v-model="passwd" type="password" placeholder="请输入密码" show-password />
                </div>

                <div class="btn-area">
                    <el-button id="custom-h2d-copy-button" type="success" @click="clickLogin()">登录</el-button>
                    <el-button type="primary" @click="clickRegister()">注册</el-button>
                </div>

            </div>
        </div>
    </div>
</template>

<script>
import { toRefs, reactive, onMounted } from 'vue';
import { ElMessage } from 'element-plus'
import { request } from '../../api/api'
import { CONSTANT } from '../../constant'
import { usePageState } from '../../store/page_sate';
import {  useRouter } from 'vue-router';


export default {
    setup() {
        const router = useRouter();


        const pageState = usePageState();
        const state = reactive({
            account: 'admin',
            passwd: '123456',
        });

        onMounted(() => {

        });

        // 登录
        const clickLogin = async () => {
            const rspe = await request.post('/auth', { username: state.account, passwd: state.passwd });
            const rsp = rspe.data;
            if (rsp.code != 200) {
                ElMessage({ message: rsp.msg, type: 'error' });
            } else {
                pageState.setToken(rsp.data.token);
                pageState.setIsLogin(true);
                localStorage.setItem(CONSTANT.STORE_TOKEN_NAME, rsp.data.token);
                router.push('/sendlogs', { replace: true })
            }
        };

        // 注册
        const clickRegister = () => {
            ElMessage({ message: `暂未开放注册！`, type: 'error' })

        };

        return { ...toRefs(state), clickLogin, clickRegister, pageState };
    }
}
</script>


<style scoped>
@import url('../../../src/assets/center_button_textarea.css');

.login-logo {
    height: 200px !important;
}

.login-center-container {
    text-align: center;
}

.login-center-container img {
    margin: 0 auto;
    display: block;
    width: 300px;
    height: 300px;
}

.desc {
    margin: 0 auto;
    display: block;
    font-size: 25px;
    margin-bottom: 80px;
    /* color: rgb(64, 87, 45); */
}

.login-text {
    font-size: 13px;
    /* width: 20px; */
    width: 20%;
    display: inline;
    /* display: flex; */
    /* justify-content: right;  */
    /* align-items: center; */
}

.login-block {
    margin-top: 20px;
}

.btn-area {
    margin-top: 30px;
}
</style>

