<template>
  <el-dialog v-model="isShow" width="400px" :close-on-press-escape="false" :before-close="() => { }"
    :show-close="false">
    <template #header="">
      <el-text class="mx-1">编辑发信任务</el-text>

    </template>

    <div class="add-top">
      <el-input v-model="currTaskInput.name" placeholder="请输入消息名称" size="small" class="nameInput"></el-input>
    </div>

    <div class="ins-area">

      <div class="ins-add">
        <el-autocomplete v-model="currSearchInputText" size="small" :fetch-suggestions="querySearchWayAsync"
          placeholder="请输入任务名进行搜索" @select="handleSearchSelect" :clearable="true" value-key="name" />

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
        <el-button type="primary" size="small" @click="handleEditCronMsg()">
          确定修改
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
import tableDeleteButton from '@/views/common/tableDeleteButton.vue'
import { ElMessage } from 'element-plus'
import { CONSTANT } from '@/constant'
import { CommonUtils } from "@/util/commonUtils.js";
import { generateBizUniqueID } from "@/util/uuid.js";


export default defineComponent({
  components: {
    QuestionFilled,
    tableDeleteButton,
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
      searchWayID: '',
      currInsInputContentType: '',
      currTaskTmp: {},
      currInsInput: {},
      currTaskInput: {
        name: '',
        title: '',
        content: '',
        cron: '',
        url: '',
      },
      sysKeptTaskIds: [CONSTANT.LOG_TASK_ID],
    });

    watch(pageState.ShowDialogData, (newValue, oldValue) => {
      if (newValue[props.componentName]) {
        let data = pageState.ShowDialogData[props.componentName];
        state.isShow = data.isShow;
        resetPageInitData();
        state.currTaskInput.taskId = data.rowData.id;
        if (data && state.isShow) {
          state.currTaskInput = data.rowData;
          InitOpenSeletValue(state.currTaskInput)
        }
      }
    });

    const InitOpenSeletValue = async (data) => {
      // 填充编辑框数据
      let task_id = data.task_id;
      const rsp = await request.get('/sendtasks/get', { params: { id: task_id } });
      let rsp_data = await rsp.data;
      state.isShowAddBox = true;
      state.currSearchInputText = rsp_data.data.name;
      state.currTaskTmp = {
        id: rsp_data.data.id,
        name: rsp_data.data.name,
        created_on: rsp_data.data.created_on,
      };
    }

    const handleCancer = () => {
      if (pageState.ShowDialogData[props.componentName]) {
        pageState.ShowDialogData[props.componentName].isShow = false;
      }
    }

    // 页面每次弹出，重置数据
    const resetPageInitData = () => {
      state.insTableData = [];
      state.currInsInput = {};
      state.currTaskTmp = {};
      state.currInsInput = {};
      state.searchWayID = '';
      state.currInsInputContentType = '';
      state.isShowAddBox = false;
      state.currTaskInput = {
        taskName: '',
        // taskId: uuidv4(),
      }
    }

    // 匹配出当前搜索的任务数据
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
      let postData = {
        name: state.currTaskInput.name,
        title: state.currTaskInput.title,
        content: state.currTaskInput.content,
        cron: state.currTaskInput.cron,
        url: state.currTaskInput.url,
        id: state.currTaskInput.id,
        enable: state.currTaskInput.enable ? 1 : 0,
        task_id: state.currTaskTmp.id,
      };
      return postData
    }

    const handleEditCronMsg = async () => {
      let postData = getFinalData();
      console.log('postdata', postData);
      const rsp = await request.post('/sendmessages/edit', postData);
      if (await rsp.data.code == 200) {
        ElMessage({ message: await rsp.data.msg, type: 'success' });
        setTimeout(() => {
          handleCancer();
          window.location.reload();
        }, 1000)
      }
    }

    return {
      ...toRefs(state), handleCancer, CONSTANT, CommonUtils,
      handleSearchSelect, querySearchWayAsync, handleEditCronMsg
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
