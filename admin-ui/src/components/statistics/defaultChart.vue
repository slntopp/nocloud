<template>
  <apexchart
    ref="apexChartRef"
    :height="height"
    :options="chartOptions"
    :type="type"
    :series="coloredSeries"
  ></apexchart>
</template>

<script setup>
import { computed, ref, shallowRef, toRefs } from "vue";
import apexchart from "vue-apexcharts";
import { useStore } from "@/store";
import { getApexChartsColors } from "@/functions";

const props = defineProps([
  "series",
  "categories",
  "summary",
  "type",
  "stacked",
  "description",
  "options",
]);
const { categories, series, summary, type, stacked, description, options } =
  toRefs(props);

const store = useStore();

const apexChartRef = ref(null);

const colors = getApexChartsColors();

const height = ref(window.innerHeight * 0.7);

const internalSeries = shallowRef(series.value || []);
const internalCategories = shallowRef(categories.value || []);
const internalSummary = shallowRef(summary.value || {});

const updateChart = (newSeries, newCategories, newSummary) => {
  internalSeries.value = newSeries;
  internalCategories.value = newCategories;
  internalSummary.value = newSummary;
  
  if (apexChartRef.value) {
    const coloredData = newSeries.map((v, index) => ({ ...v, color: colors[index] }));
    apexChartRef.value.updateSeries(coloredData, false);  
    apexChartRef.value.updateOptions({
      xaxis: {
        categories: newCategories,
      },
    }, false, false);
  }
};

defineExpose({
  updateChart,
});

const chartOptions = computed(() => ({
  ...(options.value || {}),
  dataLabels: {
    enabled: (internalSeries.value?.[0]?.data?.length || series.value?.[0]?.data?.length) < 35,
    style:
      store.getters["app/theme"] == "light"
        ? {
            colors: Array(25).fill(() => "#262525"),
          }
        : {},
    ...(options.value?.dataLabels || {}),
  },
  title: description.value
    ? {
        text: description.value,
        align: "left",
      }
    : null,
  theme: {
    palette: "palette8",
    mode: store.getters["app/theme"],
  },
  chart: {
    ...(options.value?.chart || {}),
    stacked: !!stacked.value,
  },
  xaxis: {
    categories: internalCategories.value?.length ? internalCategories.value : categories.value,
  },
  legend: {
    onItemClick: {
      toggleDataSeries: true,
    },
    show: true,
    showForSingleSeries: true,
    formatter: function (val, opts) {
      const currentSummary = Object.keys(internalSummary.value).length ? internalSummary.value : summary.value;
      const currentSeries = internalSeries.value?.length ? internalSeries.value : series.value;
      return `${val} ${
        currentSummary[currentSeries[opts.seriesIndex]?.name]
          ? currentSummary[currentSeries[opts.seriesIndex]?.name]
          : ""
      }`;
    },
    ...(options.value?.legend || {}),
  },
}));

const coloredSeries = computed(() => {
  const currentSeries = internalSeries.value?.length ? internalSeries.value : series.value;
  return currentSeries.map((v, index) => ({ ...v, color: colors[index] }));
});
</script>
