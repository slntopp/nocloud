<template>
  <statistic-item
    :period="period"
    :periodType="periodType"
    @input:period="period = $event"
    @input:period-type="periodType = $event"
    :loading="isDataLoading"
    :type="type"
    @input:type="type = $event"
    description="Accounts statistics for period"
    :all-fields="allFields"
    :fields="fields"
    fields-multiple
    @input:fields="fields = $event"
  >
    <template v-slot:content>
      <default-chart
        :type="type"
        :series="series"
        :categories="categories"
        :summary="summary"
      />
    </template>
  </statistic-item>
</template>

<script setup>
import StatisticItem from "@/components/statistics/statisticItem.vue";
import { ref, watch } from "vue";
import { useStore } from "@/store";
import { debounce } from "@/functions";
import DefaultChart from "@/components/statistics/defaultChart.vue";

const store = useStore();

const period = ref([]);
const type = ref("bar");
const periodType = ref("month");
const allFields = ref([
  { label: "Created", value: "created" },
  { label: "Active", value: "active" },
  { label: "Total", value: "total" },
]);
const fields = ref(["created"]);

const series = ref([]);
const categories = ref([]);
const summary = ref({});
const chartData = ref();

const isDataLoading = ref(false);

async function fetchData() {
  isDataLoading.value = true;

  try {
    const start_date = period.value[0].toISOString().split("T")[0];
    const end_date = period.value[1].toISOString().split("T")[0];

    const params = {
      entity: "accounts",
      params: {
        start_date,
        end_date,
        with_timeseries: true,
      },
    };

    chartData.value = await store.dispatch("statistic/get", params);
  } finally {
    isDataLoading.value = false;
  }
}

const fetchDataDebounced = debounce(fetchData, 300);

watch(period, () => {
  fetchDataDebounced();
});

watch([chartData, fields], () => {
  if (!chartData.value || !fields.value.length) {
    return;
  }

  const newSeries = [];

  fields.value.forEach((key) => {
    newSeries.push({
      name: allFields.value.find((field) => field.value === key).label,
      data: [],
      id: key,
    });
  });

  const newCategories = [];

  chartData.value.timeseries?.forEach((timeseries) => {
    newCategories.push(timeseries.ts);
    newSeries.forEach((serie) => {
      serie.data.push(timeseries[serie.id] || 0);
    });
  });

  summary.value = {};

  newSeries.forEach((serie) => {
    summary.value[serie.name] = chartData.value.summary?.[serie.id] || 0;
  });

  series.value = newSeries;
  categories.value = newCategories;
});
</script>
