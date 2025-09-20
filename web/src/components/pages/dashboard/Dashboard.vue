<script setup lang="ts">
import StatCard from '@/components/pages/dashboard/CardNum.vue'
import { DatabaseIcon, BarChartIcon, SendIcon, CheckCircleIcon, XCircleIcon } from 'lucide-vue-next'
// import { LineChart } from "@/components/ui/chart-line"
import { onMounted, reactive } from 'vue';
import { request } from '@/api/api';
import { toast } from 'vue-sonner';
// import VueApexCharts from 'vue3-apexcharts'
import ApexCharts from 'apexcharts'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface SendData {
  day: string
  day_failed_num: number
  day_succ_num: number
  num: number
  succ_num: number
}
interface CateData {
  count_num: number
  way_name: string
}

let state = reactive({
  basicData: {
    message_total_num: 0,
    hosted_message_total_num: 0,
    today_total_num: 0,
    today_succ_num: 0,
    today_failed_num: 0,
  },
  trendData: {
    latest_send_data: [] as SendData[],
  },
  channelData: {
    way_cate_data: [] as CateData[],
  },
  loading: {
    basic: false,
    trend: false,
    channels: false,
  }
});

// 获取基础统计数据
const getBasicStatisticData = async () => {
  state.loading.basic = true;
  try {
    const rsp = await request.get('/statistic?type=basic');
    if (rsp && rsp.data && rsp.data.code == 200) {
      state.basicData = rsp.data.data;
    } else {
      toast.error(rsp?.data?.msg || '获取基础统计数据失败');
    }
  } catch (error) {
    toast.error('获取基础统计数据时发生错误');
  } finally {
    state.loading.basic = false;
  }
}

// 获取趋势统计数据
const getTrendStatisticData = async () => {
  state.loading.trend = true;
  try {
    const rsp = await request.get('/statistic?type=trend');
    if (rsp && rsp.data && rsp.data.code == 200) {
      state.trendData = rsp.data.data;
      // 数据加载完成后重新渲染折线图
      setTimeout(() => {
        renderLineChart();
      }, 100);
    } else {
      toast.error(rsp?.data?.msg || '获取趋势统计数据失败');
    }
  } catch (error) {
    toast.error('获取趋势统计数据时发生错误');
  } finally {
    state.loading.trend = false;
  }
}

// 获取渠道统计数据
const getChannelStatisticData = async () => {
  state.loading.channels = true;
  try {
    const rsp = await request.get('/statistic?type=channels');
    if (rsp && rsp.data && rsp.data.code == 200) {
      state.channelData = rsp.data.data;
      // 数据加载完成后重新渲染饼图
      setTimeout(() => {
        renderPieChart();
      }, 100);
    } else {
      toast.error(rsp?.data?.msg || '获取渠道统计数据失败');
    }
  } catch (error) {
    toast.error('获取渠道统计数据时发生错误');
  } finally {
    state.loading.channels = false;
  }
}

// 并行加载所有统计数据
const loadAllStatisticData = async () => {
  await Promise.all([
    getBasicStatisticData(),
    getTrendStatisticData(),
    getChannelStatisticData()
  ]);
}

const renderLineChart = () => {
  const options = {
    series: [
      {
        name: '发送总数',
        data: state.trendData.latest_send_data.length > 0
          ? state.trendData.latest_send_data.map(item => item.num || 0)
          : []
      },
      {
        name: '发送成功数',
        data: state.trendData.latest_send_data.length > 0
          ? state.trendData.latest_send_data.map(item => item.day_succ_num || 0)
          : []
      },
      {
        name: '发送失败数',
        data: state.trendData.latest_send_data.length > 0
          ? state.trendData.latest_send_data.map(item => item.day_failed_num || 0)
          : []
      },
    ],
    chart: {
      type: 'line',
      height: 350,
      toolbar: { show: false },
      background: 'transparent',
      animations: {
        enabled: true,
        easing: 'easeinout',
        speed: 800,
        animateGradually: {
          enabled: true,
          delay: 150
        },
        dynamicAnimation: {
          enabled: true,
          speed: 350
        }
      },
      dropShadow: {
        enabled: true,
        color: '#000',
        top: 18,
        left: 7,
        blur: 10,
        opacity: 0.2
      }
    },
    stroke: {
      curve: 'smooth',
      width: 3,
      lineCap: 'round'
    },
    markers: {
      size: 6,
      colors: ['#3b82f6', '#10b981', '#ef4444'],
      strokeColors: '#fff',
      strokeWidth: 2,
      hover: {
        size: 8,
        sizeOffset: 3
      }
    },
    xaxis: {
      categories: state.trendData.latest_send_data.length > 0
        ? state.trendData.latest_send_data.map(item => item.day)
        : [],
      axisBorder: {
        show: false
      },
      axisTicks: {
        show: false
      },
      labels: {
        style: {
          colors: '#64748b',
          fontSize: '12px',
          fontFamily: 'Inter, sans-serif'
        }
      }
    },
    yaxis: {
      labels: {
        style: {
          colors: '#64748b',
          fontSize: '12px',
          fontFamily: 'Inter, sans-serif'
        },
        formatter: function (val: number) {
          return val + ' 条'
        }
      }
    },
    colors: ['#3b82f6', '#10b981', '#ef4444'], // 蓝色表示总数，绿色表示成功，红色表示失败
    fill: {
      type: 'gradient',
      gradient: {
        shade: 'light',
        type: 'vertical',
        shadeIntensity: 0.5,
        gradientToColors: ['#60a5fa', '#34d399', '#f87171'],
        inverseColors: false,
        opacityFrom: 0.8,
        opacityTo: 0.1,
        stops: [0, 100]
      }
    },
    grid: {
      borderColor: '#e2e8f0',
      strokeDashArray: 3,
      xaxis: {
        lines: {
          show: false
        }
      },
      yaxis: {
        lines: {
          show: true
        }
      },
      padding: {
        top: 0,
        right: 0,
        bottom: 0,
        left: 0
      }
    },
    legend: {
      position: 'top',
      horizontalAlign: 'right',
      floating: true,
      offsetY: -25,
      offsetX: -5,
      fontSize: '12px',
      fontFamily: 'Inter, sans-serif',
      markers: {
        width: 8,
        height: 8,
        radius: 4
      }
    },
    tooltip: {
      enabled: true,
      shared: true,
      intersect: false,
      theme: 'light',
      style: {
        fontSize: '12px',
        fontFamily: 'Inter, sans-serif'
      },
      x: {
        show: true,
        format: 'MM/dd'
      },
      y: {
        formatter: function (val: number, { seriesIndex: _seriesIndex }: { seriesIndex: number }) {
          return val + ' 条'
        }
      },
      marker: {
        show: true
      },
      custom: function ({ series, seriesIndex: _seriesIndex, dataPointIndex, w }: { series: number[][], seriesIndex: number, dataPointIndex: number, w: any }) {
        const successCount = series[1][dataPointIndex];
        const failedCount = series[2][dataPointIndex];
        const total = successCount + failedCount;
        const successRate = total > 0 ? ((successCount / total) * 100).toFixed(1) : '0.0';

        return `
          <div class="bg-white p-3 rounded-lg shadow-lg border">
            <div class="font-medium text-gray-900 mb-2">${w.globals.categoryLabels[dataPointIndex]}</div>
            <div class="space-y-1">
              <div class="flex items-center justify-between">
                <span class="flex items-center">
                  <span class="w-2 h-2 bg-green-500 rounded-full mr-2"></span>
                  <span class="text-sm text-gray-600">成功:</span>
                </span>
                <span class="text-sm font-medium text-gray-900">${successCount} 条</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="flex items-center">
                  <span class="w-2 h-2 bg-red-500 rounded-full mr-2"></span>
                  <span class="text-sm text-gray-600">失败:</span>
                </span>
                <span class="text-sm font-medium text-gray-900">${failedCount} 条</span>
              </div>
              <div class="border-t pt-1 mt-2">
                <div class="flex items-center justify-between">
                  <span class="text-sm text-gray-600">成功率:</span>
                  <span class="text-sm font-medium text-green-600">${successRate}%</span>
                </div>
              </div>
            </div>
          </div>
        `;
      }
    },
    responsive: [{
      breakpoint: 768,
      options: {
        chart: {
          height: 300
        },
        legend: {
          position: 'bottom',
          offsetY: 0
        }
      }
    }]
  }
  const chart = new ApexCharts(document.querySelector("#sales-chart"), options)
  chart.render();

}

const renderPieChart = () => {
  const options = {
    series: state.channelData.way_cate_data.length > 0
      ? state.channelData.way_cate_data.map(item => item.count_num)
      : [],
    chart: {
      type: 'pie',
      height: 350,
      toolbar: { show: false },
      background: 'transparent',
      animations: {
        enabled: true,
        easing: 'easeinout',
        speed: 800,
        animateGradually: {
          enabled: true,
          delay: 150
        },
        dynamicAnimation: {
          enabled: true,
          speed: 350
        }
      }
    },
    labels: state.channelData.way_cate_data.length > 0
      ? state.channelData.way_cate_data.map(item => item.way_name)
      : [],
    colors: ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'],
    legend: {
      position: 'bottom',
      fontSize: '10px',
      fontFamily: 'Inter, sans-serif',
      markers: {
        width: 8,
        height: 8,
        radius: 4
      }
    },
    plotOptions: {
      pie: {
        donut: {
          size: '0%'
        },
        expandOnClick: true
      }
    },
    dataLabels: {
      enabled: true,
      formatter: function (val: number) {
        return val.toFixed(1) + '%'
      },
      style: {
        fontSize: '10px',
        fontFamily: 'Inter, sans-serif',
        fontWeight: 'bold'
      }
    },
    tooltip: {
      enabled: true,
      theme: 'light',
      style: {
        fontSize: '12px',
        fontFamily: 'Inter, sans-serif'
      },
      y: {
        formatter: function (val: number) {
          return val + ' 条'
        }
      }
    },
    responsive: [{
      breakpoint: 768,
      options: {
        chart: {
          height: 300
        },
        legend: {
          position: 'bottom'
        }
      }
    }]
  }
  const pieChart = new ApexCharts(document.querySelector("#pie-chart"), options)
  pieChart.render();
}

onMounted(() => {
  loadAllStatisticData();
})



</script>

<template>
  <div class="w-[90%] mx-auto grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3  lg:grid-cols-5 gap-4">
    <StatCard title="推送留存数" :value="state.basicData.message_total_num" description="" :icon="DatabaseIcon" />
    <StatCard title="托管消息数" :value="state.basicData.hosted_message_total_num" description="" :icon="BarChartIcon" />
    <StatCard title="今日发送数" :value="state.basicData.today_total_num" description="" :icon="SendIcon" />
    <StatCard title="今日成功数" :value="state.basicData.today_succ_num" description="" :icon="CheckCircleIcon" />
    <StatCard title="今日失败数" :value="state.basicData.today_failed_num" description="" :icon="XCircleIcon" />
  </div>

  <!-- 折线图 -->
  <!-- <LineChart
    :data="data"
    index="year"
    :categories="['Export Growth Rate', 'Import Growth Rate']"
    :y-formatter="(tick, i) => {
      return typeof tick === 'number'
        ? `$ ${new Intl.NumberFormat('us').format(tick).toString()}`
        : ''
    }"
  /> -->

  <!-- 图表区域 -->
  <div class="w-[90%] mx-auto mt-8 grid grid-cols-1 lg:grid-cols-10 gap-6">
    <!-- 折线图 -->
    <Card class="w-full lg:col-span-7">
      <CardHeader>
        <CardTitle>消息发送趋势</CardTitle>
        <CardDescription>最近30天的发送情况统计</CardDescription>
      </CardHeader>
      <CardContent>
        <div id="sales-chart" class="w-full h-[350px]"></div>
      </CardContent>
    </Card>

    <!-- 饼图 -->
    <Card class="w-full lg:col-span-3">
      <CardHeader>
        <CardTitle>发送渠道分布</CardTitle>
        <CardDescription>各发送渠道的使用情况统计</CardDescription>
      </CardHeader>
      <CardContent>
        <div id="pie-chart" class="w-full h-[350px]"></div>
      </CardContent>
    </Card>
  </div>
</template>
