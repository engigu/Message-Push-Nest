// 定义一些常量名
const CONSTANT = {
    PAGE: 1,
    PAGE_SIZE: 6,
    TOTAL: 0,
    LOG_TASK_ID: "00000000-0000-0000-0000-000000000001",
    STORE_TOKEN_NAME: '__message_nest_token__',
    NO_AUTH_URL: [
        '/auth',
    ],
    WAYS_DATA: [
        {
            type: 'Email',
            label: '邮箱',
            inputs: [
                { subLabel: 'smtp服务地址', value: '', col: 'server' },
                { subLabel: 'smtp服务端口', value: '', col: 'port' },
                { subLabel: '邮箱账号', value: '', col: 'account' },
                { subLabel: '邮箱密码', value: '', col: 'passwd' },
                { subLabel: '渠道名', value: '', col: 'name' },
            ]
        },
        {
            type: 'Dtalk',
            label: '钉钉',
            inputs: [
                { subLabel: 'webhook地址', value: '', col: 'webhook_url' },
                { subLabel: '渠道名', value: '', col: 'name' },
            ]
        },
    ]
}

export { CONSTANT }