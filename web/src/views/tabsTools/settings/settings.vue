<template>
  <div class="main-center-body">
    <div class="container">
      <div class="setting-main" style="display: flex;">
        <div style="width: 120px;">
          <el-menu :default-active="currTab" mode="vertical" width="100px" @select="handleSelect">
            <el-menu-item index="resetPasswd">
              重置密码
            </el-menu-item>
            <el-menu-item index="logCleanSetting">
              日志清理
            </el-menu-item>
            <el-menu-item index="siteCustomSetting">
              站点设置
            </el-menu-item>
            <el-menu-item index="aboutSetting">
              站点关于
            </el-menu-item>
          </el-menu>
        </div>
        <div class="setting-right">
          <resetPasswd v-if="currTab == 'resetPasswd'" />
          <setLogClean v-if="currTab == 'logCleanSetting'" />
          <siteCustom v-if="currTab == 'siteCustomSetting'" />
          <aboutSetting v-if="currTab == 'aboutSetting'" />
        </div>
      </div>
    </div>
  </div>
</template>

<script >
import { reactive, toRefs, onMounted } from 'vue'
import { copyToClipboard } from '@/util/clipboard.js';

import resetPasswd from './view/resetPasswd.vue'
import setLogClean from './view/setLogClean.vue'
import siteCustom from './view/siteCustom.vue'
import aboutSetting from './view/about.vue'

export default {
  components: {
    resetPasswd,
    setLogClean,
    siteCustom,
    aboutSetting,
  },
  setup() {
    const state = reactive({
      currTab: 'resetPasswd',
    });

    const handleSelect = (index) => {
      state.currTab = index;
    }

    onMounted(async () => {
    });

    return {
      ...toRefs(state), copyToClipboard, handleSelect
    };
  }
}
</script>


<style scoped>
hr {
  color: #FAFCFF;
  background-color: #FAFCFF;
  border-color: #FAFCFF;
}

.container {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0px 2px 4px rgba(0, 0, 0, 0.1);
  max-width: 800px;
  width: 100%;
  margin-top: -10vh;
}

:deep(.el-input ) {
    margin-top: 14px;
}

.setting-right {
  width: 100%;
}
</style>