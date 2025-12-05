import { createRouter, createWebHistory, createWebHashHistory } from 'vue-router'
// import LoginInex from '../components/Login.vue'
import { CONSTANT } from '../constant'

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
      path: '/:catchAll(.*)',
      name: '404',
      component: () => import('../components/404.vue')
    },
  ]
})

// 登录失效重定向到登录页面
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem(CONSTANT.STORE_TOKEN_NAME);
  const isAuthenticated = Boolean(token && token.trim() !== '');

  // 404页面不需要登录验证
  if (to.name === '404') {
    next();
    return;
  }
  
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
