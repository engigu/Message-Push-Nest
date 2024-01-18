<template>
  <div class="main-center-body">
    <div class="container">
      <el-select v-model="optionValue" class="select" placeholder="渠道筛选" size="small" @change="filterFunc()">
        <el-option v-for="item in options" :label="item.label" :value="item.value" />
      </el-select>

      <div class="search-input-sendways">
        <el-input v-model="search" size="small" placeholder="搜索" @change="filterFunc()" />
      </div>

      <!-- <td class="line">
          <div />
        </td> -->
      <div class="search-box">
        <el-button size="small" type="primary" @click="clickAdd">新增渠道</el-button>
      </div>

      <hr />
      <div ref="refContainer">
        <el-table :data="tableData" stripe empty-text="发信渠道为空" :row-style="rowStyle()">
          <el-table-column label="ID" width="320px">
            <template #default="scope">
              {{ scope.row.id }}
              <el-icon>
                <CopyDocument @click="copyToClipboard(scope.row.id)" />
              </el-icon>
            </template>
          </el-table-column>
          <el-table-column label="渠道名" prop="name" />
          <el-table-column label="发信渠道" prop="type" />
          <el-table-column label="创建时间" prop="created_on" />
          <el-table-column fixed="right" label="操作" width="100px">
            <template #default="scope">
              <!-- <el-button link size="small" type="primary" @click="handleView(scope.$index, scope.row)">查看</el-button> -->
              <el-button link size="small" type="primary" @click="handleEdit(scope.$index, scope.row)">编辑</el-button>
              <tableDeleteButton @customHandleDelete="handleDelete(scope.$index, scope.row)" />
            </template>
          </el-table-column>

        </el-table>

      </div>
      <div class="pagination-block">
        <el-pagination layout="prev, pager, next" :total="total" :page-size="pageSize" @current-change="handPageChange" />
        <el-text class="total-tip" size="small">每页{{ pageSize }}条，共{{ total }}条</el-text>
      </div>

      <addWayComponent :componentName="addWayComponentName" />
      <editWayComponent :componentName="editWayComponentName" />

    </div>
  </div>
</template>

<script >
import { reactive, toRefs, onMounted } from 'vue'
import addWayComponent from './view/addWayPopUp.vue'
import editWayComponent from './view/editWayPopUp.vue'
import { usePageState } from '@/store/page_sate.js';
import { request } from '@/api/api'
import { CopyDocument } from '@element-plus/icons-vue'
import { copyToClipboard } from '@/util/clipboard.js';
import tableDeleteButton from '@/views/common/tableDeleteButton.vue'
import { CONSTANT } from '@/constant'

export default {
  components: {
    addWayComponent,
    editWayComponent,
    CopyDocument,
    tableDeleteButton,
  },
  setup() {
    const pageState = usePageState();
    const state = reactive({
      addWayComponentName: 'addWayComponent',
      editWayComponentName: 'editWayComponent',
      search: '',
      optionValue: '',
      dialogFormVisible: false,
      // confirmBtnVisible: false,
      tableData: [],
      total: CONSTANT.TOTAL,
      pageSize: pageState.siteConfigData.pagesize,
      currPage: CONSTANT.PAGE,
      displayCols: [
        { 'col': 'id', 'label': '渠道ID' },
        { 'col': 'name', 'label': '渠道名' },
        { 'col': 'type', 'label': '发信渠道' },
        { 'col': 'created_on', 'label': '创建时间' },
      ],
      options: [
        { label: '邮箱', value: 'Email' },
        { label: '钉钉', value: 'Dtalk' }
      ]
    });

    const handleEdit = (index, row) => {
      let name = state.editWayComponentName;
      pageState.ShowDialogData[name] = {};
      pageState.ShowDialogData[name].isShow = true;
      pageState.ShowDialogData[name].rowData = row;
    }

    const handleDelete = async (index, row) => {
      const rsp = await request.post('/sendways/delete', { id: row.id });
      if (rsp.status == 200 && await rsp.data.code == 200) {
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

    const filterFunc = async () => {
      await queryListData(state.currPage, state.pageSize, state.search, state.optionValue);

    }
    const clickAdd = () => {
      pageState.ShowDialogData[state.addWayComponentName] = {};
      pageState.ShowDialogData[state.addWayComponentName].isShow = true;
    }

    const queryListData = async (page, size, name = '', type = '') => {
      let params = { page: page, size: size, name: name, type: type };
      const rsp = await request.get('/sendways/list', { params: params });
      state.tableData = await rsp.data.data.lists;
      state.total = await rsp.data.data.total;
    }

    onMounted(async () => {
      await queryListData(1, state.pageSize);
    });

    return {
      ...toRefs(state), handleEdit, handleDelete,
      clickAdd, rowStyle, handPageChange, filterFunc, copyToClipboard
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

.el-table th,
.el-table .cell {
  font-size: 10px;
  /* 调整字体大小 */
}

/* .search-input {
  margin-right: 0px;
  margin-left: 0px;
  padding-left: 0px;
} */

.el-dialog__title {
  font-size: 14px !important;
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
  margin-left: 40px;
  width: 150px;
  display: inline-flex;
}

.op-col {
  text-align: left;
}
</style>