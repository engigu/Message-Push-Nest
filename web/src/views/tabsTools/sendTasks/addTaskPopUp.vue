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
        <el-input v-model="searchWayID" placeholder="请输入要添加的渠道id" size="small" @change="searchID"
          class="searchInput"></el-input>
        <el-button @click="searchID()" size="small" type="primary" style="margin-left: 20px;">查询</el-button>

        <div class="store-area" v-if="isShowAddBox">

          <div class="display-label">
            <el-text class="mx-1" size="small">渠道名：{{ currWayTmp.name }}</el-text><br />
            <el-text class="mx-1" size="small">渠道类型：{{ currWayTmp.type }}</el-text> <br />
            <el-text class="mx-1" size="small">渠道创建时间：{{ currWayTmp.created_on }}</el-text><br />
          </div>

          <el-radio-group v-model="currInsInput.content_type" class="ml-4">
            <el-radio label="text" size="small">text</el-radio>
            <el-radio label="html" size="small">html</el-radio>
          </el-radio-group>

          <div>
            <el-input v-model="currInsInput.toAccount" placeholder="目的邮箱账号（发给谁）" size="small"
              style="width: 200px; margin: 10px 40px 5px 0;" class="searchInput"></el-input>
            <el-input v-model="currInsInput.title" placeholder="邮箱标题" size="small"
              style="width: 200px; margin: 0px 40px 5px 0;" class="searchInput"></el-input>
            <el-button @click="clickStore()" size="small" style="width: 200px">暂存</el-button>
          </div>

        </div>
      </div>

      <div class="ins-table">
        <el-table :data="insTableData" empty-text="发信实例为空" style="width: 100%" max-height="300"
          :row-style="insRowStyle()">
          <el-table-column prop="way_name" label="渠道名" />
          <el-table-column prop="way_type" label="渠道+内容类型">
            <template #default="scope">
              {{ scope.row.way_type }}+{{ scope.row.content_type }}
            </template>
          </el-table-column>
          <el-table-column prop="way_type" label="额外信息">
            <template #default="scope">
              {{ formatExtraInfo(scope) }}
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
import { v4 as uuidv4 } from 'uuid';
import { usePageState } from '../../../store/page_sate.js';
import { request } from '../../../api/api'


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
      isShow: false,
      isShowAddBox: false,
      searchWayID: '',
      currWayTmp: {},
      currInsInput: {
        content_type: 'text'
      },
      currTaskInput: {
        taskName: '',
        taskId: uuidv4(),
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
      state.currInsInput = { content_type: 'text' };
      state.currWayTmp = {};
      state.searchWayID = '';
      state.isShowAddBox = false;
      state.currTaskInput = {
        taskName: '',
        taskId: uuidv4(),
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
        id: uuidv4(),
        task_id: state.currTaskInput.taskId,
        way_id: state.currWayTmp.id,
        way_type: state.currWayTmp.type,
        way_name: state.currWayTmp.name,
        content_type: state.currInsInput.content_type,
        config: JSON.stringify({ to_account: state.currInsInput.toAccount, title: state.currInsInput.title })
      };
      state.insTableData.push(insData);
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

    const formatExtraInfo = (scope) => {
      if (!scope.row.config) {
        return ""
      }
      let config = JSON.parse(scope.row.config)
      let info = `发送账号：${config.to_account}\n标题：${config.title}`
      return info
    }

    return {
      ...toRefs(state), handleCancer, handleSubmit,
      searchID, formatExtraInfo,
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
  max-width: 900px;

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
