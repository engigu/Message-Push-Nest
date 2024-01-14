// request.js
import axios from 'axios';
import { ElMessage } from 'element-plus'
import { usePageState } from '../store/page_sate';
import { CONSTANT } from '../constant';
import config from '../../config.js';

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
            ElMessage({ message: response.data.msg, type: 'error' })
            // Promise.reject();
        }
        return response;
    },
    (error) => {
        if (error.response && error.response.status == 401) {
            logout();
        } else if (20000 <= error.response.status && error.response.status <= 29999) {
            logout();
        } else {
            handleException(error);
        }
    }
);

// 异常处理
const handleException = (error) => {
    if (error.code == ERR_NETWORK) {
        ElMessage({ message: `网络错误！`, type: 'error' })
    } else {
        let msg = `未知错误：${error.response.status}, ${error.response.data.msg}`;
        ElMessage({ message: msg, type: 'error' })
    };

};

// 登出系统
const logout = () => {
    const pageState = usePageState();
    pageState.setIsLogin(true);
    localStorage.setItem(CONSTANT.STORE_TOKEN_NAME, "");
    setTimeout(() => {
        window.location.href = '/';
    }, 500);
};

export { request, handleException, logout };
