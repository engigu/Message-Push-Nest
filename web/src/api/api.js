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
        if (!config.url.startsWith('/api/')) {
            config.url = '/api/v1' + config.url;
        }
        const token = pageState.Token || localStorage.getItem(CONSTANT.STORE_TOKEN_NAME);
        // 统一基于领域命名空间判断公开接口：凡 /auth/ 或 /public/ 下的请求均无需携带 Token
        const isPublic = config.url.startsWith('/api/v1/auth/') || config.url.startsWith('/api/v1/public/');
        if (token && !isPublic) {
            config.headers = {
                ...config.headers,
                'm-token': token,
            };
            if (!pageState.Token) {
                pageState.setToken(token);
            }
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
                logout('登录已失效，请重新登录');
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
            logout('登录凭据已失效或后端已重启，请重新登录');
            return Promise.reject(error);
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
const logout = (reason = '') => {
    const pageState = usePageState();
    
    // 清除token和登录状态
    pageState.setToken('');
    localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME);
    
    if (reason) {
        toast.error(reason);
    }
    // 立即跳转到登录页
    router.push('/login');
};

export { request, handleException, logout };
