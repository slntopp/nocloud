<template>
  <statistic-item
    :period="period"
    :periodType="periodType"
    @input:period="period = $event"
    @input:period-type="periodType = $event"
    :loading="isDataLoading"
    description="Revenue statistics for period"
    @input:type="type = $event"
    :type="type"
    :all-fields="allFields"
    :fields="fields"
    @input:fields="fields = $event"
    fields-multiple
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
import { computed, ref, watch } from "vue";
import { useStore } from "@/store";
import { debounce } from "@/functions";
import DefaultChart from "@/components/statistics/defaultChart.vue";

const store = useStore();

const period = ref([]);
const periodType = ref("month");
const type = ref("bar");
const allFields = ref([
  { label: "Revenue", value: "revenue" },
  { label: "Revenue start", value: "revenue_start" },
  { label: "Revenue renew", value: "revenue_renew" },
  { label: "Revenue balance", value: "revenue_balance" },
]);
const fields = ref(["revenue"]);

const series = ref([]);
const categories = ref([]);
const summary = ref({});

const isDataLoading = ref(false);
const chartData = ref();

const defaultCurrency = computed(() => store.getters["currencies/default"]);

async function fetchData() {
  isDataLoading.value = true;

  try {
    const start_date = period.value[0].toISOString().split("T")[0];
    const end_date = period.value[1].toISOString().split("T")[0];

    const params = {
      entity: "revenue",
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

function getFormattedPrice(price) {
  return [price.toFixed(0), defaultCurrency.value.code].join("");
}

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
    newCategories.push(timeseries.ts.split("T")[0]);
    newSeries.forEach((serie) => {
      serie.data.push(getFormattedPrice(timeseries[serie.id] || 0));
    });
  });

  summary.value = {};

  newSeries.forEach((serie) => {
    summary.value[serie.name] = getFormattedPrice(
      chartData.value.summary?.[serie.id] || 0
    );
  });

  series.value = newSeries;
  categories.value = newCategories;
});
</script>
