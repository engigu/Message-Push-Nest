<template>
  <el-dialog v-model="isShow" width="58%" :close-on-press-escape="false" :before-close="() => { }" :show-close="false">
    <template #header="">
      <el-text class="mx-1">查看接入API</el-text>
      <el-tooltip placement="top">
        <template #content>
          一个任务可能关联多个不同渠道的实例
          <br />
          实例的内容类型大体上可以可以分为text、html、markdown          
          <br />
          微信测试公众号需要额外传递url字段，可以跳转到消息的详情地址
          <br />
          发送的消息会优先选择相应的类型消息进行发送，如果没有，将使用传的text消息进行发送
          <br />
          ** text节点必传，指定mode=sync将同步发送，默认异步发送
        </template>
        <el-icon>
          <QuestionFilled />
        </el-icon>
      </el-tooltip>
    </template>

    <el-tabs v-model="activeName">
      <el-tab-pane :label="item.label" :name="item.label" v-for="item in apiViewData">
        <pre><code :class="item.class">{{ item.code }}</code></pre>
      </el-tab-pane>
    </el-tabs>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleCancer()" size="small">取消</el-button>
      </span>
    </template>

  </el-dialog>
</template>

<script>
import { defineComponent, onMounted, watch, reactive, toRefs, onUpdated } from 'vue';
import { _ } from 'lodash';
import { usePageState } from '@/store/page_sate.js';
import Prism from "prismjs";
import { QuestionFilled } from '@element-plus/icons-vue'
import { request } from '@/api/api'
import { CONSTANT } from '@/constant'


export default defineComponent({
  components: {
    QuestionFilled,
  },
  props: {
    componentName: String,
  },
  setup(props) {
    const pageState = usePageState();
    const state = reactive({
      isShow: false,
      currOption: '',
      activeName: 'curl',
      apiViewData: CONSTANT.API_VIEW_DATA,
    });

    watch(pageState.ShowDialogData, (newValue, oldValue) => {
      if (newValue[props.componentName]) {
        state.isShow = pageState.ShowDialogData[props.componentName].isShow;
        if (state.isShow) {
          renderApiString();
        }
      }
    });

    const handleCancer = () => {
      if (pageState.ShowDialogData[props.componentName]) {
        pageState.ShowDialogData[props.componentName].isShow = false;
        state.currOption = '';
      }
    }

    // 获取接口查看格式
    const getViewOptions = async (taskId) => {
      let params = { id: taskId };
      const rsp = await request.get('/sendtasks/ins/gettask', { params: params });
      let insTableData = await rsp.data.data.ins_data;

      let viewOptions = {}
      insTableData.forEach(element => {
        viewOptions[element.content_type] = 1;
        if (element.way_type == "WeChatOFAccount") {
          viewOptions['url'] = 1;
        }
      });
      return viewOptions
    }

    // 渲染api接口代码
    const renderApiString = async () => {
      let task_id = pageState.ShowDialogData[props.componentName].rowData.id;
      if (!state.currOption) {
        state.currOption = await getViewOptions(task_id);
      }
      state.apiViewData.forEach(element => {
        // if (!element.code) {
        element.code = element.func(task_id, state.currOption);
        // }
      });
      setTimeout(() => {
        Prism.highlightAll()
      }, 1)
    }

    return {
      ...toRefs(state), handleCancer, renderApiString
    };
  },
});
</script>

<style scoped>
.language-javascript {
  background-color: #f0f0f0;
}
</style>

