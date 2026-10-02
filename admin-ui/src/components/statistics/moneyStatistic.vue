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
    :all-fields="[]"
    :comparable="comparable"
    @input:comparable="comparable = $event"
    description="Money"
  >
    <template v-slot:options>
      <v-select
        style="width: 260px"
        item-text="label"
        item-value="value"
        :items="moneyTypes"
        v-model="moneyType"
        hide-details
      />
      <v-select
        style="width: 180px"
        item-text="label"
        item-value="value"
        :items="serviceItems"
        v-model="serviceType"
        hide-details
      />
    </template>
    <template v-slot:content>
      <div v-if="qualityText" class="px-4 pb-2">{{ qualityText }}</div>
      <default-chart
        description="Money"
        :type="type"
        :series="series"
        :categories="categories"
        :summary="chartSummary"
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
  limitTop,
} from "@/hooks/useStatisticChart";

const props = defineProps(statisticPeriodProps);
const { period, periodType, periods, type } = toRefs(props);
const emit = defineEmits(statisticPeriodEmits);

const comparable = ref(false);
const moneyType = ref("paid:all");
const serviceType = ref("");
const series = ref([]);
const categories = ref([]);
const chartSummary = ref({});
const qualityText = ref("");
const knownServices = ref([]);

const moneyTypes = [
  { label: "Paid, total", value: "paid:all" },
  { label: "Paid, new orders", value: "paid:INSTANCE_START" },
  { label: "Paid, renewals", value: "paid:INSTANCE_RENEWAL" },
  { label: "Paid, balance top-up", value: "paid:BALANCE" },
  { label: "Issued, total", value: "issued:all" },
  { label: "Issued, new orders", value: "issued:INSTANCE_START" },
  { label: "Issued, renewals", value: "issued:INSTANCE_RENEWAL" },
  { label: "Issued, balance top-up", value: "issued:BALANCE" },
];

const requestParams = computed(() => {
  const [kind, invoiceType] = moneyType.value.split(":");
  return {
    kind,
    invoice_type: invoiceType,
    service_type: serviceType.value || undefined,
  };
});

const { chartData, isDataLoading } = useStatisticData({
  entity: "money",
  periodType,
  period,
  periods,
  comparable,
  params: requestParams,
  extraDeps: [moneyType, serviceType],
});

const serviceItems = computed(() => [
  { label: "All service types", value: "" },
  ...knownServices.value.map((name) => ({ label: name, value: name })),
]);

watch([chartData, serviceType, comparable], () => {
  const blocks = chartData.value;
  if (!blocks) return;
  const primary = blocks[0] || { timeseries: [], summary: {} };
  const note = [];
  if (primary.summary?.incomplete || primary.summary?.skipped) {
    note.push("Some invoices were skipped because a currency rate was missing.");
  }
  if (primary.summary?.returned_without_date) {
    note.push(
      `${primary.summary.returned_without_date} returned invoices have no return date.`
    );
  }
  if (primary.summary?.paid_without_date) {
    note.push(
      `${primary.summary.paid_without_date} paid invoices have no payment date.`
    );
  }
  qualityText.value = note.join(" ");

  if (!serviceType.value) {
    const names = new Set(knownServices.value);
    (blocks || []).forEach((block) => {
      Object.keys(block?.summary?.by_service || {}).forEach((name) => names.add(name));
    });
    knownServices.value = [...names].sort();
  }

  if (serviceType.value || comparable.value) {
    const field = "amount";
    if (!comparable.value) {
      series.value = [
        {
          name: "Amount",
          data: (primary.timeseries || []).map((row) => Number(row.amount) || 0),
        },
      ];
      categories.value = (primary.timeseries || []).map((row) =>
        String(row.ts).split("T")[0]
      );
      chartSummary.value = { Amount: Number(primary.summary?.amount) || 0 };
      return;
    }
    const built = buildComparablePeriodSeries(blocks, periods.value, field);
    series.value = built.series;
    categories.value = built.categories.map((item) => String(item));
    chartSummary.value = built.summary;
    return;
  }

  const names = new Set();
  (primary.timeseries || []).forEach((row) => {
    Object.keys(row.by_service || {}).forEach((name) => names.add(name));
  });
  let nextSeries = [...names].map((name) => ({
    name,
    data: (primary.timeseries || []).map((row) => Number(row.by_service?.[name]) || 0),
  }));
  let nextSummary = {};
  nextSeries.forEach((serie) => {
    nextSummary[serie.name] = serie.data.reduce((acc, value) => acc + value, 0);
  });
  const limited = limitTop(nextSeries, nextSummary, 10);
  series.value = limited.series;
  chartSummary.value = limited.summary;
  categories.value = (primary.timeseries || []).map((row) => String(row.ts).split("T")[0]);
});
</script>
