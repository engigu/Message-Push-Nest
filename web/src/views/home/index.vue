<template>
  <div class="inex-title-bar" v-if="pageState.isLogin">
    <el-menu :collapse="isCollapse" breakpoint="768px" mode="horizontal" @select="handleSelect()"
      :default-active="currActivate()" :ellipsis="false" :menu-width="'auto'">
      <el-menu-item index="0" :disabled="false">
        <img style="width: 60px" src="../../../public/titlelogo.svg" alt="Element logo" />
      </el-menu-item>

      <div class="flex-grow" style="flex-grow: 1" />
      <div v-for="(item, index) in menuData" :key="index" class="banner-title">
        <router-link :to="{ path: item.path }">
          <el-menu-item :index="item.id">{{ item.title }}</el-menu-item>
        </router-link>
      </div>
      <el-button plain class="logout-btn" @click="clickLogout()">登出</el-button>
    </el-menu>
  </div>
  <router-view></router-view>
</template>

<script>
import { ref, onMounted, reactive } from 'vue';
import { usePageState } from '../../store/page_sate.js';
import { CONSTANT } from '../../constant'
import { useRouter, useRoute } from 'vue-router';

export default {
  setup() {
    const pageState = usePageState();
    const router = useRouter();
    const isCollapse = ref(false);
    const menuData = reactive([
      {
        id: '1',
        title: '发信日志',
        path: '/sendlogs',
      }, {
        id: '2',
        title: '发信任务',
        path: '/sendtasks',
      }, {
        id: '3',
        title: '发信渠道',
        path: '/sendways',
      }, {
        id: '4',
        title: '设置',
        path: '/settings',
      },
    ]);

    const checkIsLogin = () => {
      pageState.isLogin = Boolean(localStorage.getItem(CONSTANT.STORE_TOKEN_NAME));
    };

    const clickLogout = () => {
      localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME);
      router.push('/login', { replace: true }).then(() => { router.go() });
    };

    const loadLocalToken = () => {
      pageState.setToken(localStorage.getItem(CONSTANT.STORE_TOKEN_NAME));
    }

    const currActivate = () => {
      const cur_path = useRoute().path;
      let result = '1';
      menuData.forEach(element => {
        if (element.path == cur_path) {
          result = element.id;
        };
      });
      return result;
    }

    onMounted(() => {
      checkIsLogin();
      loadLocalToken();
    });

    const handleSelect = () => { };

    return {
      isCollapse, handleSelect, menuData, pageState, clickLogout, currActivate
    };
  },
};
</script>

<style scoped>
.el-menu-item ul {
  height: 15vh;
}

.el-menu-item {
  width: 150px !important;
  font-size: 15px;
  justify-content: center;
  align-items: center;

}

.el-menu-item li {
  margin: 0 auto;
}

.logout-btn {
  margin: auto 40px auto 40px;
  float: right;
  /* background-color: #f7efee;
  border: none;
  padding: 10px 20px;
  text-align: center;
  text-decoration: none;
  font-size: 16px;
  cursor: pointer;
  border-radius: 5px;
  transition: background-color 0.3s ease; */
}
</style>
