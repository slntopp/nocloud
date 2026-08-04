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
    description="Instances statistics for period"
    :comparable="comparable"
    :not-comparable="seriesType !== 'amount'"
    @input:comparable="comparable = $event"
  >
    <template v-slot:content>
      <default-chart
        description="Instances statistics"
        :type="type"
        :series="series"
        :categories="categories"
        :summary="summary"
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
const seriesType = ref("amount");
const seriesTypes = [
  { label: "By types", value: "type" },
  { label: "Amount", value: "amount" },
];
const fields = ref("created");
const allFields = ref([
  { label: "Created", value: "created" },
  { label: "Active", value: "active" },
  { label: "Total", value: "total" },
]);
const comparable = ref(true);

const { chartData: data, isDataLoading } = useStatisticData({
  entity: "services",
  periodType,
  period,
  periods,
  comparable,
});

function switchFields(type, comparable) {
  if (type !== "type" && !comparable) {
    fields.value = [fields.value || ["created", "closed"]];
  } else if (comparable) {
    fields.value = fields.value[0] || "created";
  } else {
    fields.value = "created";
  }
}

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
    newSeries.push(
      ...Object.keys(tempData[0].summary || {}).map((key) => ({
        name: key,
        data: [],
      }))
    );

    tempData[0].timeseries?.forEach((timeseries) => {
      const current = tempData[0].timeseries.filter(
        (t) => t.ts === timeseries.ts
      );
      if (current.length <= 0 || newCategories.includes(timeseries.ts)) {
        return;
      }

      newCategories.push(timeseries.ts);

      current.map((ts) => {
        const index = newSeries.findIndex((series) => series.name === ts.type);
        newSeries[index].data.push(ts[fields.value] || 0);
      });
    });

    summary.value = Object.keys(tempData[0].summary || {}).reduce(
      (acc, key) => {
        acc[key] = tempData[0].summary[key][fields.value] || 0;
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
          current.reduce((acc, c) => acc + (c[series.id] || 0), 0) || 0
        );
      });
    });

    summary.value = {};
    newSeries.forEach((serie) => {
      summary.value[serie.name] =
        Object.keys(tempData[0].summary || {}).reduce(
          (acc, key) => acc + (tempData[0].summary[key][serie.id] || 0),
          0
        ) || 0;
    });
  } else {
    const datas = [];
    tempData.forEach((_, index) => {
      const timeseries = [];

      tempData[index].timeseries.forEach((ts) => {
        const index = timeseries.findIndex((el) => ts.ts == el.ts);

        if (index !== -1) {
          timeseries[index][fields.value] =
            (timeseries[index][fields.value] || 0) + (ts[fields.value] || 0);
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
