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
    :not-comparable="seriesType !== 'amount'"
    @input:comparable="comparable = $event"
    description="Chats"
  >
    <template v-slot:options>
      <v-select
        style="width: 200px"
        item-text="label"
        item-value="value"
        :items="seriesTypes"
        v-model="seriesType"
        hide-details
      />
    </template>
    <template v-slot:content>
      <div v-if="qualityText" class="px-4 pb-2">{{ qualityText }}</div>
      <default-chart
        description="Chats"
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
  buildGroupedSeries,
  limitTop,
} from "@/hooks/useStatisticChart";

const props = defineProps(statisticPeriodProps);
const { period, periodType, periods, type } = toRefs(props);
const emit = defineEmits(statisticPeriodEmits);

const comparable = ref(false);
const seriesType = ref("amount");
const fields = ref("created");
const series = ref([]);
const categories = ref([]);
const summary = ref({});
const qualityText = ref("");

const seriesTypes = [
  { label: "All", value: "amount" },
  { label: "By department", value: "departments" },
  { label: "By client", value: "clients" },
  { label: "By admin", value: "admins" },
];

const entity = computed(() => {
  if (seriesType.value === "clients") return "chat_owners";
  if (seriesType.value === "admins") return "ticket-responses";
  return "tickets";
});

const allFields = computed(() => {
  if (seriesType.value === "admins") {
    return [{ label: "Replies", value: "responses" }];
  }
  const fieldsList = [
    { label: "Created", value: "created" },
    { label: "Active", value: "active" },
    { label: "Total", value: "total" },
  ];
  if (seriesType.value === "departments") {
    fieldsList.push({ label: "Closed", value: "closed" });
  }
  return fieldsList;
});

const { chartData, isDataLoading } = useStatisticData({
  entity,
  periodType,
  period,
  periods,
  comparable,
  extraDeps: [seriesType],
});

watch(seriesType, () => {
  fields.value = seriesType.value === "admins" ? "responses" : "created";
});

watch([chartData, fields, comparable, seriesType], () => {
  const blocks = chartData.value;
  if (!blocks) return;
  const primary = blocks[0] || {};
  qualityText.value = primary.quality?.incomplete
    ? "Closed chats are incomplete for part of this period."
    : "";
  const field = fields.value;
  const nameField = seriesType.value === "clients" ? "owner" : seriesType.value === "admins" ? "user" : "dep";
  if (seriesType.value === "amount") {
    if (!comparable.value) {
      const built = buildSeriesFromTimeseries(primary.timeseries || [], field);
      series.value = built.series;
      categories.value = built.categories.map((item) => String(item).split("T")[0]);
      summary.value = built.summary;
      return;
    }
    const built = buildComparablePeriodSeries(blocks, periods.value, field);
    series.value = built.series;
    categories.value = built.categories.map((item) => String(item));
    summary.value = built.summary;
    return;
  }
  const built = buildGroupedSeries(
    primary.timeseries || [],
    nameField,
    field,
    (row, key) => Number(row?.[key]) || 0
  );
  const limited = limitTop(built.series, built.summary, 20);
  series.value = limited.series;
  summary.value = limited.summary;
  categories.value = built.categories.map((item) => String(item).split("T")[0]);
});
</script>
