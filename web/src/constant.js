import { ApiStrGenerate } from "@/util/viewApi.js";


// 定义一些常量名
const CONSTANT = {
    PAGE: 1,
    // PAGE_SIZE: 8,
    TOTAL: 0,
    DEFALUT_SITE_CONFIG: JSON.stringify({ "logo": "<svg t=\"1702547210136\" class=\"icon\" viewBox=\"0 0 1024 1024\" version=\"1.1\" xmlns=\"http://www.w3.org/2000/svg\" p-id=\"1861\" width=\"200\" height=\"200\"><path d=\"M970.091313 224.033616a14.201535 14.201535 0 0 0-14.100687-3.631838c-88.419556-13.429657-142.919111-12.953859-248.242424 23.762747-69.454869 24.217859-109.056 79.873293-135.267556 116.712728-6.692202 9.403475-18.949172 26.625293-23.77309 29.509818-5.70699-1.303273-21.929374-16.036202-31.762101-24.965172-28.626747-25.994343-64.252121-58.353778-106.767516-58.757172-0.312889-0.005172-0.627071-0.005172-0.939959-0.005171-29.62101 0-54.798222 13.434828-72.857859 38.89907-13.222788 18.651798-17.914828 37.388929-18.106182 38.176323a14.31402 14.31402 0 0 0 2.893576 12.515556 14.336 14.336 0 0 0 11.782465 5.095434l0.409858-0.024565c19.192242-1.152 59.141172-3.545212 65.807516 81.857939 4.288646 54.899071 29.742545 108.631919 69.833697 147.419798 45.499475 44.025535 106.237414 67.298263 175.650909 67.298263 98.97503 0 166.222869-24.480323 205.207272-45.015919 43.172202-22.742626 62.621737-45.918384 63.429819-46.898425a14.170505 14.170505 0 0 0 1.529535-15.874586 14.191192 14.191192 0 0 0-14.166626-7.337373c-37.924202 4.443798-92.767677 6.09099-138.666667-11.298909-7.580444-2.872889-13.878303-5.893172-18.959515-8.691071 20.273131-9.964606 51.000889-28.383677 83.621495-59.93503 34.186343-33.065374 37.722505-69.908687 40.838464-102.423273 2.297535-23.919192 4.666182-48.651636 18.500526-73.091879 28.807758-50.883232 66.645333-74.489535 74.560646-79.035475a14.201535 14.201535 0 0 0 12.135434-7.776969 14.193778 14.193778 0 0 0-2.59103-16.484849z\" fill=\"\" p-id=\"1862\"></path><path d=\"M674.834101 716.481939c-33.422222 4.90796-74.989899 16.884364-117.22602-8.102787-34.843152-20.613172-53.056646-20.993293-77.783919-18.49794-0.156444 0.015515-0.312889 0.015515-0.469334 0.034909-0.274101 0.033616-0.548202 0.060768-0.822303 0.094384-4.697212 0.487434-9.65301 1.065374-14.995394 1.62004-6.702545 0.649051-13.414141 1.19596-20.133495 1.634263-75.333818 4.102465-136.860444-5.381172-163.012525-10.404202-49.488162-9.872808-86.57196-23.921778-108.946101-33.970424-59.832889-26.868364-87.080081-56.671677-87.080081-72.989738 0-1.873455 0.384-2.222545 0.787394-2.585858 2.196687-1.974303 10.616242-6.520242 41.302626-6.004364 24.793212 0.418909 56.970343 3.858101 94.308849 7.853253 40.722101 4.348121 86.873212 9.283232 136.348444 11.686788 6.913293 0.333576 13.959758 0.625778 20.93899 0.858505 9.510788 0.307717 17.545051-7.182222 17.868283-16.707233 0.319354-9.535354-7.175758-17.550222-16.705939-17.868282a1386.24 1386.24 0 0 1-20.419233-0.839112c-48.521051-2.358303-94.161455-7.236525-134.434909-11.544565-90.868364-9.712485-139.353212-13.818828-162.333737 6.833131-8.02004 7.21196-12.262141 17.004606-12.262141 28.317737 0 0.672323 0.029737 1.348525 0.065939 2.026021l0.005172 0.100848c-0.045253 0.524929-0.071111 1.039515-0.071111 1.539879 0 41.464242 9.561212 81.661414 28.414707 119.484768 17.493333 35.088808 42.251636 66.663434 73.583192 93.927434 9.169455-6.551273 20.155475-13.604202 32.524929-20.050748 49.737697-25.921939 97.838545-29.339152 139.099798-9.884444 101.590626 47.906909 142.060606 47.837091 206.198949-0.364606 5.70699-4.287354 13.814949-3.140525 18.103596 2.567758 4.289939 5.708283 3.140525 13.813657-2.567757 18.103595-38.425859 28.877576-69.968162 41.279354-105.143596 41.280647-0.484848 0-0.969697-0.002586-1.457132-0.007758-32.760242-0.316768-69.311354-11.381657-126.16404-38.190545-51.905939-24.47903-106.215434 1.008485-139.548444 23.465374 64.611556 47.922424 146.341495 74.101657 232.704 74.101656 117.55701 0 226.204444-48.934788 292.761858-131.362909l0.002586-0.002586c18.79402-14.959192 12.028121-41.362101-23.442101-36.152889z\" fill=\"\" p-id=\"1863\"></path></svg>", "pagesize": "8", "slogan": "A Message Way Hosted Site", "title": "Message Nest" }),
    LOG_TASK_ID: "T-IM1GBswSRY",
    STORE_TOKEN_NAME: '__message_nest_token__',
    STORE_CUSTOM_NAME: '__message_nest_custom_site__',
    NO_AUTH_URL: [
        '/auth',
    ],
    WAYS_DATA: [
        {
            type: 'Email',
            label: '邮箱',
            inputs: [
                { subLabel: 'smtp服务地址', value: '', col: 'server', desc: "smtp@xyz.com" },
                { subLabel: 'smtp服务端口', value: '', col: 'port', desc: "port" },
                { subLabel: '邮箱账号', value: '', col: 'account', desc: "邮箱账号" },
                { subLabel: '邮箱密码', value: '', col: 'passwd', desc: "邮箱密码" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
                { subLabel: 'html', content: 'html' },
            ],
            taskInsInputs: [
                { value: '', col: 'to_account', desc: "目的邮箱账号（发给谁）" },
                // { value: '', col: 'title', desc: "邮箱标题" },
            ],
        },
        {
            type: 'Dtalk',
            label: '钉钉',
            inputs: [
                { subLabel: 'access_token', value: '', col: 'access_token', desc: "钉钉webhook中的access_token" },
                { subLabel: '加签', value: '', col: 'secret', desc: "加签的签名，SEC开头" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "输入框说明", desc: "钉钉支持加签和关键字过滤，如果是配置了关键字过滤，只需要消息里面包含了关键字，就会发送"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
                { subLabel: 'markdown', content: 'markdown' },
            ],
            taskInsInputs: [
            ],
        },
        {
            type: 'QyWeiXin',
            label: '企业微信',
            inputs: [
                { subLabel: 'token', value: '', col: 'access_token', desc: "企业微信webhook中的token" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
                { subLabel: 'markdown', content: 'markdown' },
            ],
            taskInsInputs: [
            ],
        },
        {
            type: 'Custom',
            label: '自定义推送',
            inputs: [
                { subLabel: 'webhook地址', value: '', col: 'webhook', desc: "自定义webhook地址" },
                { subLabel: '请求体', value: '', col: 'body', desc: "请求体, text内容请使用 TEXT 进行占位\n例如：{\"message\": \"TEXT\", \"foo\": \"bar\"}", isTextArea: true },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "自定义webhook说明", desc: "自定义webhook暂时只支持text，消息将解析TEXT占位标识进行替换，暂时只支持POST方式"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
            ],
            taskInsInputs: [
            ],
        },
        {
            type: 'WeChatOFAccount',
            label: '微信测试公众号模板消息',
            inputs: [
                { subLabel: 'appID', value: '', col: 'appID', desc: "公众号appid" },
                { subLabel: 'appsecret', value: '', col: 'appsecret', desc: "公众号appsecret" },
                { subLabel: '模板id', value: '', col: 'tempid', desc: "模板消息id" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "公众号消息说明", desc: "微信测试公众号模板消息发送，token使用内存缓存，<br />秘钥请访问 https://mp.weixin.qq.com/debug/cgi-bin/sandboxinfo?action=showinfo&t=sandbox/index"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
            ],
            taskInsInputs: [
                { value: '', col: 'to_account', desc: "要发送的OpenId（登录微信公众号后台查看）" },
            ],
        },
        {
            type: 'MessageNest',
            label: '自托管消息',
            inputs: [
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "自托管消息说明", desc: "站点本身会作为消息接收站点，接收、展示推送过来的消息。"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
            ],
            taskInsInputs: [
            ],
        },
    ],
    API_VIEW_DATA: [
        { label: "curl", class: "language-shell line-numbers", code: "", func: ApiStrGenerate.getCurlString },
        { label: "python", class: "language-python line-numbers", code: "", func: ApiStrGenerate.getPythonString },
        { label: "go", class: "language-go line-numbers", code: "", func: ApiStrGenerate.getGolangString },
        { label: "java", class: "language-java line-numbers", code: "", func: ApiStrGenerate.getJaveString },
        { label: "rust", class: "language-rust line-numbers", code: "", func: ApiStrGenerate.getRustString },
        { label: "php", class: "language-php line-numbers", code: "", func: ApiStrGenerate.getPHPString },
        { label: "node", class: "language-javascript line-numbers", code: "", func: ApiStrGenerate.getNodeString },
    ],
}


// 转换渠道map
CONSTANT.WAYS_DATA_MAP = {};
CONSTANT.WAYS_DATA.forEach(element => {
    CONSTANT.WAYS_DATA_MAP[element.type] = element
});

export { CONSTANT }