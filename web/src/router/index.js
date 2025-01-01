import { createRouter, createWebHistory, createWebHashHistory } from 'vue-router'
import LoginInex from '../views/home/login.vue'
import { CONSTANT } from '../constant'

const router = createRouter({
  // history: createWebHistory(import.meta.env.BASE_URL),
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginInex
    },
    {
      path: '/statistic',
      name: 'statistic',
      alias: '/',
      component: () => import('../views/tabsTools/statistic/statistic.vue')
    },
    {
      path: '/sendways',
      name: 'sendWays',
      component: () => import('../views/tabsTools/sendWays/sendWays.vue')
    },
    {
      path: '/sendtasks',
      name: 'sendtasks',
      component: () => import('../views/tabsTools/sendTasks/sendTasks.vue')
    },
    {
      path: '/cronmessages',
      name: 'cronmessages',
      component: () => import('../views/tabsTools/cronMessage/cronMessage.vue')
    },
    {
      path: '/sendlogs',
      name: 'sendlogs',
      component: () => import('../views/tabsTools/sendLogs/sendLogs.vue')
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../views/tabsTools/settings/settings.vue')
    },   
    {
      path: '/hostedMessage',
      name: 'hostedmessage',
      component: () => import('../views/tabsTools/hostedMessage/hostedMessage.vue')
    },
    {
      path: '/:catchAll(.*)',
      name: '404',
      component: () => import('../views/404.vue')
    },
  ]
})

// 登录失效重定向到登录页面
router.beforeEach((to, from, next) => {
  const isAuthenticated = Boolean(localStorage.getItem(CONSTANT.STORE_TOKEN_NAME));
  if (!isAuthenticated && to.path !== '/login') {
    next('/login');
    setTimeout(() => {
      window.location.reload();
    }, 100);
  } else {
    next();
  }
});

export default router
