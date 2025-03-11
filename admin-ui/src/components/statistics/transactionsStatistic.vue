<template>
  <statistic-item
    :period="period"
    :periodType="periodType"
    @input:period="period = $event"
    @input:period-type="periodType = $event"
    :loading="isDataLoading"
    @input:type="type = $event"
    :type="type"
    :all-fields="allFields"
    :fields="fields"
    @input:fields="fields = $event"
    :fields-multiple="!comparable"
    :comparable="comparable"
    @input:comparable="comparable = $event"
    :periods="periods"
    @input:periods="periods = $event"
  >
    <template v-slot:content>
      <default-chart
        description="Transactions statistics"
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
import { formatToYYMMDD } from "@/functions";

const store = useStore();

const period = ref([]);
const periodType = ref("month");
const type = ref("bar");
const allFields = ref([
  { label: "Created", value: "created" },
  { label: "Total", value: "total" },
]);
const fields = ref(["created"]);
const comparable = ref(false);
const periods = ref({ first: [], second: [] });

const series = ref([]);
const categories = ref([]);
const summary = ref({});

const isDataLoading = ref(false);
const chartData = ref();

async function fetchData() {
  isDataLoading.value = true;

  try {
    chartData.value = await store.dispatch("statistic/getForChart", {
      entity: "transactions",
      periodType: periodType.value,
      periods: !comparable.value
        ? [period.value]
        : [periods.value.first, periods.value.second],
    });
  } finally {
    isDataLoading.value = false;
  }
}

const fetchDataDebounced = debounce(fetchData, 1000);

watch([period, periods, comparable], () => {
  if (!chartData.value) {
    fetchData();
  } else {
    fetchDataDebounced();
  }
});

watch(comparable, () => {
  if (comparable.value) {
    fields.value = "created";
  } else {
    fields.value = ["created"];
  }
});

watch([chartData, fields], () => {
  if (!chartData.value || !fields.value.length) {
    return;
  }

  const newSeries = [];
  const newCategories = [];
  summary.value = {};

  const tempData = JSON.parse(JSON.stringify(chartData.value));

  if (!comparable.value) {
    fields.value.forEach((key) => {
      newSeries.push({
        name: allFields.value.find((field) => field.value === key).label,
        data: [],
        id: key,
      });
    });

    tempData[0].timeseries?.forEach((timeseries) => {
      newCategories.push(timeseries.ts);
      newSeries.forEach((serie) => {
        serie.data.push(timeseries[serie.id] || 0);
      });
    });

    newSeries.forEach((serie) => {
      summary.value[serie.name] = tempData[0].summary?.[serie.id] || 0;
    });
  } else {
    Object.keys(periods.value).forEach((key) => {
      newSeries.push({
        name: `${formatToYYMMDD(periods.value[key][0])}/${formatToYYMMDD(
          periods.value[key][1]
        )}`,
        data: [],
      });
    });

    for (
      let index = 0;
      index <
      Math.max(
        tempData[0]?.timeseries?.length || 0,
        tempData[1]?.timeseries?.length || 0
      );
      index++
    ) {
      const first = tempData[0]?.timeseries?.[index];
      const second = tempData[1]?.timeseries?.[index];

      if (!newCategories.includes(index + 1)) {
        newCategories.push(index + 1);
      }

      newSeries[0].data.push(first?.[fields.value] || 0);
      newSeries[1].data.push(second?.[fields.value] || 0);
    }

    newSeries.forEach((serie) => {
      summary.value[serie.name] =
        serie.data.reduce((acc, a) => acc + a, 0) || 0;
    });
  }

  series.value = newSeries;
  categories.value = newCategories.map((c) => c.toString().split("T")[0]);
});
</script>
