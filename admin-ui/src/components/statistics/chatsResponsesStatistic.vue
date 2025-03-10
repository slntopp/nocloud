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
        :custom-legend-formater="legendFomatter"
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
    </template>
  </statistic-item>
</template>

<script setup>
import StatisticItem from "@/components/statistics/statisticItem.vue";
import { ref, watch } from "vue";
import { useStore } from "@/store";
import { debounce } from "@/functions";
import api from "@/api";
import DefaultChart from "@/components/statistics/defaultChart.vue";

const store = useStore();

const period = ref([]);
const periodType = ref("month");
const type = ref("bar");

const series = ref([]);
const categories = ref([]);
const summary = ref({});
const accounts = ref({});
const chartData = ref();
const seriesType = ref("amount");
const seriesTypes = [
  { label: "By users", value: "users" },
  { label: "Amount", value: "amount" },
];

const isDataLoading = ref(false);

function legendFomatter(val, opts) {
  const account = accounts.value[val] ?? { title: val };

  return `${account.title} ${
    summary.value[series.value[opts.seriesIndex]?.name]
      ? summary.value[series.value[opts.seriesIndex]?.name]
      : ""
  }`;
}

async function fetchData() {
  isDataLoading.value = true;

  try {
    const start_date = period.value[0].toISOString().split("T")[0];
    const end_date = period.value[1].toISOString().split("T")[0];

    const params = {
      entity: "ticket-responses",
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

watch([chartData, seriesType], async () => {
  if (!chartData.value) {
    return;
  }

  const newSeries = [];
  const newCategories = [];
  summary.value = {};

  const tempData = JSON.parse(JSON.stringify(chartData.value));

  if (seriesType.value == "users") {
    chartData.value.timeseries?.forEach((timeseries) => {
      const current = tempData.timeseries.filter((t) => t.ts === timeseries.ts);
      if (current.length <= 0) {
        return;
      }
      newCategories.push(timeseries.ts);

      current.map((ts) => {
        let index = newSeries.findIndex((series) => series.name === ts.user);
        if (index === -1) {
          newSeries.push({ name: ts.user, data: [] });
          index = newSeries.length - 1;
        }

        newSeries[index].data.push(ts.responses || 0);
      });

      tempData.timeseries = tempData.timeseries.filter(
        (t) => t.ts !== timeseries.ts
      );
    });

    await Promise.all(
      newSeries.map(async ({ name }) => {
        try {
          if (!accounts.value[name]) {
            accounts.value[name] = api.accounts.get(name);
            accounts.value[name] = await accounts.value[name];
          }
        } catch {
          accounts.value[name] = undefined;
        }
      })
    );

    summary.value = newSeries.reduce((acc, series) => {
      acc[series.name] = series.data.reduce((acc, v) => acc + v, 0);
      return acc;
    }, {});
  } else {
    newSeries.push({
      name: "Responses",
      data: [],
    });

    chartData.value.timeseries?.forEach((timeseries) => {
      const current = tempData.timeseries.filter((t) => t.ts === timeseries.ts);
      if (current.length <= 0) {
        return;
      }
      newCategories.push(timeseries.ts);

      newSeries[0].data.push(
        current.reduce((acc, ts) => acc + (ts.responses || 0), 0)
      );

      tempData.timeseries = tempData.timeseries.filter(
        (t) => t.ts !== timeseries.ts
      );
    });

    summary.value = newSeries.reduce((acc, series) => {
      acc[series.name] = series.data.reduce((acc, v) => acc + v, 0);
      return acc;
    }, {});
  }

  series.value = newSeries;
  categories.value = newCategories;
});
</script>
