<template>
  <statistic-item
    :period="period"
    :periodType="periodType"
    @input:period="period = $event"
    @input:period-type="periodType = $event"
    @input:type="type = $event"
    :loading="isDataLoading"
    :type="type"
    description="Instances statistics for period"
  >
    <template v-slot:content>
      <default-chart
        :type="type"
        :series="series"
        :categories="categories"
        :summary="summary"
      />
    </template>

    <template v-slot:options>
      <v-select
        style="width: 150px"
        item-text="label"
        item-value="value"
        :items="seriesTypes"
        v-model="seriesType"
      />

      <v-select
        v-if="seriesType === 'type'"
        style="width: 100px"
        class="mr-2 ml-2"
        item-text="label"
        item-value="value"
        :items="seriesTypesSubKeys"
        v-model="seriesTypeSubKey"
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

const data = ref({});
const series = ref([]);
const categories = ref([]);
const summary = ref({});
const seriesType = ref("amount");
const seriesTypeSubKey = ref("total");

const seriesTypes = [
  { label: "By types", value: "type" },
  { label: "Amount", value: "amount" },
];

const seriesTypesSubKeys = [
  { label: "Created", value: "created" },
  { label: "Total", value: "total" },
];

const isDataLoading = ref(false);

async function fetchData() {
  isDataLoading.value = true;

  try {
    const start_date = period.value[0].toISOString().split("T")[0];
    const end_date = period.value[1].toISOString().split("T")[0];

    const params = {
      entity: "services",
      params: {
        start_date,
        end_date,
        with_timeseries: true,
      },
    };

    const response = await store.dispatch("statistic/get", params);
    data.value = response;
  } finally {
    isDataLoading.value = false;
  }
}

const fetchDataDebounced = debounce(fetchData, 300);

watch(period, () => {
  fetchDataDebounced();
});

watch([data, seriesType, seriesTypeSubKey], () => {
  const newSeries = [];
  const newCategories = [];

  const tempData = JSON.parse(JSON.stringify(data.value));

  if (seriesType.value === "amount") {
    newSeries.push(
      {
        name: "Created",
        data: [],
      },
      {
        name: "Total",
        data: [],
      }
    );

    data.value.timeseries?.forEach((timeseries) => {
      const current = tempData.timeseries.filter((t) => t.ts === timeseries.ts);
      if (current.length <= 0) {
        return;
      }

      newCategories.push(timeseries.ts);
      newSeries[0].data.push(
        current.reduce((acc, c) => acc + (c.created || 0), 0) || 0
      );
      newSeries[1].data.push(
        current.reduce((acc, c) => acc + (c.total || 0), 0) || 0
      );

      tempData.timeseries = tempData.timeseries.filter(
        (t) => t.ts !== timeseries.ts
      );
    });

    summary.value = {
      Created:
        Object.keys(data.value.summary || {}).reduce(
          (acc, key) => acc + data.value.summary[key].created,
          0
        ) || 0,
      Total:
        Object.keys(data.value.summary || {}).reduce(
          (acc, key) => acc + data.value.summary[key].total,
          0
        ) || 0,
    };
  } else {
    newSeries.push(
      ...Object.keys(data.value.summary || {}).map((key) => ({
        name: key,
        data: [],
      }))
    );

    data.value.timeseries?.forEach((timeseries) => {
      const current = tempData.timeseries.filter((t) => t.ts === timeseries.ts);
      if (current.length <= 0) {
        return;
      }

      newCategories.push(timeseries.ts);

      current.map((ts) => {
        const index = newSeries.findIndex((series) => series.name === ts.type);
        newSeries[index].data.push(ts[seriesTypeSubKey.value] || 0);
      });

      tempData.timeseries = tempData.timeseries.filter(
        (t) => t.ts !== timeseries.ts
      );
    });

    summary.value = Object.keys(data.value.summary || {}).reduce((acc, key) => {
      acc[key] = data.value.summary[key][seriesTypeSubKey.value] || 0;
      return acc;
    }, {});
  }

  series.value = newSeries;
  categories.value = newCategories;
});
</script>
