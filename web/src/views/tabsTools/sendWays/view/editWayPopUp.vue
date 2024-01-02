<template>
  <el-dialog v-model="isShow" width="500px" :close-on-press-escape="false" :before-close="() => { }" :show-close="false">
    <template #header="">
      <el-text class="mx-1">编辑发信渠道</el-text>
    </template>

    <el-tabs class="demo-tabs" @tab-click="handleClick">

      <el-form label-width="100px" v-for="item in waysLabelData">
        <!-- <el-tab-pane :label="item.label" :name="item.type"> -->
        <el-form-item :label="one.subLabel" v-for="one in item.inputs">
          <el-input v-model="one.value" />
        </el-form-item>
        <!-- </el-tab-pane> -->
      </el-form>

    </el-tabs>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="testMseeageDialogVisible = true" size="small">测试发信</el-button>
        <el-button @click="handleCancer()" size="small">取消</el-button>
        <el-button type="primary" size="small" @click="handleSubmit()">
          确定编辑
        </el-button>
      </span>
    </template>
  </el-dialog>

  <el-dialog v-model="testMseeageDialogVisible" width="50%" align-center>
    <span>将发送一条测试信息，将注意查收</span>
    <template #footer>
      <span class="dialog-footer">
        <el-button size="small" @click="testMseeageDialogVisible = false">取消</el-button>
        <el-button size="small" type="primary" @click="(testMseeageDialogVisible = false) || handleTest()">
          确定发送
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


export default defineComponent({
  props: {
    componentName: String
  },
  setup(props) {
    const pageState = usePageState();
    const state = reactive({
      isShow: false,
      testMseeageDialogVisible: false,
      waysLabelData: [],
      editData: {},

    });

    const dealDisplayData = () => {

    }

    // 监测父页面传过来的数据
    watch(pageState.ShowDialogData, (newValue, oldValue) => {
      if (newValue[props.componentName]) {
        // 弹出编辑框
        state.isShow = pageState.ShowDialogData[props.componentName].isShow;

        // 展示编辑框
        if (newValue[props.componentName].rowData) {
          const row = pageState.ShowDialogData[props.componentName].rowData;
          let nowData = [];
          _.cloneDeep(CONSTANT.WAYS_DATA).forEach(element => {
            if (element.type == row.type) {
              // 填充输入框的值
              state.editData = row;
              element.inputs.forEach(one => {
                let newRow = Object.assign(row, JSON.parse(row.auth));
                if (newRow[one.col]) {
                  one.value = newRow[one.col];
                };
              });
              nowData.push(element);
            };
          });
          state.waysLabelData = nowData;
        }
      };
    });


    const handleCancer = () => {
      if (pageState.ShowDialogData[props.componentName]) {
        pageState.ShowDialogData[props.componentName].isShow = false;
      }
    }


    const handleClick = () => {
    }

    const getEditData = (type) => {
      for (const element of state.waysLabelData) {
        // if (element.type == type) {
        const data = {};
        for (const item of element.inputs) {
          data[item.col] = item.value;
        }
        return data;
        // }
      }
      return {};
    }
    const getFinalData = () => {
      const editData = getEditData(state.waysLabelData);
      const { name, ...nameObject } = editData;
      let postData = { name: name }
      const { name: _, ...auth } = editData;
      if (state.editData.type == 'Email') {
        auth.port = Number(auth.port)
      };
      postData.auth = JSON.stringify(auth);
      postData.type = state.editData.type;
      postData.id = state.editData.id;
      return postData
    }

    const handleTest = async () => {
      let postData = getFinalData();
      const rsp = await request.post('/sendways/test', postData);
      if (await rsp.data.code == 200) {
        ElMessage({ message: response.data.msg, type: 'success' })
      }
    }

    const handleSubmit = async () => {
      let postData = getFinalData();
      const rsp = await request.post('/sendways/edit', postData);
      // console.log('edit res',rsp)
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
} */

/* :global(.el-dialog label) {
  font-size: 13px;
} */
</style>
