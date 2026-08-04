<template>
  <v-chart
    class="default-chart"
    :option="chartOption"
    autoresize
    @click="handleClick"
  />
</template>

<script setup>
import { computed, toRefs } from "vue";
import VChart from "vue-echarts";
import { use } from "echarts/core";
import { CanvasRenderer } from "echarts/renderers";
import { BarChart, LineChart } from "echarts/charts";
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
} from "echarts/components";
import { useStore } from "@/store";
import { getApexChartsColors } from "@/functions";

use([
  CanvasRenderer,
  BarChart,
  LineChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
]);

const props = defineProps({
  series: { type: Array, default: () => [] },
  categories: { type: Array, default: () => [] },
  summary: { type: Object, default: () => ({}) },
  type: { type: String, default: "bar" },
  stacked: { type: Boolean, default: false },
  description: { type: String, default: "" },
  options: { type: Object, default: () => ({}) },
  valueFormatter: { type: Function, default: null },
  tooltipFormatter: { type: Function, default: null },
  legendFormatter: { type: Function, default: null },
  onPointClick: { type: Function, default: null },
});
const {
  series,
  categories,
  summary,
  type,
  stacked,
  description,
  options,
  valueFormatter,
  tooltipFormatter,
  legendFormatter,
  onPointClick,
} = toRefs(props);

const store = useStore();
const colors = getApexChartsColors();

const theme = computed(() => store.getters["app/theme"]);
const textColor = computed(() =>
  theme.value === "light" ? "#262626" : "#e6e6e6"
);
const axisLineColor = computed(() =>
  theme.value === "light" ? "#d8d8d8" : "#3a3a3a"
);
const splitLineColor = computed(() =>
  theme.value === "light" ? "rgba(0,0,0,0.06)" : "rgba(255,255,255,0.08)"
);
const tooltipBg = computed(() =>
  theme.value === "light" ? "#ffffff" : "#2b2b2b"
);

const formatValue = (value) => {
  if (value == null || value === "") return "";
  return valueFormatter.value ? valueFormatter.value(value) : value;
};

const defaultLegendFormatter = (name) => {
  const total = summary.value?.[name];
  return total != null && total !== "" ? `${name}   ${formatValue(total)}` : name;
};

const defaultTooltipFormatter = (params) => {
  const list = Array.isArray(params) ? params : [params];
  const header = list[0]?.axisValueLabel ?? list[0]?.name ?? "";

  const rows = list
    .filter((p) => p.value != null && p.value !== "")
    .slice()
    .sort((a, b) => (b.value || 0) - (a.value || 0))
    .map(
      (p) =>
        `<div style="display:flex;align-items:center;gap:6px;margin:2px 0;">` +
        `<span style="width:8px;height:8px;border-radius:50%;background:${p.color};display:inline-block;"></span>` +
        `<span>${p.seriesName}: <b>${formatValue(p.value)}</b></span></div>`
    )
    .join("");

  return `<div style="font-weight:600;margin-bottom:4px;">${header}</div>${rows}`;
};

const baseSeriesType = computed(() => (type.value === "bar" ? "bar" : "line"));
const labelsEnabled = computed(() => (categories.value?.length || 0) < 35);

const echartSeries = computed(() =>
  (series.value || []).map((serie, index) => ({
    name: serie.name,
    type: baseSeriesType.value,
    stack: stacked.value ? "total" : undefined,
    smooth: type.value !== "bar",
    areaStyle: type.value === "area" ? { opacity: 0.22 } : undefined,
    symbolSize: type.value === "bar" ? undefined : 6,
    barMaxWidth: 40,
    itemStyle: { color: colors[index % colors.length] },
    lineStyle: type.value !== "bar" ? { width: 2.5 } : undefined,
    emphasis: { focus: "series" },
    label: {
      show: labelsEnabled.value,
      position: "top",
      color: textColor.value,
      fontSize: 11,
      formatter: (p) => formatValue(p.value),
    },
    data: serie.data,
  }))
);

const baseOption = computed(() => ({
  color: colors,
  textStyle: { color: textColor.value, fontFamily: "Inter, sans-serif" },
  title: description.value
    ? {
        text: description.value,
        left: "left",
        textStyle: { color: textColor.value, fontSize: 16, fontWeight: 600 },
      }
    : undefined,
  grid: {
    left: 8,
    right: 16,
    top: description.value ? 56 : 28,
    bottom: 64,
    containLabel: true,
  },
  legend: {
    bottom: 0,
    type: "scroll",
    textStyle: { color: textColor.value },
    formatter: legendFormatter.value || defaultLegendFormatter,
  },
  tooltip: {
    trigger: "axis",
    axisPointer: { type: type.value === "bar" ? "shadow" : "line" },
    backgroundColor: tooltipBg.value,
    borderColor: axisLineColor.value,
    textStyle: { color: textColor.value },
    formatter: tooltipFormatter.value || defaultTooltipFormatter,
    confine: true,
    extraCssText:
      "max-width: min(360px, 90vw); max-height: 70vh; overflow-y: auto; white-space: normal;",
  },
  xAxis: {
    type: "category",
    data: categories.value,
    axisLabel: {
      color: textColor.value,
      rotate: (categories.value?.length || 0) > 14 ? 45 : 0,
    },
    axisLine: { lineStyle: { color: axisLineColor.value } },
    axisTick: { alignWithLabel: true },
  },
  yAxis: {
    type: "value",
    axisLabel: { color: textColor.value, formatter: valueFormatter.value },
    axisLine: { lineStyle: { color: axisLineColor.value } },
    splitLine: { lineStyle: { color: splitLineColor.value } },
  },
  series: echartSeries.value,
}));

const mergeDeep = (base, override) => {
  if (!override) return base;
  const result = { ...base };
  Object.keys(override).forEach((key) => {
    const overrideValue = override[key];
    const baseValue = base[key];
    if (
      overrideValue &&
      typeof overrideValue === "object" &&
      !Array.isArray(overrideValue) &&
      baseValue &&
      typeof baseValue === "object" &&
      !Array.isArray(baseValue)
    ) {
      result[key] = mergeDeep(baseValue, overrideValue);
    } else {
      result[key] = overrideValue;
    }
  });
  return result;
};

const chartOption = computed(() => mergeDeep(baseOption.value, options.value));

const handleClick = (params) => {
  if (!onPointClick.value || params.componentType !== "series") return;

  const serie = series.value?.[params.seriesIndex];
  onPointClick.value({
    seriesIndex: params.seriesIndex,
    seriesName: params.seriesName,
    dataIndex: params.dataIndex,
    category: categories.value?.[params.dataIndex],
    value: params.value,
    meta: serie?.meta?.[params.dataIndex],
  });
};
</script>

<style scoped>
.default-chart {
  width: 100%;
  height: 70vh;
  min-height: 320px;
  max-height: 640px;
}

@media (max-width: 600px) {
  .default-chart {
    height: 60vh;
    min-height: 260px;
  }
}
</style>
