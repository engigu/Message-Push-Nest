// request.js
import axios from 'axios';

import { usePageState } from '../store/page_sate';
import { CONSTANT } from '../constant';
import config from '../../config.js';
import { toast } from "vue-sonner"


const ERR_NETWORK = "ERR_NETWORK";

const request = axios.create({
    baseURL: config.apiUrl,
    timeout: 50000,
    withCredentials: true,
});



// 请求拦截器
request.interceptors.request.use(
    (config) => {
        const pageState = usePageState();
        if (!CONSTANT.NO_AUTH_URL.includes(config.url)) {
            config.url = '/api/v1' + config.url;
        }
        if (pageState.Token && !CONSTANT.NO_AUTH_URL.includes(config.url)) {
            config.headers = {
                ...config.headers,
                'm-token': pageState.Token,
            };
        }
        return config;
    },
    (error) => {
        handleException(error);
    }
);

// 响应拦截器
request.interceptors.response.use(
    (response) => {

        if (response && response.data.code != 200) {
            toast.error(response.data.msg, {
                description: '接口逻辑错误'
            })
            // Promise.reject();
        }
        return response;
    },
    (error) => {

        if (error.response && error.response.status == 401) {
            logout();
        } else if (error.response && 20000 <= error.response.status && error.response.status <= 29999) {
            logout();
        } else {
            handleException(error);
        }
    }
);

// 异常处理
const handleException = (error) => {
    if (!error.response) {

        return
    };

    if (error.code == ERR_NETWORK) {
        toast(`网络错误！`)
    } else {
        let msg = `未知错误：${error.response.status}, ${error.response.data.msg}`;
        toast(msg)
    };

};

// 登出系统
const logout = () => {
    const pageState = usePageState();
    pageState.setIsLogin(false);
    localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME);
    setTimeout(() => {
        window.location.href = '/login';
    }, 500);
};

export { request, handleException, logout };
