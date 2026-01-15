// config.js

const isProduction = process.env.NODE_ENV === 'prod';

// 从 window 对象获取路径前缀（由后端注入或通过 API 获取）
const getPathPrefix = () => {
    return window.__URL_PATH_PREFIX__ || '';
};

const config = {
    apiUrl: isProduction ? '' : 'http://localhost:8000',
    pathPrefix: getPathPrefix(),
};

export default config;

