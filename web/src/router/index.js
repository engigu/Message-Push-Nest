import { createRouter, createWebHistory, createWebHashHistory } from 'vue-router'
// import LoginInex from '../components/Login.vue'
import { CONSTANT } from '../constant'

const router = createRouter({
  // 使用 HTML5 History 模式，确保 URL 变化反映在浏览器地址栏中
  history: createWebHistory(),
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
    // {
    //   path: '/:catchAll(.*)',
    //   name: '404',
    //   component: () => import('../views/404.vue')
    // },
  ]
})

// 登录失效重定向到登录页面
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem(CONSTANT.STORE_TOKEN_NAME);
  const isAuthenticated = Boolean(token && token.trim() !== '');

  
  if (!isAuthenticated && to.path !== '/login') {

    next('/login');
  } else if (isAuthenticated && to.path === '/login') {

    next('/');
  } else {
    next();
  }
});

export default router
