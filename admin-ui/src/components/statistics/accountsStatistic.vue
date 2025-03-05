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

const series = ref([]);
const categories = ref([]);
const summary = ref({});

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

    const response = await store.dispatch("statistic/get", params);
    const newSeries = [
      {
        name: "Created",
        data: [],
      },
      {
        name: "Active",
        data: [],
      },
      {
        name: "Total",
        data: [],
      },
    ];
    const newCategories = [];

    response.timeseries?.forEach((timeseries) => {
      newCategories.push(timeseries.ts);
      newSeries[0].data.push(timeseries.created || 0);
      newSeries[1].data.push(timeseries.active || 0);
      newSeries[2].data.push(timeseries.total || 0);
    });

    summary.value = {
      Active: response.summary?.active || 0,
      Created: response.summary?.created || 0,
      Total: response.summary?.total || 0,
    };
    series.value = newSeries;
    categories.value = newCategories;
  } finally {
    isDataLoading.value = false;
  }
}

const fetchDataDebounced = debounce(fetchData, 300);

watch(period, () => {
  fetchDataDebounced();
});
</script>
