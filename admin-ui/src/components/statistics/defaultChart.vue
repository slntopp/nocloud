<template>
  <apexchart
    :height="height"
    :options="chartOptions"
    :type="type"
    :series="coloredSeries"
  ></apexchart>
</template>

<script setup>
import { computed, ref, toRefs } from "vue";
import apexchart from "vue-apexcharts";
import { useStore } from "@/store";
import { getApexChartsColors } from "@/functions";

const props = defineProps([
  "series",
  "categories",
  "summary",
  "customLegendFormater",
  "type",
  "stacked",
  "description",
  "options",
]);
const { categories, series, summary, type, stacked, description, options } =
  toRefs(props);

const store = useStore();

const colors = getApexChartsColors();

const height = ref(window.innerHeight * 0.7);

const chartOptions = computed(() => ({
  ...(options.value || {}),
  dataLabels: {
    enabled: series.value?.[0]?.data?.length < 35,
    style:
      store.getters["app/theme"] == "light"
        ? {
            colors: Array(25).fill(() => "#262525"),
          }
        : {},
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
    stacked: !!stacked.value,
  },
  xaxis: {
    categories: categories.value,
  },
  legend: {
    onItemClick: {
      toggleDataSeries: true,
    },
    show: true,
    showForSingleSeries: true,
    formatter: props.customLegendFormater
      ? props.customLegendFormater
      : function (val, opts) {
          return `${val} ${
            summary.value[series.value[opts.seriesIndex]?.name]
              ? summary.value[series.value[opts.seriesIndex]?.name]
              : ""
          }`;
        },
  },
}));

const coloredSeries = computed(() => {
  return series.value.map((v, index) => ({ ...v, color: colors[index] }));
});
</script>
