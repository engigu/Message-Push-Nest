import { createRouter, createWebHistory, createWebHashHistory } from 'vue-router'
// import LoginInex from '../components/Login.vue'
import { CONSTANT } from '../constant'
import { usePageState } from '../store/page_sate'

const router = createRouter({
  // 使用 HTML5 History 模式，确保 URL 变化反映在浏览器地址栏中
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component:() => import('../components/Login.vue')
    },
    {
      path: '/',
      name: 'index',
      component: () => import('../components/Index.vue'),
      children: [
        {
          // 默认子路由，显示 Dashboard
          path: '',
          name: 'dashboard',
          component: () => import('../components/pages/dashboard/Dashboard.vue')
        },
        {
          path: 'sendlogs',
          name: 'sendlogs',
          component: () => import('../components/pages/sendLogs/SendLogs.vue')
        },
        {
          path: 'hostedmessage',
          name: 'hostedmessage',
          component: () => import('../components/pages/hostedMessage/HostedMessage.vue')
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('../components/pages/settings/Settings.vue')
        },
        {
          path: 'sendways',
          name: 'sendways',
          component: () => import('../components/pages/sendWays/SendWays.vue')
        },
        {
          path: 'sendtasks',
          name: 'sendtasks',
          component: () => import('../components/pages/sendTasks/SendTasks.vue')
        },
        {
          path: 'cronmessages',
          name: 'cronmessages',
          component: () => import('../components/pages/cronMessages/CronMessages.vue')
        },
        {
          path: 'templates',
          name: 'templates',
          component: () => import('../components/pages/messageTemplate/MessageTemplate.vue')
        }
      ]
    },
    // {
    //   path: '/settings',
    //   name: 'settings',
    //   component: () => import('../views/tabsTools/settings/settings.vue')
    // },   
    // {
    //   path: '/hostedMessage',
    //   name: 'hostedmessage',
    //   component: () => import('../views/tabsTools/hostedMessage/hostedMessage.vue')
    // },
    {
      path: '/preview/:key',
      name: 'preview',
      component: () => import('../components/pages/hostedMessage/MessagePreview.vue')
    },
    {
      path: '/notice',
      name: 'notice',
      component: () => import('../components/pages/Notice.vue')
    },
    {
      path: '/:catchAll(.*)',
      name: '404',
      component: () => import('../components/404.vue')
    },
  ]
})

// 登录失效重定向与 SSO Token 自动拦截处理
router.beforeEach((to, from, next) => {
  // 如果 URL query 中携带了 token（来自 OIDC / OAuth 2.0 单点登录成功回调）
  if (to.query && to.query.token) {
    const ssoToken = String(to.query.token);
    if (ssoToken.trim() !== '') {
      localStorage.setItem(CONSTANT.STORE_TOKEN_NAME, ssoToken);
      try {
        const pageState = usePageState();
        pageState.setToken(ssoToken);
      } catch (e) {
        console.warn('同步 SSO Token 失败:', e);
      }
      // 清除 URL query 避免敏感 Token 长期停留在地址栏
      const nextQuery = { ...to.query };
      delete nextQuery.token;
      next({ path: to.path === '/login' ? '/' : to.path, query: nextQuery, replace: true });
      return;
    }
  }

  // 404页面、公开预览页面或系统提示页面不需要登录验证
  if (to.name === '404' || to.name === 'preview' || to.name === 'notice' || to.path === '/notice') {
    next();
    return;
  }

  // 如果访问登录页且带有错误或提示信息，直接导向提示信息页面展示
  if (to.path === '/login' && to.query && (to.query.error || to.query.message)) {
    next({ path: '/notice', query: to.query });
    return;
  }

  const token = localStorage.getItem(CONSTANT.STORE_TOKEN_NAME);
  const isAuthenticated = Boolean(token && token.trim() !== '');
  
  // 如果没有token且不是访问登录页，跳转到登录页
  if (!isAuthenticated && to.path !== '/login') {
    next('/login');
  } 
  // 如果有token且访问登录页，跳转到首页
  else if (isAuthenticated && to.path === '/login') {
    next('/');
  } 
  // 其他情况正常访问
  else {
    next();
  }
});

export default router
