<template>
  <el-dialog v-model="isShow" width="58%" :close-on-press-escape="false" :before-close="() => { }" :show-close="false">
    <template #header="">
      <el-text class="mx-1">添加发信任务(绑定关联实例)</el-text>
      <el-tooltip placement="top">
        <template #content>
          一个任务可以关联创建多个实例
          <br />
          选择不同的渠道，填写的实例信息也不一样
          <br />
          一个任务可以绑定一个实例，也可以绑定多个实例，多个实例意味着一个消息可以推送给多个消息渠道
        </template>
        <el-icon>
          <QuestionFilled />
        </el-icon>
      </el-tooltip>
    </template>

    <div class="add-top">
      <el-input v-model="currTaskInput.taskName" placeholder="请输入任务名" size="small" class="taskNameInput"></el-input>
    </div>

    <div class="dashed" />

    <div class="ins-area">

      <div class="ins-add">
        <el-autocomplete v-model="currSearchInputText" size="small" :fetch-suggestions="querySearchWayAsync"
          placeholder="请输入渠道名进行搜索" @select="handleSearchSelect" :clearable="true" value-key="name" />

        <div class="store-area" v-if="isShowAddBox">

          <div class="display-label">
            <el-text class="mx-1" size="small">渠道名：{{ currWayTmp.name }}</el-text><br />
            <el-text class="mx-1" size="small">渠道类型：{{ currWayTmp.type }}</el-text> <br />
            <el-text class="mx-1" size="small">渠道创建时间：{{ currWayTmp.created_on }}</el-text><br />
          </div>

          <el-radio-group v-model="currInsInputContentType">
            <el-radio v-for="item in CONSTANT.WAYS_DATA_MAP[currWayTmp.type].taskInsRadios" :label="item.subLabel"
              size="small">
              {{ item.content }}
            </el-radio>
          </el-radio-group>

          <div>
            <el-input v-model="currInsInput[item.col]"
              v-for="item in CONSTANT.WAYS_DATA_MAP[currWayTmp.type].taskInsInputs" :placeholder="item.desc" size="small"
              style="width: 200px; margin: 10px 40px 5px 0;">
            </el-input>
          </div>

          <el-button @click="clickStore()" size="small" style="width: 200px">暂存</el-button>
        </div>
      </div>

      <div class="ins-table">
        <el-table :data="insTableData" empty-text="发信实例为空" style="width: 100%" max-height="300"
          :row-style="insRowStyle()">
          <el-table-column prop="way_name" label="渠道名" />
          <el-table-column prop="way_type" label="渠道+内容类型" width="140px">
            <template #default="scope">
              {{ CommonUtils.formatWayName(scope.row.way_type) }}+{{ scope.row.content_type }}
            </template>
          </el-table-column>
          <el-table-column prop="way_type" label="额外信息">
            <template #default="scope">
              {{ CommonUtils.formatInsConfigDisplay(scope) }}
            </template>
          </el-table-column>
          <el-table-column fixed="right" label="操作" width="60px">
            <template #default="scope">
              <el-button link type="primary" size="small" @click.prevent="insTableData.splice(scope.$index, 1)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleCancer()" size="small">取消</el-button>
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
      searchWayID: '',
      currWayTmp: {},
      currInsInput: {},
      currInsInputContentType: 'text',
      currTaskInput: {
        taskName: '',
        taskId: generateBizUniqueID('T'),
      },
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
      state.currWayTmp = {};
      state.searchWayID = '';
      state.isShowAddBox = false;
      state.currTaskInput = {
        taskName: '',
        taskId: generateBizUniqueID('T'),
      }
    }

    const insRowStyle = () => {
      return {
        'font-size': '12px'
      }
    }

    // 点击暂存实例
    const clickStore = () => {
      let insData = {
        id: generateBizUniqueID('I'),
        task_id: state.currTaskInput.taskId,
        way_id: state.currWayTmp.id,
        way_type: state.currWayTmp.type,
        way_name: state.currWayTmp.name,
        content_type: state.currInsInputContentType,
        config: JSON.stringify(state.currInsInput)
      };
      state.insTableData.push(insData);
    }

    // 匹配出当前搜索的渠道数据
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
      let currWay = matchSearchData(state.currSearchInputText)
      state.isShowAddBox = Boolean(currWay);
      if (currWay) {
        state.currWayTmp = currWay;
        // 初始化currInsInput
        CONSTANT.WAYS_DATA_MAP[currWay.type].taskInsInputs.forEach(element => {
          state.currInsInput[element.col] = ""
        });
      }
    }

    const querySearchWayAsync = async (query, cb) => {
      let params = { name: query };
      const rsp = await request.get('/sendways/list', { params: params });
      let tableData = await rsp.data.data.lists;
      cb(tableData);
      state.currSearchWaysData = tableData;
    }

    const getFinalData = () => {
      let postData = { id: state.currTaskInput.taskId, name: state.currTaskInput.taskName }
      postData.ins_data = state.insTableData
      return postData
    }

    const handleSubmit = async () => {
      let postData = getFinalData();
      const rsp = await request.post('/sendtasks/ins/addmany', postData);
      if (await rsp.data.code == 200) {
        handleCancer();
        window.location.reload();
      }
    }

    return {
      ...toRefs(state), handleCancer, handleSubmit, querySearchWayAsync,
      handleSearchSelect, CONSTANT, CommonUtils,
      clickStore, insRowStyle
    };
  },
});
</script>

<style scoped>
::v-deep(.el-table .cell) {
  white-space: pre-line !important;
}

.dashed {
  border-top: 2px dashed var(--el-border-color);
  margin-bottom: 20px;
}

.ins-area {
  display: flex;
  /* max-width: 900px; */
}

.add-top {
  margin-bottom: 20px;
}

.ins-add {
  flex: 45%;
}

.ins-table {
  flex: 55%;
}

:deep(.el-autocomplete) {
  width: 200px;
}


.taskNameInput {
  max-width: 200px;
}

.display-label {
  margin-top: 10px;
  margin-bottom: 10px;
}
</style>
