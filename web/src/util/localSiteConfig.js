import { CONSTANT } from '@/constant'
import { request } from '@/api/api'


class LocalStieConfigUtils {


    // 更新本地的site设置
    static updateLocalConfig = (data) => {
        localStorage.setItem(CONSTANT.STORE_CUSTOM_NAME, JSON.stringify(data));
    }

    // 更新本地的site设置
    static getLocalConfig = () => {
        let d = localStorage.getItem(CONSTANT.STORE_CUSTOM_NAME);
        if (!d) {
            d = CONSTANT.DEFALUT_SITE_CONFIG;
        }
        let data = JSON.parse(d);
        data.pagesize = Number(data.pagesize);
        return data
    }

    // 获取最新的本地的site设置
    static getLatestLocalConfig = async () => {
        let params = { params: { section: "site_config" } };
        const rsp = await request.get('/settings/getsetting', params);
        if (await rsp.data.code == 200) {
            let data = await rsp.data.data;
            LocalStieConfigUtils.updateLocalConfig(data);
        }
        return LocalStieConfigUtils.getLocalConfig();
    }

}

export { LocalStieConfigUtils };
