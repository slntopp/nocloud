<template>
  <statistic-item
    :period="period"
    :periodType="periodType"
    @input:period="period = $event"
    @input:period-type="periodType = $event"
    @input:type="type = $event"
    :loading="isDataLoading"
    :type="type"
    description="Chats statistics for period"
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
import { computed, ref, watch } from "vue";
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
const seriesTypeSubKey = ref("all");

const seriesTypes = [
  { label: "By departments", value: "departments" },
  { label: "Amount", value: "amount" },
];

const seriesTypesSubKeys = computed(() =>
  [
    seriesType.value === "amount" && { label: "All", value: "all" },
    { label: "Created", value: "created" },
    { label: "Closed", value: "closed" },
    { label: "Total", value: "total" },
  ].filter((v) => !!v)
);

function capitalizeFirstLetter(val) {
  return String(val).charAt(0).toUpperCase() + String(val).slice(1);
}

const isDataLoading = ref(false);

async function fetchData() {
  isDataLoading.value = true;

  try {
    const start_date = period.value[0].toISOString().split("T")[0];
    const end_date = period.value[1].toISOString().split("T")[0];

    const params = {
      entity: "tickets",
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

watch(seriesType, () => {
  if (seriesType.value === "departments") {
    seriesTypeSubKey.value = "total";
  }
});

watch([data, seriesType, seriesTypeSubKey], () => {
  const newSeries = [];
  const newCategories = [];

  const tempData = JSON.parse(JSON.stringify(data.value));

  if (seriesType.value === "amount") {
    if (seriesTypeSubKey.value === "all") {
      newSeries.push(
        {
          name: "Created",
          data: [],
        },
        {
          name: "Closed",
          data: [],
        },
        {
          name: "Total",
          data: [],
        }
      );
    } else {
      newSeries.push({
        name: capitalizeFirstLetter(seriesTypeSubKey.value),
        data: [],
      });
    }

    data.value.timeseries?.forEach((timeseries) => {
      const current = tempData.timeseries.filter((t) => t.ts === timeseries.ts);
      if (current.length <= 0) {
        return;
      }

      for (const series of newSeries) {
        series.data.push(
          current.reduce(
            (acc, c) => acc + (c[series.name.toLowerCase()] || 0),
            0
          ) || 0
        );
      }

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
    data.value.timeseries?.forEach((timeseries) => {
      const current = tempData.timeseries.filter((t) => t.ts === timeseries.ts);
      if (current.length <= 0) {
        return;
      }

      newCategories.push(timeseries.ts);

      current.map((ts) => {
        let index = newSeries.findIndex((series) => series.name === ts.dep);

        if (index == -1) {
          newSeries.push({ name: ts.dep, data: [] });
          index = newSeries.length - 1;
        }

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
