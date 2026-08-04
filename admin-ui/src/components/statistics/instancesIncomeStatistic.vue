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
    :fields-multiple="seriesType === 'amount' && !comparable"
    :comparable="comparable"
    :not-comparable="seriesType !== 'amount'"
    @input:comparable="comparable = $event"
  >
    <template v-slot:content>
      <default-chart
        description="Instances income statistics"
        :type="type"
        :series="series"
        :categories="categories"
        :summary="summary"
        :value-formatter="getFormattedPrice"
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
import { computed, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
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

const store = useStore();

const fields = ref("total");
const allFields = ref([
  { label: "Periodical payments", value: "revenue" },
  { label: "First payment", value: "revenue_new" },
  { label: "Total", value: "total" },
]);

const series = ref([]);
const categories = ref([]);
const summary = ref({});
const seriesType = ref("amount");

const seriesTypes = [
  { label: "By types", value: "type" },
  { label: "Amount", value: "amount" },
];

const comparable = ref(true);

const { chartData: data, isDataLoading } = useStatisticData({
  entity: "services_revenue",
  periodType,
  period,
  periods,
  comparable,
});

function switchFields(type, comparable) {
  if (type === "type") {
    fields.value = "total";
  } else if (comparable) {
    fields.value = "revenue";
  } else {
    fields.value = ["revenue", "revenue_new"];
  }
}

function getPrice(c, id) {
  if (id === "total") {
    return (c?.revenue || 0) + (c?.revenue_new || 0);
  } else {
    return c?.[id] || 0;
  }
}

function getFormattedPrice(price) {
  return [(price || 0).toFixed(0), defaultCurrency.value.code].join("");
}

const defaultCurrency = computed(() => store.getters["currencies/default"]);

watch(seriesType, (val) => {
  comparable.value = false;
  switchFields(val, false);
});

watch(comparable, (val) => {
  switchFields(seriesType.value, val);
});

watch([data, seriesType, fields], () => {
  if (!fields.value || !fields.value.length) {
    return;
  }

  const newSeries = [];
  const newCategories = [];

  const tempData = JSON.parse(JSON.stringify(data.value));

  if (seriesType.value === "type") {
    tempData[0].timeseries?.forEach((timeseries) => {
      const current = tempData[0].timeseries.filter(
        (t) => t.ts === timeseries.ts
      );
      if (current.length <= 0 || newCategories.includes(timeseries.ts)) {
        return;
      }

      newCategories.push(timeseries.ts);

      current.map((ts) => {
        let index = newSeries.findIndex((series) => series.name === ts.type);

        if (index == -1) {
          newSeries.push({ name: ts.type, data: [] });
          index = newSeries.length - 1;
        }

        newSeries[index].data.push(getPrice(ts, fields.value));
      });
    });

    summary.value = Object.keys(tempData[0].summary || {}).reduce(
      (acc, key) => {
        acc[key] = getPrice(tempData[0].summary[key], fields.value);
        return acc;
      },
      {}
    );
  } else if (!comparable.value) {
    fields.value.forEach((key) => {
      newSeries.push({
        name: allFields.value.find((field) => field.value === key).label,
        data: [],
        id: key,
      });
    });

    tempData[0].timeseries?.forEach((timeseries) => {
      const current = tempData[0].timeseries.filter(
        (t) => t.ts === timeseries.ts
      );
      if (current.length <= 0 || newCategories.includes(timeseries.ts)) {
        return;
      }

      newCategories.push(timeseries.ts);

      newSeries.forEach((series) => {
        series.data.push(
          current.reduce((acc, c) => acc + getPrice(c, series.id), 0)
        );
      });
    });

    summary.value = {};
    newSeries.forEach((serie) => {
      summary.value[serie.name] = Object.keys(tempData[0].summary || {}).reduce(
        (acc, key) => acc + getPrice(tempData[0].summary[key], serie.id),
        0
      );
    });
  } else {
    const datas = [];
    tempData.forEach((_, index) => {
      const timeseries = [];

      tempData[index].timeseries.forEach((ts) => {
        const index = timeseries.findIndex((el) => ts.ts == el.ts);

        if (index !== -1) {
          timeseries[index][fields.value] =
            (timeseries[index][fields.value] || 0) +
            (getPrice(ts, fields.value) || 0);
        } else {
          timeseries.push(ts);
        }
      });

      datas.push({ timeseries: timeseries });
    });

    const cmp = buildComparablePeriodSeries(datas, periods.value, fields.value);
    newSeries.push(...cmp.series);
    newCategories.push(...cmp.categories);
    summary.value = cmp.summary;
  }

  series.value = newSeries;
  categories.value = newCategories.map((c) => c.toString().split("T")[0]);
});
</script>
