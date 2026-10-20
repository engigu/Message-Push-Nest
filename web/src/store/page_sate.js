
import { defineStore } from 'pinia';
import { LocalStieConfigUtils } from '@/util/localSiteConfig'
import { CONSTANT } from '../constant'


function parseJwt(token) {
    if (!token) return null;
    try {
        const base64Url = token.split('.')[1];
        if (!base64Url) return null;
        const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
        const jsonPayload = decodeURIComponent(
            atob(base64)
                .split('')
                .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
                .join('')
        );
        return JSON.parse(jsonPayload);
    } catch (e) {
        return null;
    }
}

export const usePageState = defineStore('pageState', {
    state: () => {
        const token = localStorage.getItem(CONSTANT.STORE_TOKEN_NAME) || '';
        const payload = parseJwt(token);
        return {
            isLogin: Boolean(token && token.trim() !== ''), // 全局的登录状态
            Token: token, // 全局的登录状态
            userId: payload?.user_id || '',
            userName: payload?.username || '',
            userRole: payload?.role || 'user',
            isShowAddWayDialog: false,
            siteConfigData: LocalStieConfigUtils.getLocalConfig(),
            ShowDialogData: {}
        }
    },
    getters: {
        isAdmin: (state) => state.userRole === 'admin',
    },
    actions: {
        setIsLogin(state) {
            this.isLogin = state;
        },
        setToken(token) {
            this.Token = token;
            if (token && token.trim() !== '') {
                localStorage.setItem(CONSTANT.STORE_TOKEN_NAME, token);
                this.isLogin = true;
                const payload = parseJwt(token);
                this.userId = payload?.user_id || '';
                this.userName = payload?.username || '';
                this.userRole = payload?.role || 'user';
            } else {
                localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME);
                this.isLogin = false;
                this.userId = '';
                this.userName = '';
                this.userRole = 'user';
            }
        },
        setShowAddWayDialog(status) {
            this.isShowAddWayDialog = status;
        },
        setSiteConfigData(configData) {
            this.siteConfigData = configData;
        },
    },
});
