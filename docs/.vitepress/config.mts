import { defineConfig } from 'vitepress'

export default defineConfig({
  title: "Message Nest",
  description: "灵活而强大的消息推送整合平台",
  lang: 'zh-CN',
  
  // 如果部署到 GitHub Pages 的子路径，需要设置 base
  // 例如：https://engigu.github.io/Message-Push-Nest/
  // base: '/Message-Push-Nest/',
  
  // 如果使用自定义域名或部署到根路径，注释掉 base 或设置为 '/'
  base: '/Message-Push-Nest/',
  
  themeConfig: {
    logo: '/logo.svg',
    
    nav: [
      { text: '首页', link: '/' },
      { text: '指南', link: '/guide/introduction' },
      { text: '部署', link: '/deployment/overview' },
      { text: 'API', link: '/api/usage' },
      { text: '演示站点', link: 'https://message-nest-demo-site.qwapi.eu.org/' }
    ],

    sidebar: {
      '/guide/': [
        {
          text: '开始',
          items: [
            { text: '介绍', link: '/guide/introduction' },
            { text: '特色功能', link: '/guide/features' },
            { text: '更新日志', link: '/guide/changelog' }
          ]
        },
        {
          text: '配置',
          items: [
            { text: '配置说明', link: '/guide/configuration' },
            { text: 'EmbedHtml说明', link: '/guide/embed-html' }
          ]
        },
        // {
        //   text: '文档部署',
        //   items: [
        //     { text: '部署到 GitHub Pages', link: '/guide/deploy-to-github-pages' }
        //   ]
        // }
      ],
      '/deployment/': [
        {
          text: '部署方式',
          items: [
            { text: '部署概览', link: '/deployment/overview' },
            { text: '直接运行', link: '/deployment/direct-run' },
            { text: '开发调试', link: '/deployment/development' },
            { text: 'Docker部署', link: '/deployment/docker' },
            { text: 'Docker Compose', link: '/deployment/docker-compose' }
          ]
        }
      ],
      '/api/': [
        {
          text: 'API文档',
          items: [
            { text: '使用说明', link: '/api/usage' },
            { text: '调用示例', link: '/api/examples' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/engigu/Message-Push-Nest' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024-present Message Nest'
    },

    search: {
      provider: 'local'
    },

    outline: {
      level: [2, 3],
      label: '目录'
    },

    docFooter: {
      prev: '上一页',
      next: '下一页'
    },

    lastUpdated: {
      text: '最后更新于',
      formatOptions: {
        dateStyle: 'short',
        timeStyle: 'short'
      }
    }
  }
})
