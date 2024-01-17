
import { defineStore } from 'pinia';
import { LocalStieConfigUtils } from '@/util/localSiteConfig'


export const usePageState = defineStore({
    id: 'pageState',
    state: () => ({
        isLogin: false, // 全局的登录状态
        Token: '', // 全局的登录状态
        isShowAddWayDialog: false,
        siteConfigData: LocalStieConfigUtils.getLocalConfig(),
        ShowDialogData: {}
    }),
    actions: {
        setIsLogin(state) {
            this.isLogin = state;
        },
        setToken(token) {
            this.Token = token;
        },
        setShowAddWayDialog(status) {
            this.isShowAddWayDialog = status;
        },
    },
});
