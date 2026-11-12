<script setup lang="ts">
import StatCard from '@/components/pages/dashboard/CardNum.vue'
import {
  InboxIcon,
  FileTextIcon,
  SendIcon,
  CheckCircleIcon,
  XCircleIcon,
  RadioIcon
} from 'lucide-vue-next'
import { onMounted, onUnmounted, reactive } from 'vue';
import { request } from '@/api/api';
import { toast } from 'vue-sonner';
import ApexCharts from 'apexcharts'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

let lineChartInstance: ApexCharts | null = null;
let pieChartInstance: ApexCharts | null = null;

// 图表延时动画定时器管理（防止路由快速切换时引发内存泄漏或操作已卸载 DOM）
let chartTimeouts: ReturnType<typeof setTimeout>[] = [];
const registerChartTimeout = (fn: () => void, delay: number) => {
  const id = setTimeout(() => {
    chartTimeouts = chartTimeouts.filter(t => t !== id);
    fn();
  }, delay);
  chartTimeouts.push(id);
  return id;
};
const clearAllChartTimeouts = () => {
  chartTimeouts.forEach(t => clearTimeout(t));
  chartTimeouts = [];
};

const isDarkMode = () => document.documentElement.classList.contains('dark');

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
  trendDays: 30,
  basicData: {
    message_total_num: 0,
    hosted_message_total_num: 0,
    ways_total_num: 0,
    today_total_num: 0,
    today_succ_num: 0,
    today_failed_num: 0,
  },
  trendData: {
    latest_send_data: [] as SendData[] | null,
  },
  channelData: {
    way_cate_data: [] as CateData[] | null,
  }
});

// 获取基础统计数据
const getBasicStatisticData = async () => {
  try {
    const rsp = await request.get('/statistic?type=basic');
    if (rsp && rsp.data && rsp.data.code == 200) {
      state.basicData = { ...state.basicData, ...rsp.data.data };
      if (rsp.data.data.ways_total_num === undefined) {
        // 兼容当前运行后端若尚未热重载时的兜底统计
        const wayRsp = await request.get('/sendways/list', { params: { page: 1, size: 1 } });
        if (wayRsp?.data?.code === 200 && wayRsp.data.data?.total !== undefined) {
          state.basicData.ways_total_num = wayRsp.data.data.total;
        }
      }
    } else {
      toast.error(rsp?.data?.msg || '获取基础统计数据失败');
    }
  } catch (error) {
    toast.error('获取基础统计数据时发生错误');
  }
}

// 获取趋势统计数据
const getTrendStatisticData = async () => {
  try {
    // 根据屏幕大小决定请求的天数
    const isSmallScreen = window.innerWidth < 768;
    const days = isSmallScreen ? 15 : 30;
    state.trendDays = days;

    const rsp = await request.get(`/statistic?type=trend&days=${days}`);
    if (rsp && rsp.data && rsp.data.code == 200) {
      state.trendData = rsp.data.data;
      renderLineChart();
    } else {
      toast.error(rsp?.data?.msg || '获取趋势统计数据失败');
    }
  } catch (error) {
    toast.error('获取趋势统计数据时发生错误');
  }
}

// 获取渠道统计数据
const getChannelStatisticData = async () => {
  try {
    const rsp = await request.get('/statistic?type=channels');
    if (rsp && rsp.data && rsp.data.code == 200) {
      state.channelData = rsp.data.data;
      renderPieChart();
    } else {
      toast.error(rsp?.data?.msg || '获取渠道统计数据失败');
    }
  } catch (error) {
    toast.error('获取渠道统计数据时发生错误');
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

// 基于 requestAnimationFrame 逐帧插值驱动折线从左至右连线，折线上的数据点随线条前进慢慢浮现
let lineAnimRafId: number | null = null;
const animateLineChartDraw = (duration = 1600) => {
  const chartElement = document.querySelector("#sales-chart") as HTMLElement;
  if (!chartElement) return;

  if (lineAnimRafId !== null) {
    cancelAnimationFrame(lineAnimRafId);
    lineAnimRafId = null;
  }

  // 1. 查找折线真实 SVG Path（仅匹配折线路径，坚决排除任何 marker 数据点与图例）
  let linePaths = Array.from(
    chartElement.querySelectorAll<SVGPathElement>("path.apexcharts-line")
  ).filter(path => {
    const cls = path.getAttribute("class") || "";
    return !cls.includes("apexcharts-marker") && !path.closest('.apexcharts-legend');
  });

  if (!linePaths.length) {
    linePaths = Array.from(
      chartElement.querySelectorAll<SVGPathElement>(".apexcharts-series > path")
    ).filter(path => {
      const cls = path.getAttribute("class") || "";
      return !cls.includes("apexcharts-marker") && !cls.includes("apexcharts-grid") && !path.closest('.apexcharts-legend');
    });
  }

  if (!linePaths.length) {
    chartElement.classList.remove('chart-animating');
    return;
  }

  // 记录各条折线精确周长
  const pathLengths = linePaths.map(path => {
    let len = 0;
    try {
      len = path.getTotalLength();
    } catch {
      len = 3000;
    }
    return len > 0 ? len : 3000;
  });

  // 初始将折线隐藏在起始位置，设置 opacity: 1 准备展开
  linePaths.forEach((path, i) => {
    const len = pathLengths[i];
    path.setAttribute('stroke-dasharray', `${len}`);
    path.setAttribute('stroke-dashoffset', `${len}`);
    path.style.opacity = '1';
  });

  // 2. 严格按系列获取折线数据点（排序保证从左到右点位严格对齐连线进度，彻底排除图例 marker）
  const seriesWraps = Array.from(
    chartElement.querySelectorAll<SVGGElement>(".apexcharts-series-markers-wrap")
  ).sort((a, b) => {
    const idxA = parseInt(a.getAttribute("data:realIndex") || "0", 10);
    const idxB = parseInt(b.getAttribute("data:realIndex") || "0", 10);
    return idxA - idxB;
  });

  let seriesMarkerList: SVGElement[][] = [];
  if (seriesWraps.length > 0) {
    seriesMarkerList = seriesWraps.map(wrap =>
      Array.from(wrap.querySelectorAll<SVGElement>(".apexcharts-marker"))
        .filter(m => !m.closest('.apexcharts-legend') && !m.classList.contains('apexcharts-legend-marker'))
        .sort((a, b) => {
          const jA = parseInt(a.getAttribute("j") || a.getAttribute("rel") || "0", 10);
          const jB = parseInt(b.getAttribute("j") || b.getAttribute("rel") || "0", 10);
          return jA - jB;
        })
    );
  } else {
    const allLineMarkers = Array.from(
      chartElement.querySelectorAll<SVGElement>(".apexcharts-series .apexcharts-marker")
    ).filter(m => !m.closest('.apexcharts-legend') && !m.classList.contains('apexcharts-legend-marker'));
    const seriesCount = linePaths.length || 3;
    const perCount = Math.max(1, Math.floor(allLineMarkers.length / seriesCount));
    for (let s = 0; s < seriesCount; s++) {
      seriesMarkerList.push(allLineMarkers.slice(s * perCount, (s + 1) * perCount));
    }
  }

  // 初始仅设置透明隐藏，不修改 transform / translate，确保圆点形状绝对完整
  seriesMarkerList.forEach(list => {
    list.forEach(marker => {
      marker.style.opacity = '0';
      marker.style.transition = 'none';
    });
  });

  const shownSets = seriesMarkerList.map(() => new Set<number>());
  const startTime = performance.now();

  const step = (now: number) => {
    const elapsed = now - startTime;
    const t = Math.min(1, elapsed / duration);
    // easeOutCubic 曲线：平滑向右连线
    const progress = 1 - Math.pow(1 - t, 3);

    // 逐帧更新折线 stroke-dashoffset
    linePaths.forEach((path, i) => {
      const len = pathLengths[i];
      path.setAttribute('stroke-dashoffset', `${len * (1 - progress)}`);
    });

    // 各系列对应数据点在线条到达时，以从容速度淡入（慢慢出现且形状保持饱满正圆）
    seriesMarkerList.forEach((markers, sIdx) => {
      const total = markers.length;
      const set = shownSets[sIdx];
      markers.forEach((marker, pIdx) => {
        if (set.has(pIdx)) return;
        const threshold = total > 1 ? pIdx / (total - 1) : 0;
        if (progress >= threshold * 0.95) {
          set.add(pIdx);
          marker.style.transition = 'opacity 0.45s cubic-bezier(0.2, 0.8, 0.2, 1)';
          marker.style.opacity = '1';
        }
      });
    });

    if (t < 1) {
      lineAnimRafId = requestAnimationFrame(step);
    } else {
      lineAnimRafId = null;
      // 动画完成：确保所有折线与所有数据点 100% 完整显示并恢复原生状态
      linePaths.forEach(path => {
        path.removeAttribute('stroke-dasharray');
        path.removeAttribute('stroke-dashoffset');
        path.style.opacity = '';
      });
      seriesMarkerList.forEach(list => {
        list.forEach(marker => {
          marker.style.opacity = '';
          marker.style.transition = '';
        });
      });
      chartElement.classList.remove('chart-animating');
    }
  };

  lineAnimRafId = requestAnimationFrame(step);
}

// 饼图轻快顺时针一个接一个拼满整张饼图（间隔70ms）
const animatePieChartDraw = () => {
  const chartElement = document.querySelector("#pie-chart") as HTMLElement;
  if (!chartElement) return;

  const slices = chartElement.querySelectorAll<SVGElement>(
    '.apexcharts-pie-series, .apexcharts-series.apexcharts-pie-series'
  );
  if (!slices.length) {
    chartElement.classList.remove('chart-animating');
    return;
  }

  // 数据标签先隐藏，待全部切片拼满后淡入
  const dataLabels = chartElement.querySelectorAll<SVGElement>(
    '.apexcharts-datalabels, .apexcharts-pie-label, .apexcharts-datalabel, .apexcharts-data-labels'
  );
  dataLabels.forEach((label) => {
    label.style.opacity = '0';
    label.style.transition = 'none';
  });

  const sliceDelay = 70; // 每个切片间隔70ms，整体500~600ms左右拼满
  const total = slices.length;

  slices.forEach((slice, idx) => {
    const delay = idx * sliceDelay;

    slice.style.opacity = '0';
    slice.style.transform = 'scale(0.3)';
    slice.style.transformOrigin = 'center';
    slice.style.transformBox = 'fill-box';
    slice.style.transition = 'none';

    registerChartTimeout(() => {
      slice.style.transition = 'opacity 0.16s ease-out, transform 0.24s cubic-bezier(0.34, 1.56, 0.64, 1)';
      slice.style.opacity = '1';
      slice.style.transform = 'scale(1)';
    }, delay);
  });

  // 全部切片拼接完成后，迅速淡入百分比标签
  const allSlicesFinished = total * sliceDelay + 120;
  registerChartTimeout(() => {
    dataLabels.forEach((label) => {
      label.style.transition = 'opacity 0.25s ease-in-out';
      label.style.opacity = '1';
    });
  }, allSlicesFinished);

  // 动画结束清理内联样式，交还 ApexCharts 处理 Hover 凸起与交互
  registerChartTimeout(() => {
    slices.forEach((slice) => {
      slice.style.opacity = '';
      slice.style.transform = '';
      slice.style.transformOrigin = '';
      slice.style.transformBox = '';
      slice.style.transition = '';
    });
    dataLabels.forEach((label) => {
      label.style.opacity = '';
      label.style.transition = '';
    });
    chartElement.classList.remove('chart-animating');
  }, allSlicesFinished + 250);
}

const renderLineChart = async () => {
  const latestSendData = state.trendData.latest_send_data || [];

  const options: any = {
    series: [
      {
        name: '发送总数',
        data: latestSendData.length > 0
          ? latestSendData.map(item => item.num || 0)
          : []
      },
      {
        name: '发送成功数',
        data: latestSendData.length > 0
          ? latestSendData.map(item => item.day_succ_num || 0)
          : []
      },
      {
        name: '发送失败数',
        data: latestSendData.length > 0
          ? latestSendData.map(item => item.day_failed_num || 0)
          : []
      },
    ],
    chart: {
      type: 'line',
      height: 350,
      toolbar: { show: false },
      background: 'transparent',
      animations: {
        // 禁用底层 baseline 变形，由 animateLineChartDraw 精确驱动从首点到尾点的连续连线动效
        enabled: false
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
    noData: {
      text: '暂无消息发送趋势数据',
      align: 'center',
      verticalAlign: 'middle',
      style: {
        color: '#64748b',
        fontSize: '13px',
        fontFamily: 'Inter, sans-serif'
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
      categories: latestSendData.length > 0
        ? latestSendData.map(item => item.day)
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
      },
      tooltip: {
        enabled: false
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
      borderColor: isDarkMode() ? '#334155' : '#e2e8f0',
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
        shape: 'circle',
        size: 5,
        strokeWidth: 0,
        radius: 12
      }
    },
    tooltip: {
      enabled: true,
      shared: true,
      intersect: false,
      theme: isDarkMode() ? 'dark' : 'light',
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

        const containerCls = 'bg-card text-foreground p-3 rounded-lg shadow-lg border border-border';
        const labelMutedCls = 'text-sm text-muted-foreground';
        const valueStrongCls = 'text-sm font-medium text-foreground';
        const successRateCls = 'text-sm font-medium';

        return `
          <div class="${containerCls}">
            <div class="font-medium mb-2">${w.globals.categoryLabels[dataPointIndex]}</div>
            <div class="space-y-1">
              <div class="flex items-center justify-between">
                <span class="flex items-center">
                  <span class="w-2 h-2 bg-green-500 rounded-full mr-2"></span>
                  <span class="${labelMutedCls}">成功:</span>
                </span>
                <span class="${valueStrongCls}">${successCount} 条</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="flex items-center">
                  <span class="w-2 h-2 bg-red-500 rounded-full mr-2"></span>
                  <span class="${labelMutedCls}">失败:</span>
                </span>
                <span class="${valueStrongCls}">${failedCount} 条</span>
              </div>
              <div class="border-t border-border pt-1 mt-2">
                <div class="flex items-center justify-between">
                  <span class="${labelMutedCls}">成功率:</span>
                  <span class="${successRateCls}">${successRate}%</span>
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
          position: 'top',
          horizontalAlign: 'right',
          floating: true,
          offsetY: -20,
          offsetX: -5,
          fontSize: '10px',
          markers: {
            shape: 'circle',
            size: 4,
            strokeWidth: 0,
            radius: 12
          },
          itemMargin: {
            horizontal: 5,
            vertical: 0
          }
        }
      }
    }]
  }
  const chartElement = document.querySelector("#sales-chart") as HTMLElement;
  if (!chartElement) return;
  if (lineAnimRafId !== null) {
    cancelAnimationFrame(lineAnimRafId);
    lineAnimRafId = null;
  }
  if (lineChartInstance) {
    lineChartInstance.destroy();
    lineChartInstance = null;
  }
  chartElement.innerHTML = '';
  if (latestSendData.length > 0) {
    chartElement.classList.add('chart-animating');
  } else {
    chartElement.classList.remove('chart-animating');
  }
  lineChartInstance = new ApexCharts(chartElement, options);
  await lineChartInstance.render();
  if (latestSendData.length > 0) {
    requestAnimationFrame(() => {
      animateLineChartDraw(1600);
    });
  }
}

const renderPieChart = async () => {
  const wayCateData = state.channelData.way_cate_data || [];
  const options: any = {
    series: wayCateData.length > 0
      ? wayCateData.map(item => item.count_num)
      : [],
    chart: {
      type: 'pie',
      height: 350,
      toolbar: { show: false },
      background: 'transparent',
      animations: {
        // 关闭 ApexCharts 原生瞬时重绘，由 animatePieChartDraw 接管顺时针一个接一个拼满圆盘的动效
        enabled: false
      }
    },
    noData: {
      text: '暂无发送渠道数据',
      align: 'center',
      verticalAlign: 'middle',
      style: {
        color: '#64748b',
        fontSize: '13px',
        fontFamily: 'Inter, sans-serif'
      }
    },
    labels: wayCateData.length > 0
      ? wayCateData.map(item => item.way_name)
      : [],
    colors: [
      '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6',
      '#06b6d4', '#ec4899', '#f97316', '#14b8a6', '#6366f1'
    ],
    legend: {
      position: 'bottom',
      fontSize: '10px',
      fontFamily: 'Inter, sans-serif',
      markers: {
        shape: 'circle',
        size: 5,
        strokeWidth: 0,
        radius: 12
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
      theme: isDarkMode() ? 'dark' : 'light',
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
  const pieChartElement = document.querySelector("#pie-chart") as HTMLElement;
  if (!pieChartElement) return;
  clearAllChartTimeouts();
  if (pieChartInstance) {
    pieChartInstance.destroy();
    pieChartInstance = null;
  }
  pieChartElement.innerHTML = '';
  if (wayCateData.length > 0) {
    pieChartElement.classList.add('chart-animating');
  } else {
    pieChartElement.classList.remove('chart-animating');
  }
  pieChartInstance = new ApexCharts(pieChartElement, options);
  await pieChartInstance.render();
  if (wayCateData.length > 0) {
    requestAnimationFrame(() => {
      animatePieChartDraw();
    });
  }
}

// 获取今日日期
const getTodayDate = () => {
  const today = new Date();
  return today.toISOString().split('T')[0]; // 返回 YYYY-MM-DD 格式
}

onMounted(() => {
  loadAllStatisticData();
})

onUnmounted(() => {
  clearAllChartTimeouts();
  if (lineAnimRafId !== null) {
    cancelAnimationFrame(lineAnimRafId);
    lineAnimRafId = null;
  }
  if (lineChartInstance) {
    lineChartInstance.destroy();
    lineChartInstance = null;
  }
  if (pieChartInstance) {
    pieChartInstance.destroy();
    pieChartInstance = null;
  }
})



</script>

<template>
  <div class="w-[90%] mx-auto pt-4 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">

    <StatCard title="托管消息数" :value="state.basicData.hosted_message_total_num" description="" :icon="InboxIcon"
      route-path="/hostedmessage" />
    <StatCard title="发送渠道数" :value="state.basicData.ways_total_num" description="" :icon="RadioIcon"
      route-path="/sendways" />
    <StatCard title="发送日志数" :value="state.basicData.message_total_num" description="" :icon="FileTextIcon"
      route-path="/sendlogs" />
    <StatCard title="今日发送数" :value="state.basicData.today_total_num" description="" :icon="SendIcon"
      :route-path="`/sendlogs?query=${encodeURIComponent(JSON.stringify({ day_created_on: getTodayDate() }))}`" />
    <StatCard title="今日成功数" :value="state.basicData.today_succ_num" description="" :icon="CheckCircleIcon"
      :route-path="`/sendlogs?query=${encodeURIComponent(JSON.stringify({ status: '1', day_created_on: getTodayDate() }))}`" />
    <StatCard title="今日失败数" :value="state.basicData.today_failed_num" description="" :icon="XCircleIcon"
      :route-path="`/sendlogs?query=${encodeURIComponent(JSON.stringify({ status: '0', day_created_on: getTodayDate() }))}`" />
  </div>

  <!-- 图表区域 -->
  <div class="w-[90%] mx-auto mt-8 grid grid-cols-1 lg:grid-cols-10 gap-6">
    <!-- 折线图 -->
    <Card class="w-full lg:col-span-7 hover:shadow-lg transition-all duration-300">
      <CardHeader>
        <CardTitle>消息发送趋势</CardTitle>
        <CardDescription>最近{{ state.trendDays }}天的发送情况统计</CardDescription>
      </CardHeader>
      <CardContent>
        <div id="sales-chart" class="w-full h-[350px]"></div>
      </CardContent>
    </Card>

    <!-- 饼图 -->
    <Card class="w-full lg:col-span-3 hover:shadow-lg transition-all duration-300">
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

<style scoped>
/* 动画中折线初始隐藏，防止渲染瞬间显示完整折线导致的闪烁 */
:deep(#sales-chart.chart-animating path.apexcharts-line) {
  opacity: 0;
}

/* 动画中折线数据点初始隐藏，通过 JS 沿线时序逐步淡入 */
:deep(#sales-chart.chart-animating .apexcharts-series-markers .apexcharts-marker),
:deep(#sales-chart.chart-animating .apexcharts-series-markers-wrap .apexcharts-marker) {
  opacity: 0;
}

/* 饼图动画中初始隐藏切片与数据标签，防止渲染瞬间闪现完整饼图 */
:deep(#pie-chart.chart-animating .apexcharts-pie-series) {
  opacity: 0;
  transform: scale(0.3);
  transform-origin: center;
  transform-box: fill-box;
}

:deep(#pie-chart.chart-animating .apexcharts-datalabels),
:deep(#pie-chart.chart-animating .apexcharts-pie-label),
:deep(#pie-chart.chart-animating .apexcharts-datalabel),
:deep(#pie-chart.chart-animating .apexcharts-data-labels) {
  opacity: 0;
}

/* 确保图例小圆点 100% 完整饱满正圆居中，彻底防止 SVG 边缘裁切成四分之一残缺扇形 */
:deep(.apexcharts-legend-marker) {
  overflow: visible !important;
}
:deep(.apexcharts-legend-marker svg) {
  overflow: visible !important;
}
:deep(.apexcharts-legend-marker svg path:not(.apexcharts-marker-triangle)) {
  transform: translate(50%, 50%) !important;
}

/* 折线图数据点悬停交互高亮 */
:deep(#sales-chart .apexcharts-series-markers .apexcharts-marker:hover) {
  transform: scale(1.35) !important;
  transition: transform 0.18s cubic-bezier(0.34, 1.56, 0.64, 1) !important;
}

/* 饼图各切片顺时针渐进展开展开时的圆心基准与柔和渲染 */
:deep(#pie-chart .apexcharts-pie-series path) {
  transform-origin: center;
  transition: opacity 0.3s ease, filter 0.2s ease;
}

:deep(#pie-chart .apexcharts-pie-series path:hover) {
  filter: brightness(1.08);
}
</style>

