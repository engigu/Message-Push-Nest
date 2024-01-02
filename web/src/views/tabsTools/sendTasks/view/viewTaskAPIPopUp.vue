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
          发送的消息会优先现在相应的类型消息进行发送，如果没有，将使用传的text消息进行发送          
          <br />
          *text节点必传
        </template>
        <el-icon>
          <QuestionFilled />
        </el-icon>
      </el-tooltip>
    </template>

    <el-tabs v-model="activeName" @tab-click="renderApiString">
      <el-tab-pane label="curl" name="curl">
        <pre><code class="language-shell line-numbers">{{ currCode }}</code></pre>
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
import { ApiStrGenerate } from "@/util/viewApi.js";
import { QuestionFilled } from '@element-plus/icons-vue'

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
      currCode: '',
      activeName: 'curl',
      language: "python",
    });

    watch(pageState.ShowDialogData, (newValue, oldValue) => {
      if (newValue[props.componentName]) {
        state.isShow = pageState.ShowDialogData[props.componentName].isShow;
        renderApiString();
      }
    });

    const handleCancer = () => {
      if (pageState.ShowDialogData[props.componentName]) {
        pageState.ShowDialogData[props.componentName].isShow = false;
      }
    }

    const renderApiString = () => {
      let task_id = pageState.ShowDialogData[props.componentName].rowData.id;
      if (state.activeName == 'curl') {
        state.currCode = ApiStrGenerate.getCurlString(task_id, {});
      }
      setTimeout(() => {
        Prism.highlightAll()
      }, 100)
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


/* pre {
    overflow: hidden !important;
    code{ 
        display: inline-block;
        padding-bottom: 20px;
        position: relative;
        top: 20px;
    }
    &::before {
        content: "";
        position: absolute;
        background: red;
        width: 10px;
        height: 10px;
        border-radius: 50%;
        top: 10px;
        left: 15px;
        transform: translate(-50%);
    }
    &::after {
        content: "";
        position: absolute;
        background: sandybrown;
        width: 10px;
        height: 10px;
        border-radius: 50%;
        top: 10px;
        left: 30px;
        transform: translate(-50%);
    }
    code:first-child{
        &::after{
            content: "";
            position: absolute;
            background: limegreen;
            width: 10px;
            height: 10px;
            border-radius: 50%;
            top: -24px;
            left: -7px;
            transform: translate(-50%);
        }
    }
} */
</style>

