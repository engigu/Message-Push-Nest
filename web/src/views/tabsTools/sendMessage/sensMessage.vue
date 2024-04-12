<template>
  <div class="main-center-body">
    <div class="container">

      <div class="search-input-sendways">
        <el-input v-model="search" size="small" placeholder="搜索" @change="filterFunc()" />
      </div>
      <div class="search-box">
        <el-button size="small" type="primary" @click="clickAdd">新增定时消息</el-button>
      </div>

      <hr />

      <div ref="refContainer">
        <el-table :data="tableData" stripe empty-text="定时发信为空" :row-style="rowStyle()">
          <el-table-column label="消息ID" prop="id"/>
          <el-table-column label="关联ID" prop="task_id" />
          <el-table-column label="消息名" prop="name"  show-overflow-tooltip/>
          <el-table-column label="Crontab" prop="cron" />
          <el-table-column label="下次执行" prop="next_time" show-overflow-tooltip/>
          <!-- <el-table-column label="Crontab">
            <template #default="scope">
          {{ scope.row.cron }}
          <br/>
          <el-text class="mx-1" type="primary" size="small"> next: {{ scope.row.next_time }}</el-text>
         
            </template>
          </el-table-column> -->
          <el-table-column label="消息内容" prop="content" show-overflow-tooltip />
          <el-table-column label="创建时间" prop="created_on" show-overflow-tooltip/>
          <el-table-column fixed="right" label="操作" width="190px">
            <template #default="scope">

              <div>
                <el-button link size="small" type="primary"
                  @click="handleViewLogs(scope.$index, scope.row)">日志</el-button>
                <el-button link size="small" type="primary" @click="handleEdit(scope.$index, scope.row)">编辑</el-button>
                <tableDeleteButton @customHandleDelete="handleDelete(scope.$index, scope.row)" />

                <el-switch v-model.bool="scope.row.enable" inline-prompt active-text="开启" inactive-text="关闭"
                  @click="updateEnableStatus(scope.row)" />
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="pagination-block">
        <el-pagination layout="prev, pager, next" :total="total" :page-size="pageSize"
          @current-change="handPageChange" />
        <el-text class="total-tip" size="small">每页{{ pageSize }}条，共{{ total }}条</el-text>
      </div>

    </div>
  </div>

  <addCronMsgPopUpComponent :componentName="addCronMsgPopUpComponentName" />
  <editCronMsgPopUpComponent :componentName="editCronMsgPopUpComponentName" />
</template>

<script>
import { computed, ref, reactive, toRefs, onMounted } from 'vue'
import { usePageState } from '@/store/page_sate';
import { request } from '@/api/api'
import addCronMsgPopUpComponent from './view/addCronMsgPopUp.vue'
import editCronMsgPopUpComponent from './view/editCronMsgPopUp.vue'
import tableDeleteButton from '@/views/common/tableDeleteButton.vue'
import { useRouter } from 'vue-router';
import { CONSTANT } from '@/constant'
import { ElMessage } from 'element-plus'

export default {
  components: {
    addCronMsgPopUpComponent,
    editCronMsgPopUpComponent,
    tableDeleteButton,
  },
  setup() {
    const pageState = usePageState();
    const router = useRouter();

    const state = reactive({
      addCronMsgPopUpComponentName: 'addCronMsgPopUpComponentName',
      editCronMsgPopUpComponentName: 'editCronMsgPopUpComponentName',
      search: '',
      optionValue: '',
      dialogFormVisible: false,
      tableData: [],
      total: CONSTANT.TOTAL,
      pageSize: pageState.siteConfigData.pagesize,
      currPage: CONSTANT.PAGE,
    });

    const handleEdit = (index, row) => {
      let name = state.editCronMsgPopUpComponentName;
      pageState.ShowDialogData[name] = {};
      pageState.ShowDialogData[name].isShow = true;
      pageState.ShowDialogData[name].rowData = row;
    }

    const handleDelete = async (index, row) => {
      const rsp = await request.post('/sendmessages/delete', { id: row.id });
      if (rsp.status == 200) {
        state.tableData.splice(index, 1);
      }
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
      router.push('/sendlogs?taskid=' + row.task_id, { replace: true });
    }

    const filterFunc = async () => {
      await queryListData(state.currPage, state.pageSize, state.search, state.optionValue);
    }

    const clickAdd = () => {
      pageState.ShowDialogData[state.addCronMsgPopUpComponentName] = {};
      pageState.ShowDialogData[state.addCronMsgPopUpComponentName].isShow = true;
    }

    const queryListData = async (page, size, name = '') => {
      let params = { page: page, size: size, name: name };
      const rsp = await request.get('/sendmessages/list', { params: params });
      state.tableData = await rsp.data.data.lists;
      dealInsEnableStatus(state.tableData)
      state.total = await rsp.data.data.total;
    }

    const dealInsEnableStatus = (data) => {
      data.forEach(ins => {
        ins.enable = ins.enable == 1;
      });
    }

    // 开启暂停定时任务
    const updateEnableStatus = async (data) => {
      data.enable = !Boolean(data.enable) ? 0 : 1
      const rsp = await request.post('/sendmessages/edit', data);
      if (await rsp.data.code == 200) {
        ElMessage({ message: await rsp.data.msg, type: 'success' });
        data.enable = Boolean(data.enable)
      }
    }

    onMounted(async () => {
      await queryListData(1, state.pageSize);
    });

    return {
      ...toRefs(state), handleEdit, handleDelete, handleViewLogs,
      clickAdd, rowStyle, handPageChange, filterFunc, updateEnableStatus
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
  /* margin-top: -10vh; */
}

.pagination-block {
  margin-top: 15px;
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