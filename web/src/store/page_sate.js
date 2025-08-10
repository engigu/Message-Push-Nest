
import { defineStore } from 'pinia';
import { LocalStieConfigUtils } from '@/util/localSiteConfig'
import { CONSTANT } from '../constant'


export const usePageState = defineStore('pageState', {
    // id: 'pageState',
    state: () => {
        const token = localStorage.getItem(CONSTANT.STORE_TOKEN_NAME) || '';
        return {
            isLogin: Boolean(token && token.trim() !== ''), // 全局的登录状态
            Token: token, // 全局的登录状态
            isShowAddWayDialog: false,
            siteConfigData: LocalStieConfigUtils.getLocalConfig(),
            ShowDialogData: {}
        }
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
            } else {
                localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME);
                this.isLogin = false;
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
