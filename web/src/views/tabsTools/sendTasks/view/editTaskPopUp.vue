<template>
  <el-dialog v-model="isShow" width="58%" :close-on-press-escape="false" :before-close="() => { }" :show-close="false">
    <template #header="">
      <el-text class="mx-1">编辑发信任务</el-text>
      <el-tooltip placement="top">
        <template #content>
          实例可以实时暂停或者删除，意味着可以实时控制发送的渠道
          <br />
          ** 暂停或者删除，都将不会往该实例发送
        </template>
        <el-icon>
          <QuestionFilled />
        </el-icon>
      </el-tooltip>
    </template>

    <div class="add-top">
      <el-input v-model="currTaskInput.taskName" placeholder="请输入任务名" size="small" class="taskNameInput"></el-input>
      <el-button @click="handleEditTask()" size="small" type="primary" style="margin-left: 20px;">修改</el-button>
    </div>

    <div class="dashed" />

    <div class="ins-area">

      <div class="ins-add">
        <el-input v-model="searchWayID" placeholder="请输入要添加的渠道id" size="small" @change="searchID"
          class="searchInput"></el-input>
        <el-button @click="searchID()" size="small" type="primary" style="margin-left: 20px;">查询</el-button>

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
              style="width: 200px; margin: 10px 40px 5px 0;" class="searchInput">
            </el-input>

          </div>
          <div>
            <el-button @click="handleAddSubmit()" size="small" style="width: 200px">添加</el-button>
          </div>

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
          <el-table-column label="发送状态" prop="status" width="100px">
            <template #default="scope">
              <el-switch v-model.bool="scope.row.enable" inline-prompt active-text="开启发送" inactive-text="关闭发送"
                @click="updateInsEnableStatus(scope.row)" />
            </template>
          </el-table-column>
          <el-table-column fixed="right" label="操作" width="60px">
            <template #default="scope">
              <tableDeleteButton @customHandleDelete="handleDelete(scope.$index, scope.row)" />
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleCancer()" size="small">取消</el-button>
      </span>
    </template>

  </el-dialog>
</template>

<script>
import { defineComponent, onMounted, watch, reactive, toRefs, computed } from 'vue';
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
      isShow: false,
      isShowAddBox: false,
      searchWayID: '',
      currInsInputContentType: '',
      currWayTmp: {},
      currInsInput: {},
      currTaskInput: {
        taskName: '',
        taskId: '',
      },
    });

    watch(pageState.ShowDialogData, (newValue, oldValue) => {
      if (newValue[props.componentName]) {
        state.isShow = pageState.ShowDialogData[props.componentName].isShow;
        resetPageInitData();
        state.currTaskInput.taskId = pageState.ShowDialogData[props.componentName].rowData.id;
        if (state.isShow) {
          queryInsListData();
        }
      }
    });

    // 将实例的开启状态转换为布尔值，组件绑定需要布尔值
    const dealInsEnableStatus = (data) => {
      data.forEach(ins => {
        ins.enable = ins.enable == 1;
      });
    }

    const queryInsListData = async () => {
      let params = { id: state.currTaskInput.taskId };
      const rsp = await request.get('/sendtasks/ins/gettask', { params: params });
      state.insTableData = await rsp.data.data.ins_data;
      // state.total = await rsp.data.data.total;
      dealInsEnableStatus(state.insTableData);
      state.currTaskInput.taskName = await rsp.data.data.name;
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
      state.currWayTmp = {};
      state.currInsInput = {};
      state.searchWayID = '';
      state.currInsInputContentType = '';
      state.isShowAddBox = false;
      state.currTaskInput = {
        taskName: '',
        // taskId: uuidv4(),
      }
    }

    const insRowStyle = () => {
      return {
        'font-size': '12px'
      }
    }

    const handleDelete = async (index, row) => {
      const rsp = await request.post('/sendtasks/ins/delete', { id: row.id });
      if (rsp.status == 200) {
        state.insTableData.splice(index, 1);
      }
    }

    const searchID = async () => {
      const rsp = await request.get('/sendways/get', { params: { id: state.searchWayID } });
      let data = await rsp.data;
      state.isShowAddBox = Boolean(data.data);
      if (data.data) {
        state.currWayTmp = data.data;
      }
    }

    const getFinalData = () => {
      let postData = {
        id: generateBizUniqueID('I'),
        enable: 1,
        task_id: state.currTaskInput.taskId,
        way_id: state.currWayTmp.id,
        way_type: state.currWayTmp.type,
        way_name: state.currWayTmp.name,
        content_type: state.currInsInputContentType,
        config: JSON.stringify(state.currInsInput)
      }
      return postData
    }

    const handleAddSubmit = async () => {
      let postData = getFinalData();
      const rsp = await request.post('/sendtasks/ins/addone', postData);
      if (await rsp.data.code == 200) {
        postData.enable = true;
        state.insTableData.push(postData);
      }
    }

    const updateInsEnableStatus = async (row) => {
      let status = row.enable ? 1 : 0;
      let postData = { ins_id: row.id, status: status };
      const rsp = await request.post('/sendtasks/ins/update_enable', postData);
      if (await rsp.data.code == 200) {
        row.enable = Boolean(status);    //更新当前的状态值
        ElMessage({ message: await rsp.data.msg, type: 'success' });
      }
    }

    const handleEditTask = async () => {
      let postData = { id: state.currTaskInput.taskId, name: state.currTaskInput.taskName };
      const rsp = await request.post('/sendtasks/edit', postData);
      if (await rsp.data.code == 200) {
        ElMessage({ message: await rsp.data.msg, type: 'success' });
      }
    }

    return {
      ...toRefs(state), handleCancer, handleAddSubmit, handleEditTask, CONSTANT, CommonUtils,
      searchID, handleDelete, insRowStyle, updateInsEnableStatus
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
  /* max-width: 1000px; */

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

.searchInput {
  max-width: 200px;
}

.wayTitleInput {
  max-width: 200px;
}

.taskNameInput {
  /* width: 80%; */
  max-width: 200px;
}

.display-label {
  margin-top: 10px;
  margin-bottom: 10px;
}
</style>
