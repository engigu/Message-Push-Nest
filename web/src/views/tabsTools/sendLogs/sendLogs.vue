<template>
  <div class="main-center-body">
    <div class="container">

      <div class="search-input-sendways">
        <el-input v-model="search" size="small" placeholder="根据任务名字搜索相应日志" @change="filterFunc()" />
      </div>
      <div class="search-box">
        <el-text class="refresh-time" v-if="refreshText" size="small">{{ refreshText }}</el-text>
        <el-input-number class="refresh-box" v-model="refreshSec" size="small" :step="2" :min="5"
          controls-position="right" />
        <el-switch v-model="refreshSwitch" class="ml-2" width="80" inline-prompt active-text="自动刷新" inactive-text="自动刷新"
          @click="clickFreshSwitch" />
      </div>

      <hr />
      <div ref="refContainer">
        <el-table :data="tableData" stripe empty-text="发信日志为空" :row-style="rowStyle()">
          <el-table-column label="ID" prop="id" width="50px" />
          <el-table-column label="任务名" prop="task_name" show-overflow-tooltip width="150px" />
          <el-table-column label="发信日志" prop="log">
            <template #default="scope">
              <el-tooltip enterable placement="top">
                <template #content>
                  <div v-html="TransHtml(scope.row.log)"></div>
                </template>
                <span class="log-overflow">{{ scope.row.log }}</span>
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column label="发送时间" prop="created_on" width="160px" />
          <el-table-column label="详情/状态" prop="status" width="120px" fixed="right">
            <template #default="scope">
              <el-button link size="small" style="margin-right: 10px;" type="primary"
                @click="drawer = true; logText = scope.row.log">日志</el-button>
              <el-tag v-if="scope.row.status == 0" type="danger">失败</el-tag>
              <el-tag v-if="scope.row.status == 1" type="success">成功</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <el-drawer v-model="drawer" :with-header="false">
        <el-text v-html="TransHtml(logText)" size="small"></el-text>
      </el-drawer>

      <div class="pagination-block">
        <el-pagination layout="prev, pager, next" :total="total" :page-size="pageSize" @current-change="handPageChange" />
        <el-text class="total-tip" size="small">每页{{ pageSize }}条，共{{ total }}条</el-text>
      </div>

    </div>
  </div>
</template>

<script >
import { reactive, toRefs, onMounted, watch, ref } from 'vue'
import { request } from '../../../api/api'
import { copyToClipboard } from '../../../util/clipboard.js';
import { useRoute } from 'vue-router';
import { CONSTANT } from '@/constant'
import { usePageState } from '@/store/page_sate.js';
import { CommonUtils } from "@/util/commonUtils.js";

export default {
  components: {
  },
  setup() {
    const pageState = usePageState();
    const router = useRoute();
    const state = reactive({
      search: '',
      refreshText: '',
      refreshSwitch: false,
      refreshSec: 20,
      refreshIntervalFuncList: [],
      optionValue: '',
      logText: '',
      drawer: false,
      tableData: [],
      total: CONSTANT.TOTAL,
      pageSize: pageState.siteConfigData.pagesize,
      currPage: CONSTANT.PAGE,
    });

    const handleDelete = async (index, row) => {
      const rsp = await request.post('/sendways/delete', { id: row.id });
      if (rsp.status == 200) {
        state.tableData.splice(index, 1);
      }
    }

    const TransHtml = (raw) => {
      if (raw) {
        return raw.replace(/\n/g, '<br />')
      }
      return ''
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

    const clickFreshSwitch = () => {
      if (state.refreshSwitch) {
        let flag = setInterval(async function () {
          await filterFunc();
          state.refreshText = `自动刷新于：${CommonUtils.getCurrentTimeStr()}`;
        }, state.refreshSec * 1000);
        state.refreshIntervalFuncList.push(flag);
      } else {
        state.refreshIntervalFuncList.forEach(intervalId => {
          clearInterval(intervalId);
        });
        state.refreshIntervalFuncList = [];
        state.refreshText = '';
      }
    }


    const filterFunc = async () => {
      await queryListData(state.currPage, state.pageSize, state.search, state.optionValue);

    }

    const queryListData = async (page, size, name = '', taskid = '') => {
      let params = { page: page, size: size, name: name, taskid: taskid };
      const rsp = await request.get('/sendlogs/list', { params: params });
      state.tableData = await rsp.data.data.lists;
      state.total = await rsp.data.data.total;
    }

    onMounted(async () => {
      state.search = router.query.name;
      await queryListData(1, state.pageSize, router.query.name, router.query.taskid);
    });

    return {
      ...toRefs(state), handleDelete, TransHtml, clickFreshSwitch,
      rowStyle, handPageChange, filterFunc, copyToClipboard
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

.search-box {
  float: right;
  margin-top: -2px;
}

.refresh-box {
  width: 80px;
  margin-right: 5px;
}

.refresh-time {
  margin-right: 10px;
}

.pagination-block {
  margin-top: 30px;
  display: flex;
  justify-content: flex-end;

}

.total-tip {
  display: inline-block;
}

.search-input-sendways {
  width: 200px;
  display: inline-flex;
}


.log-overflow {
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
}
</style>