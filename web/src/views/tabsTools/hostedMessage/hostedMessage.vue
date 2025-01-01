<template>
  <div class="main-center-body">
    <div class="container">

      <div class="search-input-sendways">
        <el-input v-model="search" size="small" placeholder="根据关键字搜索相应托管消息" @change="filterFunc()" />
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
        <el-table :data="tableData" stripe empty-text="托管消息为空" :row-style="rowStyle()">
          <el-table-column label="ID" prop="id" width="85px" />
          <el-table-column label="消息标题" prop="title" show-overflow-tooltip width="150px" />
          <el-table-column label="托管消息" prop="content">
            <template #default="scope">
              <el-tooltip enterable placement="top">
                <template #content>
                  <div v-html="formatLogDisplayHtml(scope)"></div>
                </template>
                <span class="log-overflow">{{ scope.row.content }}</span>
              </el-tooltip>
            </template>
          </el-table-column>
          <el-table-column label="发送时间" prop="created_on" width="160px" />
          <el-table-column label="详情" prop="status" width="120px" fixed="right">
            <template #default="scope">
              <el-button link size="small" style="margin-right: 10px;" type="primary"
                @click="drawer = true; logText = formatLogDisplayHtml(scope)">查看消息</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <el-drawer v-model="drawer" :with-header="false">
        <el-text v-html="logText" size="small"></el-text>
      </el-drawer>

      <div class="pagination-block">
        <el-pagination layout="prev, pager, next" :total="total" :page-size="pageSize"
          @current-change="handPageChange" />
        <el-text class="total-tip" size="small">每页{{ pageSize }}条，共{{ total }}条</el-text>
      </div>

    </div>
  </div>
</template>

<script>
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

    const TransHtml = (raw) => {
      if (raw) {
        return raw.replace(/\n/g, '<br />')
      }
      return ''
    }

    const formatLogDisplayHtml = (scope) => {
      let content = TransHtml(scope.row.content);
      return content;
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

    const queryListData = async (page, size, name = '', query = '') => {
      let params = { page: page, size: size, text: name, query: query };
      const rsp = await request.get('/hostedmessages/list', { params: params });
      state.tableData = await rsp.data.data.lists;
      state.total = await rsp.data.data.total;
    }

    onMounted(async () => {
      state.search = router.query.name;
      await queryListData(1, state.pageSize, router.query.name, router.query.taskid, router.query.query);
    });

    return {
      ...toRefs(state), TransHtml, clickFreshSwitch,
      rowStyle, handPageChange, filterFunc, copyToClipboard, formatLogDisplayHtml
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
  margin-top: 15px;
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