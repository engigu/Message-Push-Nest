<template>
  <div class="main-center-body">
    <div class="container">
      <el-row :gutter="18">
        <el-col :span="4">
          <div class="statistic-card">
            <el-statistic :value="data.message_total_num">
              <template #title>
                <div style="display: inline-flex; align-items: center">
                  当前消息留存数
                </div>
              </template>
            </el-statistic>
          </div>
        </el-col>
        <el-col :span="5">
          <div class="statistic-card">
            <el-statistic :value="data.hosted_message_total_num">
              <template #title>
                <div style="display: inline-flex; align-items: center">
                  托管消息数
                </div>
              </template>
            </el-statistic>
          </div>
        </el-col>
        <el-col :span="5">
          <div class="statistic-card">
            <el-statistic :value="data.today_total_num">
              <template #title>
                <div style="display: inline-flex; align-items: center">
                  今日发送总数
                </div>
              </template>
            </el-statistic>
          </div>
        </el-col>
        <el-col :span="5">
          <div class="statistic-card">
            <el-statistic :value="data.today_succ_num">
              <template #title>
                <div style="display: inline-flex; align-items: center">
                  今日发送成功数
                </div>
              </template>
            </el-statistic>
          </div>
        </el-col>
        <el-col :span="5">
          <div class="statistic-card">
            <el-statistic :value="data.today_failed_num" :value-style="formatFailedNumStyle()">
              <template #title>
                <div style="display: inline-flex; align-items: center">
                  今日发送失败数
                </div>
              </template>
              <template #suffix>
                <el-icon :size="14">
                  <ArrowRight v-if="data.today_failed_num > 0" @click="clickErrorNumDetail()" />
                </el-icon>
              </template>
            </el-statistic>
          </div>
        </el-col>
      </el-row>

      <el-divider />

      <div class="echarts-box">
        <div ref="dailyChart" style="width: 65%; height: 350px;"></div>
        <div ref="sendCateChart" style="width: 35%; height: 320px;"></div>
      </div>

    </div>
  </div>
</template>

<script>
import {
  ArrowRight,
  CaretBottom,
  CaretTop,
  Warning,
} from '@element-plus/icons-vue'
import { reactive, toRefs, onMounted, onUnmounted, inject, ref } from 'vue'
import { request } from '@/api/api'
import { useRouter } from 'vue-router';
import { CommonUtils } from "@/util/commonUtils.js";

export default {
  components: {
    ArrowRight,
  },
  setup() {
    const dailyChart = ref(null);
    const sendCateChart = ref(null);
    const router = useRouter();
    const state = reactive({
      data: {},
      displayCharts: [],
    });

    const echart = inject('$echarts');

    const getStatisticData = async () => {
      const rsp = await request.get('/statistic');
      if (await rsp.data.code == 200) {
        let data = await rsp.data.data;
        state.data = data;
      }
    }

    // 格式化发送失败的数字样式
    const formatFailedNumStyle = () => {
      let style = {};
      if (state.data.today_failed_num) {
        style['color'] = 'red';
      }
      return style;
    }

    const clickErrorNumDetail = () => {
      let query = { status: 0, day_created_on: CommonUtils.getCurrentTimeStr().slice(0, 10) };
      let queryStr = encodeURIComponent(JSON.stringify(query));
      router.push('/sendlogs?query=' + queryStr, { replace: true });

    }

    onMounted(async () => {
      await getStatisticData();
      formatFailedNumStyle();
      initDailyChart();
      initSendCateChart();
    });

    onUnmounted(() => {
      state.displayCharts.forEach(element => {
        element.dispose();
      });
    });

    const getLasteDetailData = (key) => {
      let result = [];
      state.data.latest_send_data.forEach(element => {
        result.push(element[key]);
      })
      return result;
    }

    // 最近30天数据图
    function initDailyChart() {
      let chartobj = echart.init(dailyChart.value);
      state.displayCharts.push(chartobj);

      let xAxisdata = [];
      state.data.latest_send_data.forEach(element => {
        xAxisdata.push(element.day);
      });

      chartobj.setOption({
        tooltip: {
          trigger: 'axis'
        },
        title: {
          subtext: '最近30天发送消息数据',
          top: 0,
          textStyle: {
            color: '#333',
            fontSize: 18,
          },
        },
        xAxis: {
          type: 'category',
          boundaryGap: false,
          data: xAxisdata
        },
        tooltip: {
          trigger: "axis"
        },
        yAxis: {
          type: "value"
        },
        // legend: {
        //   data: ['发送数', '成功数', '失败数']
        // },
        series: [
          {
            name: '发送数',
            data: getLasteDetailData('num'),
            type: "line",
            smooth: true
          },
          {
            name: '成功数',
            data: getLasteDetailData('day_succ_num'),
            type: "line",
            smooth: true
          },
          {
            name: '失败数',
            data: getLasteDetailData('day_failed_num'),
            type: "line",
            smooth: true
          },
        ]
      });

    }

    // 发送消息实例类别图
    function initSendCateChart() {
      let chartobj = echart.init(sendCateChart.value);
      state.displayCharts.push(chartobj);

      let data = [];
      state.data.way_cate_data.forEach(element => {
        data.push({ name: element.way_name, value: element.count_num });
      });
      chartobj.setOption({
        tooltip: {
          trigger: 'item'
        },
        grid: {
          width: '60%',
          height: '60%',
        },
        title: {
          subtext: '发送消息实例渠道',
          top: 0,
          textStyle: {
            color: '#333',
            fontSize: 18,
          },
        },
        series: [
          {
            type: 'pie',
            data: data,
            radius: '40%',
            // roseType: 'area'
          }
        ]
      });
    }

    return {
      ...toRefs(state), formatFailedNumStyle, clickErrorNumDetail, dailyChart, sendCateChart
    };
  }
}
</script>

<style scoped>
.container {
  max-width: 1000px;
  height: 450px;
}


.el-statistic {
  --el-statistic-content-font-size: 28px;
  text-align: center;
}

.statistic-card {
  height: 100%;
  padding: 20px;
  border-radius: 4px;
  background-color: var(--el-bg-color-overlay);
}

.echarts-box {
  display: flex;
}

.green {
  color: var(--el-color-success);
}

.red {
  color: var(--el-color-error);
}
</style>