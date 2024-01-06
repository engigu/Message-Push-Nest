<template>
  <el-dialog v-model="isShow" width="500px" :close-on-press-escape="false" :before-close="() => { }" :show-close="false">
    <template #header="">
      <el-text class="mx-1">新增发信渠道</el-text>
    </template>

    <el-tabs v-model="activeName" class="demo-tabs" @tab-click="handleClick">
      <el-form label-width="100px" v-for="item in waysLabelData">
        <el-tab-pane :label="item.label" :name="item.type">
          <el-form-item :label="one.subLabel" v-for="one in item.inputs">
            <el-input v-if="one.isTextArea != true" size="small" v-model="one.value" :placeholder="one.desc" />
            <el-input v-if="one.isTextArea == true" size="small" type="textarea" v-model="one.value"
              :placeholder="one.desc" :autosize="{ minRows: 4, maxRows: 10 }" />
          </el-form-item>
          <textTips v-if="item.tips" :text="item.tips.text" :desc="item.tips.desc" />
        </el-tab-pane>
      </el-form>
    </el-tabs>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleCancer()" size="small">取消</el-button>
        <testSendButton @customhandleSubmit="handleTest()" />
        <el-button type="primary" size="small" @click="handleSubmit()">
          确定添加
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script>
import { defineComponent, onMounted, watch, reactive, toRefs } from 'vue';
import { usePageState } from '@/store/page_sate.js';
import { request } from '@/api/api'
import { CONSTANT } from '@/constant'
import { _ } from 'lodash';
import { ElMessage } from 'element-plus'
import textTips from '@/views/common/textTips.vue'
import testSendButton from './testSendButton.vue'


export default defineComponent({
  components: {
    textTips,
    testSendButton,
  },
  props: {
    componentName: String
  },
  setup(props) {
    const pageState = usePageState();
    const state = reactive({
      isShow: false,
      testMseeageDialogVisible: false,
      activeName: "Email",
      waysLabelData: _.cloneDeep(CONSTANT.WAYS_DATA),

    });

    watch(pageState.ShowDialogData, (newValue, oldValue) => {
      if (newValue[props.componentName]) {
        state.isShow = pageState.ShowDialogData[props.componentName].isShow;
      }
    });

    const handleCancer = () => {
      if (pageState.ShowDialogData[props.componentName]) {
        pageState.ShowDialogData[props.componentName].isShow = false;
      }
    }

    const handleClick = () => {
    }

    const getInputData = (type) => {
      for (const element of state.waysLabelData) {
        if (element.type == type) {
          const data = {};
          for (const item of element.inputs) {
            data[item.col] = item.value;
          }
          return data;
        }
      }
      return {};
    }

    const getFinalData = () => {
      const inputData = getInputData(state.activeName);
      const { name, ...nameObject } = inputData;
      let postData = { name: name }
      const { name: _, ...auth } = inputData;
      if (state.activeName == 'Email') {
        auth.port = Number(auth.port)
      };
      postData.auth = JSON.stringify(auth);
      postData.type = state.activeName;
      return postData
    }

    const handleTest = async () => {
      let postData = getFinalData();
      const rsp = await request.post('/sendways/test', postData);
      if (await rsp.data.code == 200) {
        ElMessage({ message: await rsp.data.msg, type: 'success' })
      }
    }

    const handleSubmit = async () => {
      let postData = getFinalData();
      const rsp = await request.post('/sendways/add', postData);
      if (await rsp.data.code == 200) {
        handleCancer();
      }

    }

    return {
      ...toRefs(state), handleCancer, handleSubmit, handleClick, handleTest
    };
  },
});
</script>

<style scoped>
/* :global(.el-dialog) {
  width: 500px;

} */
/* :global(.el-dialog__title) {
  font-size: 14px;
}

:global(.el-dialog label) {
  font-size: 13px;
} */
</style>
