// request.js
import axios from 'axios';
import router from '../router';

import { usePageState } from '../store/page_sate';
import { CONSTANT } from '../constant';
import config from '../../config.js';
import { toast } from "vue-sonner"


const ERR_NETWORK = "ERR_NETWORK";

// 获取路径前缀
const getPathPrefix = () => {
    return config.pathPrefix || '';
};

const request = axios.create({
    baseURL: config.apiUrl + getPathPrefix(),
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
        // 检查业务逻辑错误
        if (response && response.data.code != 200) {
            // 检查是否是token相关的错误码
            const tokenErrorCodes = [20001, 20002, 20003, 20004, 20005];
            if (tokenErrorCodes.includes(response.data.code)) {
                // Token失效，执行登出
                logout();
                return Promise.reject(response);
            }
            
            // 其他业务错误显示toast
            toast.error(response.data.msg, {
                description: '接口逻辑错误'
            });
        }
        return response;
    },
    (error) => {
        // HTTP状态码401表示未授权
        if (error.response && error.response.status == 401) {
            // 检查响应体中的业务错误码
            if (error.response.data && error.response.data.code) {
                const tokenErrorCodes = [20001, 20002, 20003, 20004, 20005];
                if (tokenErrorCodes.includes(error.response.data.code)) {
                    logout();
                    return Promise.reject(error);
                }
            }
            
            // 其他401错误也执行登出
            logout();
        } else {
            handleException(error);
        }
        return Promise.reject(error);
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
    
    // 清除token和登录状态
    pageState.setToken('');
    localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME);
    
    // 立即跳转到登录页
    router.push('/login');
};

export { request, handleException, logout };
