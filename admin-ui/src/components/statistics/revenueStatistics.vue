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
        :value-formatter="getFormattedPrice"
      />
    </template>
  </statistic-item>
</template>

<script setup>
import StatisticItem from "@/components/statistics/statisticItem.vue";
import { computed, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import DefaultChart from "@/components/statistics/defaultChart.vue";
import {
  statisticPeriodProps,
  statisticPeriodEmits,
  useStatisticData,
  buildComparablePeriodSeries,
} from "@/hooks/useStatisticChart";

const store = useStore();

const props = defineProps(statisticPeriodProps);
const { period, periodType, periods, type } = toRefs(props);

const emit = defineEmits(statisticPeriodEmits);

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

const comparable = ref(true);
const defaultCurrency = computed(() => store.getters["currencies/default"]);

const { chartData, isDataLoading } = useStatisticData({
  entity: "revenue",
  periodType,
  period,
  periods,
  comparable,
});

function getFormattedPrice(price) {
  return [(price || 0).toFixed(0), defaultCurrency.value.code].join("");
}

function getPrice(c, id) {
  if (id === "total") {
    return (
      (c?.revenue || 0) +
      (c?.revenue_start || 0) +
      (c?.revenue_balance || 0) +
      (c?.revenue_renew || 0)
    );
  } else {
    return c?.[id] || 0;
  }
}

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
        serie.data.push(getPrice(timeseries, serie.id));
      });
    });

    newSeries.forEach((serie) => {
      summary.value[serie.name] = getPrice(tempData[0].summary, serie.id);
    });
  } else {
    const cmp = buildComparablePeriodSeries(
      tempData,
      periods.value,
      fields.value,
      getPrice
    );
    newSeries.push(...cmp.series);
    newCategories.push(...cmp.categories);
    summary.value = cmp.summary;
  }

  series.value = newSeries;
  categories.value = newCategories.map((c) => c.toString().split("T")[0]);
});
</script>
