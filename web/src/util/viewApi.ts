import { CodeTemplates } from './code-templates';

const gethttpOrigin = () => {
    return window.location.origin
}

// ==================== 公共加密工具 ====================

class TokenEncryption {
    // 根据字符串内容生成确定性 salt（范围 0~255）
    static getDeterministicSalt(text: string) {
        let sum = 0;
        for (let i = 0; i < text.length; i++) {
            sum = (sum + text.charCodeAt(i) * (i + 1)) & 0xFF;
        }
        return sum;
    }

    // 加密：首字节为salt，后续为按位异或后的数据
    static encryptHex(text: string, key: number) {
        const salt = TokenEncryption.getDeterministicSalt(text);
        let result = salt.toString(16).padStart(2, '0');
        for (let i = 0; i < text.length; i++) {
            const code = text.charCodeAt(i) ^ (key & 0xFF) ^ ((salt + i) & 0xFF);
            result += code.toString(16).padStart(2, '0');
        }
        return result;
    }
}

// ==================== 公共代码模板生成器 (已移至分离的模块) ====================

// ==================== 发信任务 API (V1) ====================

export class ApiStrGenerate {
    static getDataString(task_id: string, options: any) {
        let data: any = { token: TokenEncryption.encryptHex(task_id, 71) };
        data.title = 'message title';
        data.text = 'Hello World!';
        if (options.html) data.html = '<h1> Hello World! </h1>';
        if (options.markdown) data.markdown = '**Hello World!**';
        if (options.url) data.url = 'https://github.com';
        if (options.at_mobiles) data.at_mobiles = ['13800138000', '13900139000'];
        if (options.at_user_ids) data.at_user_ids = ['zhangsan', 'lisi'];
        if (options.at_all) data.at_all = true;
        // 动态接收者字段（用于邮箱、微信公众号等支持动态接收者的渠道）
        if (options.recipients) data.recipients = ['user1@example.com', 'user2@example.com'];
        return JSON.stringify(data, null, 4);
    }

    static getApiUrl() {
        return `${gethttpOrigin()}/api/v1/message/send`;
    }

    static getCurlString(task_id: string, options: any, isFunction: boolean = false) {
        return CodeTemplates.getCurl(this.getApiUrl(), this.getDataString(task_id, options), isFunction);
    }

    static getGolangString(task_id: string, options: any, isFunction: boolean = false) {
        return CodeTemplates.getGolang(this.getApiUrl(), this.getDataString(task_id, options), isFunction);
    }

    static getPythonString(task_id: string, options: any, isFunction: boolean = false) {
        return CodeTemplates.getPython(this.getApiUrl(), this.getDataString(task_id, options), isFunction);
    }

    static getJaveString(task_id: string, options: any, isFunction: boolean = false) {
        return CodeTemplates.getJava(this.getApiUrl(), this.getDataString(task_id, options), isFunction);
    }

    static getRustString(task_id: string, options: any, isFunction: boolean = false) {
        return CodeTemplates.getRust(this.getApiUrl(), this.getDataString(task_id, options), isFunction);
    }

    static getPHPString(task_id: string, options: any, isFunction: boolean = false) {
        return CodeTemplates.getPHP(this.getApiUrl(), this.getDataString(task_id, options), isFunction);
    }

    static getNodeString(task_id: string, options: any, isFunction: boolean = false) {
        return CodeTemplates.getNode(this.getApiUrl(), this.getDataString(task_id, options), isFunction);
    }
}

// ==================== 模板 API (V2) ====================

export class TemplateApiStrGenerate {
    static getTemplateDataString(template_id: string, placeholders_json: string, options: any = {}) {
        // 解析占位符配置
        let placeholders: any = {};
        try {
            const placeholdersList = JSON.parse(placeholders_json || '[]');
            // 根据占位符配置生成示例值
            placeholdersList.forEach((p: any) => {
                placeholders[p.key] = p.default || `mock_${p.key}`;
            });
        } catch (e) {
            // 如果解析失败，使用默认示例
            placeholders = {
                'username': 'John Doe',
                'email': 'john@example.com',
                'phone': '13800138000'
            };
        }

        let data: any = {
            token: TokenEncryption.encryptHex(template_id, 71),
            title: 'message title',
            placeholders: placeholders
        };

        // 添加动态接收者字段（如果需要）
        if (options.recipients) {
            data.recipients = ['user1@example.com', 'user2@example.com'];
        }

        return JSON.stringify(data, null, 4);
    }

    static getApiUrl() {
        return `${gethttpOrigin()}/api/v2/message/send`;
    }

    static getCurlString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getCurl(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }

    static getGolangString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getGolang(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }

    static getPythonString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getPython(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }

    static getJavaString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getJava(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }

    static getRustString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getRust(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }

    static getPHPString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getPHP(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }

    static getNodeString(template_id: string, placeholders_json: string, options: any = {}, isFunction: boolean = false) {
        return CodeTemplates.getNode(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json, options), isFunction);
    }
}
