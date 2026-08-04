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
    :not-comparable="seriesType === 'users'"
    :comparable="comparable"
    @input:comparable="comparable = $event"
  >
    <template v-slot:content>
      <default-chart
        description="Chat responses statistics"
        :type="type"
        :series="series"
        :categories="categories"
        :summary="summary"
        :legend-formatter="legendFormatter"
        :tooltip-formatter="seriesType === 'users' ? tooltipFormatter : null"
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
import { ref, toRefs, watch } from "vue";
import api from "@/api";
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

const series = ref([]);
const categories = ref([]);
const summary = ref({});
const accounts = ref({});
const seriesType = ref("amount");
const seriesTypes = [
  { label: "By users", value: "users" },
  { label: "Amount", value: "amount" },
];
const comparable = ref(true);

const { chartData, isDataLoading } = useStatisticData({
  entity: "ticket-responses",
  periodType,
  period,
  periods,
  comparable,
});

const legendFormatter = (name) => {
  const account = accounts.value[name] ?? { title: name };
  const total = summary.value[name];
  return `${account.title}${total ? `   ${total}` : ""}`;
};

const tooltipFormatter = (params) => {
  const list = Array.isArray(params) ? params : [params];
  const header = list[0]?.axisValueLabel ?? list[0]?.name ?? "";

  const rows = list
    .filter((p) => p.value != null && p.value !== "")
    .map(
      (p) =>
        `<div style="display:flex;align-items:center;gap:6px;margin:2px 0;">` +
        `<span style="width:8px;height:8px;border-radius:50%;background:${p.color};display:inline-block;"></span>` +
        `<span>${accounts.value[p.seriesName]?.title ?? p.seriesName}: <b>${p.value}</b></span></div>`
    )
    .join("");

  return `<div style="font-weight:600;margin-bottom:4px;">${header}</div>${rows}`;
};

watch(seriesType, () => {
  comparable.value = false;
});

watch([chartData, seriesType], async ([value]) => {
  if (!value || !value.length) {
    return;
  }

  const newSeries = [];
  const newCategories = [];
  summary.value = {};
  const tempData = JSON.parse(JSON.stringify(value));

  if (seriesType.value == "users") {
    tempData[0].timeseries?.forEach((timeseries) => {
      const current = tempData[0].timeseries.filter(
        (t) => t.ts === timeseries.ts
      );
      if (current.length <= 0 || newCategories.includes(timeseries.ts)) {
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
    const datas = [];
    tempData.forEach((_, index) => {
      const timeseries = [];

      tempData[index].timeseries.forEach((ts) => {
        const index = timeseries.findIndex((el) => ts.ts == el.ts);

        if (index !== -1) {
          timeseries[index].responses =
            (timeseries[index].responses || 0) + (ts.responses || 0);
        } else {
          timeseries.push(ts);
        }
      });

      datas.push({ timeseries: timeseries });
    });

    if (comparable.value) {
      const cmp = buildComparablePeriodSeries(
        datas,
        periods.value,
        "responses"
      );
      newSeries.push(...cmp.series);
      newCategories.push(...cmp.categories);
      summary.value = cmp.summary;
    } else {
      newSeries.push({
        name: "Responses",
        data: [],
      });

      datas[0].timeseries?.forEach((timeseries) => {
        newCategories.push(timeseries.ts);

        newSeries[0].data.push(timeseries.responses || 0);
      });

      summary.value = newSeries.reduce((acc, series) => {
        acc[series.name] = series.data.reduce((acc, v) => acc + v, 0);
        return acc;
      }, {});
    }
  }

  series.value = newSeries;
  categories.value = newCategories.map((c) => c.toString().split("T")[0]);
});
</script>
