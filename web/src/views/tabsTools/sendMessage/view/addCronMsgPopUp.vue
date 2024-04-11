<template>
  <el-dialog v-model="isShow" width="400px" :close-on-press-escape="false" :before-close="() => { }"
    :show-close="false">
    <template #header="">
      <el-text class="mx-1">添加定时消息发送</el-text>
      <el-tooltip placement="top">
        <template #content>
          可以定制一些定时消息通知，完成一些简单的提醒事件
          <br />
          ** 需要基于已经创建的发信任务进行绑定
        </template>
        <el-icon>
          <QuestionFilled />
        </el-icon>
      </el-tooltip>
    </template>

    <div class="add-top">
      <el-input v-model="currTaskInput.name" placeholder="请输入消息名称" size="small" class="nameInput"></el-input>
    </div>

    <div class="ins-area">

      <div class="ins-add">
        <el-autocomplete v-model="currSearchInputText" size="small" :fetch-suggestions="querySearchWayAsync"
          placeholder="请输入消息名进行搜索" @select="handleSearchSelect" :clearable="true" value-key="name" />

        <div class="store-area" v-if="isShowAddBox">

          <div class="display-label">
            <el-text class="mx-1" size="small">消息ID：{{ currTaskTmp.id }}</el-text><br />
            <el-text class="mx-1" size="small">消息名：{{ currTaskTmp.name }}</el-text><br />
            <el-text class="mx-1" size="small">消息创建时间：{{ currTaskTmp.created_on }}</el-text><br />
          </div>

          <el-input v-model="currTaskInput.title" placeholder="请输入消息标题" size="small" class="msg-input"></el-input>
          <el-input type="textarea" :rows="5" v-model="currTaskInput.content" placeholder="请输入消息内容" size="small"
            class="msg-input"></el-input>
          <el-input v-model="currTaskInput.cron" placeholder="请输入定时cron表达式" size="small" class="msg-input"></el-input>
          <el-input v-model="currTaskInput.url" placeholder="请输入消息详情url(可选)" size="small" class="msg-input"></el-input>

        </div>
      </div>

    </div>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleCancer()" size="small">取消</el-button>
        <!-- <el-button type="primary" size="small" @click="cliclSendNow()">
          立即发送
        </el-button> -->
        <el-button type="primary" size="small" @click="handleSubmit()">
          确定添加
        </el-button>
      </span>
    </template>

  </el-dialog>
</template>

<script>
import { defineComponent, onMounted, watch, reactive, toRefs } from 'vue';
import { _ } from 'lodash';
import { QuestionFilled } from '@element-plus/icons-vue'
import { usePageState } from '@/store/page_sate.js';
import { request } from '@/api/api'
import { CONSTANT } from '@/constant'
import { CommonUtils } from "@/util/commonUtils.js";
import { generateBizUniqueID } from "@/util/uuid.js";
import { ElMessage } from 'element-plus'

export default defineComponent({
  components: {
    QuestionFilled,
  },
  props: {
    componentName: String
  },
  setup(props) {
    const pageState = usePageState();
    const state = reactive({
      insTableData: [],
      currSearchWaysData: [],
      currSearchInputText: '',
      isShow: false,
      isShowAddBox: false,
      searchway_id: '',
      currTaskTmp: {},
      currInsInput: {},
      currInsInputContentType: 'text',
      currTaskInput: {
        name: '',
        title: '',
        content: '',
        cron: '',
        url: '',
        id: generateBizUniqueID('C'),
      },
      sysKeptTaskIds: [CONSTANT.LOG_TASK_ID],

    });

    watch(pageState.ShowDialogData, (newValue, oldValue) => {
      if (newValue[props.componentName]) {
        state.isShow = pageState.ShowDialogData[props.componentName].isShow;
        resetPageInitData();
      }
    });

    const handleCancer = () => {
      if (pageState.ShowDialogData[props.componentName]) {
        pageState.ShowDialogData[props.componentName].isShow = false;
      }
    }

    // 页面每次弹出，重置数据
    const resetPageInitData = () => {
      state.insTableData = [];
      state.currInsInputContentType = 'text';
      state.currInsInput = {};
      state.currTaskTmp = {};
      state.searchway_id = '';
      state.isShowAddBox = false;
      state.currTaskInput = {
        name: '',
        id: generateBizUniqueID('C'),
      }
    }

    const cliclSendNow = () => {
      console.log('cliclSendNow');
    }

    // 匹配出当前搜索的消息数据
    const matchSearchData = (way_name) => {
      let result = {};
      state.currSearchWaysData.forEach(element => {
        if (element.name == way_name) {
          result = element;
        }
      });
      return result;
    }

    const handleSearchSelect = async () => {
      let currTask = matchSearchData(state.currSearchInputText)
      state.isShowAddBox = Boolean(currTask);
      if (currTask) {
        state.currTaskTmp = currTask;
      }
    }

    const querySearchWayAsync = async (query, cb) => {
      let params = { name: query };
      const rsp = await request.get('/sendtasks/list', { params: params });
      let tableData = await rsp.data.data.lists;
      tableData = filtersysKeptTask(tableData);
      cb(tableData);
      state.currSearchWaysData = tableData;
    }

    const filtersysKeptTask = (data) => {
      let result = data.filter(item => !state.sysKeptTaskIds.includes(item.id));
      return result;
    }

    const getFinalData = () => {
      state.currTaskInput.task_id = state.currTaskTmp.id;
      return state.currTaskInput
    }

    const handleSubmit = async () => {
      let postData = getFinalData();
      const rsp = await request.post('/sendmessages/addone', postData);
      if (await rsp.data.code == 200) {
        ElMessage({ message: await rsp.data.msg, type: 'success' });
        handleCancer();
        window.location.reload();
      }
    }

    return {
      ...toRefs(state), handleCancer, handleSubmit, querySearchWayAsync,
      handleSearchSelect, CONSTANT, CommonUtils,
      cliclSendNow
    };
  },
});
</script>

<style scoped>
:deep(.el-autocomplete) {
  width: 100%;
}

.msg-input {
  /* width: 200px; */
  margin: 5px 10px 5px 0;
}

.add-top {
  margin-bottom: 20px;
  margin-top: -20px;
}

.display-label {
  margin-top: 10px;
  margin-bottom: 10px;
}
</style>
