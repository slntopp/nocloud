<template>
  <statistic-item
    :period="period"
    :periodType="periodType"
    @input:period="period = $event"
    @input:period-type="periodType = $event"
    @input:type="type = $event"
    :loading="isDataLoading"
    :type="type"
    :all-fields="allFields"
    :fields="fields"
    @input:fields="fields = $event"
    :fields-multiple="seriesType === 'amount'"
    description="Instances income statistics for period"
  >
    <template v-slot:content>
      <default-chart
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
import { computed, ref, watch } from "vue";
import { useStore } from "@/store";
import { debounce } from "@/functions";
import DefaultChart from "@/components/statistics/defaultChart.vue";

const store = useStore();

const period = ref([]);
const type = ref("bar");
const periodType = ref("month");
const fields = ref(["revenue", "revenue_new"]);
const allFields = ref([
  { label: "Periodical payments", value: "revenue" },
  { label: "First payment", value: "revenue_new" },
  { label: "Total", value: "total" },
]);

const data = ref({});
const series = ref([]);
const categories = ref([]);
const summary = ref({});
const seriesType = ref("amount");

const seriesTypes = [
  { label: "By types", value: "type" },
  { label: "Amount", value: "amount" },
];

const isDataLoading = ref(false);

const defaultCurrency = computed(() => store.getters["currencies/default"]);

async function fetchData() {
  isDataLoading.value = true;

  try {
    const start_date = period.value[0].toISOString().split("T")[0];
    const end_date = period.value[1].toISOString().split("T")[0];

    const params = {
      entity: "services_revenue",
      params: {
        start_date,
        end_date,
        with_timeseries: true,
      },
    };

    const response = await store.dispatch("statistic/get", params);
    data.value = response;
  } finally {
    isDataLoading.value = false;
  }
}

const fetchDataDebounced = debounce(fetchData, 300);

function getPrice(c, id) {
  if (id === "total") {
    return (c.revenue || 0) + (c.revenue_new || 0);
  } else {
    return c[id] || 0;
  }
}

function getFormattedPrice(price) {
  return [price.toFixed(0), defaultCurrency.value.code].join("");
}

watch(period, () => {
  fetchDataDebounced();
});

watch(seriesType, (val) => {
  if (val === "type") {
    fields.value = fields.value[0] || "total";
  } else {
    fields.value = [fields.value || ["revenue", "revenue_new"]];
  }
});

watch([data, seriesType, fields], () => {
  if (!fields.value || !fields.value.length) {
    return;
  }

  const newSeries = [];
  const newCategories = [];

  const tempData = JSON.parse(JSON.stringify(data.value));

  if (seriesType.value === "amount") {
    fields.value.forEach((key) => {
      newSeries.push({
        name: allFields.value.find((field) => field.value === key).label,
        data: [],
        id: key,
      });
    });

    data.value.timeseries?.forEach((timeseries) => {
      const current = tempData.timeseries.filter((t) => t.ts === timeseries.ts);
      if (current.length <= 0) {
        return;
      }

      newCategories.push(timeseries.ts);

      newSeries.forEach((series) => {
        series.data.push(
          getFormattedPrice(
            current.reduce((acc, c) => acc + getPrice(c, series.id), 0) || 0
          )
        );
      });

      tempData.timeseries = tempData.timeseries.filter(
        (t) => t.ts !== timeseries.ts
      );
    });

    summary.value = {};
    newSeries.forEach((serie) => {
      summary.value[serie.name] = getFormattedPrice(
        Object.keys(data.value.summary || {}).reduce(
          (acc, key) => acc + getPrice(data.value.summary[key], serie.id),
          0
        ) || 0
      );
    });
  } else {
    data.value.timeseries?.forEach((timeseries) => {
      const current = tempData.timeseries.filter((t) => t.ts === timeseries.ts);
      if (current.length <= 0) {
        return;
      }

      newCategories.push(timeseries.ts);

      current.map((ts) => {
        let index = newSeries.findIndex((series) => series.name === ts.type);

        if (index == -1) {
          newSeries.push({ name: ts.type, data: [] });
          index = newSeries.length - 1;
        }

        newSeries[index].data.push(
          getFormattedPrice(getPrice(ts, fields.value) || 0)
        );
      });

      tempData.timeseries = tempData.timeseries.filter(
        (t) => t.ts !== timeseries.ts
      );
    });

    summary.value = Object.keys(data.value.summary || {}).reduce((acc, key) => {
      acc[key] = getFormattedPrice(
        getPrice(data.value.summary[key], fields.value) || 0
      );
      return acc;
    }, {});
  }

  series.value = newSeries;
  categories.value = newCategories.map((c) => c.split("T")[0]);
});
</script>
