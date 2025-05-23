<template>
  <statistic-item
    :period="period"
    :periodType="periodType"
    :periods="periods"
    :type="type"
    :period-offset="periodOffset"
    :periods-first-offset="periodsFirstOffset"
    :periods-second-offset="periodsSecondOffset"
    @input:period="emit('update:period', $event)"
    @input:period-type="emit('update:period-type', $event)"
    @input:periods="emit('update:periods', $event)"
    @input:type="emit('update:type', $event)"
    @input:period-offset="emit('update:period-offset', $event)"
    @input:periods-first-offset="emit('update:periods-first-offset', $event)"
    @input:periods-second-offset="emit('update:periods-second-offset', $event)"
    :loading="isDataLoading"
    :all-fields="allFields"
    :fields="fields"
    @input:fields="fields = $event"
    :fields-multiple="!comparable"
    :comparable="comparable"
    @input:comparable="comparable = $event"
  >
    <template v-slot:content>
      <default-chart
        description="Revenue statistics"
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
import { computed, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import { debounce } from "@/functions";
import DefaultChart from "@/components/statistics/defaultChart.vue";
import { formatToYYMMDD } from "@/functions";

const store = useStore();

const props = defineProps({
  period: { type: Array, default: () => [] },
  periodType: { type: String, default: "month" },
  type: { type: String, default: "bar" },
  periods: { type: Object, default: () => ({ first: [], second: [] }) },
  periodOffset: { type: Number, default: 0 },
  periodsFirstOffset: { type: Number, default: 0 },
  periodsSecondOffset: { type: Number, default: -1 },
});
const { period, periodType, periods, type } = toRefs(props);

const emit = defineEmits([
  "update:period",
  "update:periods",
  "update:period-type",
  "update:type",
  "update:period-offset",
  "update:periods-first-offset",
  "update:periods-second-offset",
]);

const allFields = ref([
  { label: "Other invoices", value: "revenue" },
  { label: "Instance start", value: "revenue_start" },
  { label: "Instance renew", value: "revenue_renew" },
  { label: "Top-up balance", value: "revenue_balance" },
  { label: "Total", value: "total" },
]);
const fields = ref("total");

const series = ref([]);
const categories = ref([]);
const summary = ref({});

const isDataLoading = ref(false);
const chartData = ref();
const comparable = ref(true);
const defaultCurrency = computed(() => store.getters["currencies/default"]);

function getFormattedPrice(price) {
  return [price.toFixed(0), defaultCurrency.value.code].join("");
}

function getPrice(c, id) {
  if (id === "total") {
    return (
      (c.revenue || 0) +
      (c.revenue_start || 0) +
      (c.revenue_balance || 0) +
      (c.revenue_renew || 0)
    );
  } else {
    return c[id] || 0;
  }
}

async function fetchData() {
  isDataLoading.value = true;

  try {
    chartData.value = await store.dispatch("statistic/getForChart", {
      entity: "revenue",
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

debounce(fetchData, 100)();

watch([period, periods, comparable], () => {
  if (!chartData.value) {
    fetchData();
  } else {
    fetchDataDebounced();
  }
});

watch(comparable, (val) => {
  if (val) {
    fields.value = "total";
  } else {
    fields.value = ["total"];
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
      newCategories.push(timeseries.ts.split("T")[0]);
      newSeries.forEach((serie) => {
        serie.data.push(getFormattedPrice(getPrice(timeseries, serie.id) || 0));
      });
    });

    newSeries.forEach((serie) => {
      summary.value[serie.name] = getFormattedPrice(
        getPrice(tempData[0].summary, serie.id) || 0
      );
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

      newSeries[0].data.push(getPrice(first || {}, fields.value) || 0);
      newSeries[1].data.push(getPrice(second || {}, fields.value) || 0);
    }

    newSeries.forEach((serie) => {
      summary.value[serie.name] = getFormattedPrice(
        serie.data.reduce((acc, a) => acc + a, 0) || 0
      );

      serie.data = serie.data.map((el) => getFormattedPrice(el));
    });
  }

  series.value = newSeries;
  categories.value = newCategories.map((c) => c.toString().split("T")[0]);
});
</script>
