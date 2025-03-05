<template>
  <apexchart
    width="100%"
    :options="chartOptions"
    :type="type"
    :series="series"
  ></apexchart>
</template>

<script setup>
import { computed, toRefs } from "vue";
import apexchart from "vue-apexcharts";
import { useStore } from "@/store";

const props = defineProps([
  "series",
  "categories",
  "summary",
  "customLegendFormater",
  "type",
]);
const { categories, series, summary, type } = toRefs(props);

const store = useStore();

const chartOptions = computed(() => ({
  dataLabels: {
    enabled: series.value?.[0]?.data?.length < 35,
  },
  theme: {
    palette: "palette8",
    mode: store.getters["app/theme"],
  },
  chart: {
    stacked: type.value === "bar",
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
</script>
