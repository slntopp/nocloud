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
    :all-fields="allFields"
    :loading="isDataLoading"
    :fields="fields"
    :fields-multiple="!comparable"
    @input:fields="fields = $event"
    :comparable="comparable"
    @input:comparable="comparable = $event"
  >
    <template v-slot:content>
      <default-chart
        description="Accounts statistics"
        :stacked="type === 'bar' && !comparable"
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
import { ref, toRefs, watch } from "vue";
import DefaultChart from "@/components/statistics/defaultChart.vue";
import {
  statisticPeriodProps,
  statisticPeriodEmits,
  useStatisticData,
  buildComparablePeriodSeries,
} from "@/hooks/useStatisticChart";

const props = defineProps(statisticPeriodProps);
const { period, periodType, periods, type } = toRefs(props);

const emit = defineEmits(statisticPeriodEmits);

const comparable = ref(true);
const allFields = ref([
  { label: "Created", value: "created" },
  { label: "Active", value: "active" },
  { label: "Total", value: "total" },
]);
const fields = ref("created");

const series = ref([]);
const categories = ref([]);
const summary = ref({});

const { chartData, isDataLoading } = useStatisticData({
  entity: "accounts",
  periodType,
  period,
  periods,
  comparable,
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
    const cmp = buildComparablePeriodSeries(
      tempData,
      periods.value,
      fields.value
    );
    newSeries.push(...cmp.series);
    newCategories.push(...cmp.categories);
    Object.assign(summary.value, cmp.summary);
  }

  series.value = newSeries;
  categories.value = newCategories.map((c) => c.toString().split("T")[0]);
});
</script>
