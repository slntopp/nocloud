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
    :comparable="comparable"
    @input:comparable="comparable = $event"
    description="Balance"
  >
    <template v-slot:options>
      <v-select
        style="width: 220px"
        item-text="label"
        item-value="value"
        :items="typeItems"
        v-model="txType"
        hide-details
      />
    </template>
    <template v-slot:content>
      <div v-if="qualityText" class="px-4 pb-2">{{ qualityText }}</div>
      <default-chart
        description="Balance"
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
import DefaultChart from "@/components/statistics/defaultChart.vue";
import {
  statisticPeriodProps,
  statisticPeriodEmits,
  useStatisticData,
  buildComparablePeriodSeries,
  buildSeriesFromTimeseries,
} from "@/hooks/useStatisticChart";

const props = defineProps(statisticPeriodProps);
const { period, periodType, periods, type } = toRefs(props);
const emit = defineEmits(statisticPeriodEmits);

const comparable = ref(false);
const txType = ref("");
const fields = ref("amount");
const series = ref([]);
const categories = ref([]);
const summary = ref({});
const qualityText = ref("");

const allFields = [
  { label: "Amount", value: "amount" },
  { label: "Count", value: "count" },
];

const requestParams = computed(() => ({ tx_type: txType.value || undefined }));

const { chartData, isDataLoading } = useStatisticData({
  entity: "balance",
  periodType,
  period,
  periods,
  comparable,
  params: requestParams,
  extraDeps: [txType],
});

const typeItems = computed(() => {
  const names = new Set();
  (chartData.value || []).forEach((block) => {
    (block?.summary?.types || []).forEach((name) => names.add(name));
  });
  return [
    { label: "All types", value: "" },
    ...[...names].sort().map((name) => ({ label: name, value: name })),
  ];
});

watch([chartData, fields, comparable], () => {
  const blocks = chartData.value;
  if (!blocks) return;
  qualityText.value = blocks[0]?.summary?.detail || "";
  const field = fields.value;
  if (!comparable.value) {
    const built = buildSeriesFromTimeseries(
      blocks[0]?.timeseries || [],
      field,
      (row, key) => Number(row?.[key]) || 0
    );
    series.value = built.series;
    categories.value = built.categories.map((item) => String(item).split("T")[0]);
    summary.value = built.summary;
    return;
  }
  const built = buildComparablePeriodSeries(
    blocks,
    periods.value,
    field,
    (row, key) => Number(row?.[key]) || 0
  );
  series.value = built.series;
  categories.value = built.categories.map((item) => String(item));
  summary.value = built.summary;
});
</script>
