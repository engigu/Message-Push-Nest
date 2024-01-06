<template>
  <div class="main-center-body">
    <div class="container">

      <div class="search-input-sendways">
        <el-input v-model="search" size="small" placeholder="搜索" @change="filterFunc()" />
      </div>

      <div class="search-box">
        <el-button size="small" type="primary" @click="clickAdd">新增任务</el-button>
      </div>

      <hr />

      <div ref="refContainer">
        <el-table :data="tableData" stripe empty-text="发信任务为空" :row-style="rowStyle()">
          <el-table-column label="ID" prop="id" width="320px" />
          <el-table-column label="任务名" prop="name" />
          <el-table-column label="创建时间" prop="created_on" />
          <el-table-column fixed="right" label="操作" width="190px">
            <template #default="scope">
              <div v-if="sysKeptTaskIds.includes(scope.row.id)">
                <el-button link size="small" type="primary"
                  @click="handleViewLogs(scope.$index, scope.row)">日志</el-button>
                <el-text class="mx-1" type="info" size="small" style="margin-left: 15px;">系统保留任务</el-text>
              </div>
              <div v-if="!sysKeptTaskIds.includes(scope.row.id)">
                <el-button link size="small" type="primary" @click="handleViewAPI(scope.$index, scope.row)">接口</el-button>
                <el-button link size="small" type="primary"
                  @click="handleViewLogs(scope.$index, scope.row)">日志</el-button>
                <el-button link size="small" type="primary" @click="handleEdit(scope.$index, scope.row)">编辑</el-button>
                <tableDeleteButton @customHandleDelete="handleDelete(scope.$index, scope.row)" />
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="pagination-block">
        <el-pagination layout="prev, pager, next" :total="total" :page-size="pageSize" @current-change="handPageChange" />
        <el-text class="total-tip" size="small">共{{ total }}条</el-text>
      </div>

    </div>
  </div>

  <addTaskPopUpComponent :componentName="addTaskPopUpComponentName" />
  <editTaskPopUpComponent :componentName="editTaskPopUpComponentName" />
  <viewTaskAPIPopUpComponent :componentName="viewTaskAPIPopUpComponentName" />
</template>

<script >
import { computed, ref, reactive, toRefs, onMounted } from 'vue'
// import { InfoFilled } from '@element-plus/icons-vue'
import { usePageState } from '@/store/page_sate';
import { request } from '@/api/api'
import addTaskPopUpComponent from './view/addTaskPopUp.vue'
import editTaskPopUpComponent from './view/editTaskPopUp.vue'
import viewTaskAPIPopUpComponent from './view/viewTaskAPIPopUp.vue'
import tableDeleteButton from '@/views/common/tableDeleteButton.vue'
import { useRouter } from 'vue-router';
import { CONSTANT } from '@/constant'

export default {
  components: {
    addTaskPopUpComponent,
    editTaskPopUpComponent,
    tableDeleteButton,
    viewTaskAPIPopUpComponent,
    // InfoFilled,
  },
  setup() {
    const pageState = usePageState();
    const router = useRouter();

    const state = reactive({
      addTaskPopUpComponentName: 'addTaskPopUpComponentName',
      editTaskPopUpComponentName: 'editTaskPopUpComponentName',
      viewTaskAPIPopUpComponentName: 'viewTaskAPIPopUpComponentName',
      search: '',
      optionValue: '',
      dialogFormVisible: false,
      // confirmBtnVisible: false,
      tableData: [],
      total: CONSTANT.TOTAL,
      pageSize: CONSTANT.PAGE_SIZE,
      currPage: CONSTANT.PAGE,
      displayCols: [
        { 'col': 'id', 'label': '任务ID' },
        { 'col': 'name', 'label': '任务名' },
        { 'col': 'type', 'label': '发信任务' },
        { 'col': 'created_on', 'label': '创建时间' },
      ],
      options: [
        { label: '邮箱', value: 'Email' },
        { label: '钉钉', value: 'Dtalk' }
      ],
      sysKeptTaskIds: [CONSTANT.LOG_TASK_ID],

    });

    const handleEdit = (index, row) => {
      let name = state.editTaskPopUpComponentName;
      pageState.ShowDialogData[name] = {};
      pageState.ShowDialogData[name].isShow = true;
      pageState.ShowDialogData[name].rowData = row;
    }

    const handleDelete = async (index, row) => {
      const rsp = await request.post('/sendtasks/delete', { id: row.id });
      if (rsp.status == 200) {
        state.tableData.splice(index, 1);
      }
    }

    const handleViewAPI = (index, row) => {
      let name = state.viewTaskAPIPopUpComponentName;
      pageState.ShowDialogData[name] = {};
      pageState.ShowDialogData[name].isShow = true;
      pageState.ShowDialogData[name].rowData = row;
    }

    const handPageChange = async (pageNum) => {
      state.currPage = pageNum;
      await queryListData(pageNum, state.pageSize);

    }
    const rowStyle = () => {
      return {
        'font-size': '13px',
      }
    }

    const handleViewLogs = (index, row) => {
      router.push('/sendlogs?taskid=' + row.id, { replace: true });
    }

    const filterFunc = async () => {
      await queryListData(state.currPage, state.pageSize, state.search, state.optionValue);

    }
    const clickAdd = () => {
      pageState.ShowDialogData[state.addTaskPopUpComponentName] = {};
      pageState.ShowDialogData[state.addTaskPopUpComponentName].isShow = true;
    }

    const queryListData = async (page, size, name = '') => {
      let params = { page: page, size: size, name: name };
      const rsp = await request.get('/sendtasks/list', { params: params });
      state.tableData = await rsp.data.data.lists;
      state.total = await rsp.data.data.total;
    }

    onMounted(async () => {
      await queryListData(1, state.pageSize);
    });


    return {
      ...toRefs(state), handleEdit, handleDelete, handleViewAPI, handleViewLogs,
      clickAdd, rowStyle, handPageChange, filterFunc
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
  max-width: 1000px;
  width: 100%;
  margin-top: -10vh;
}

.pagination-block {
  margin-top: 30px;
  display: flex;
  justify-content: flex-end;
}

.total-tip {
  display: inline-block;
}

.search-box {
  float: right;
}

:global(.select .el-input__inner) {
  width: 80px;
}

.search-input-sendways {
  /* margin-left: 40px; */
  width: 150px;
  display: inline-flex;
}

.op-col {
  text-align: left;
}
</style>