<template>
  <statistic-item
    :period="period"
    :periodType="periodType"
    @input:period="period = $event"
    @input:period-type="periodType = $event"
    :loading="isDataLoading"
    @input:type="type = $event"
    :type="type"
    :all-fields="allFields"
    :fields="fields"
    @input:fields="fields = $event"
    :fields-multiple="!comparable"
    :comparable="comparable"
    @input:comparable="comparable = $event"
    :periods="periods"
    @input:periods="periods = $event"
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
import { computed, ref, watch } from "vue";
import { useStore } from "@/store";
import { debounce } from "@/functions";
import DefaultChart from "@/components/statistics/defaultChart.vue";

const store = useStore();

const period = ref([]);
const periodType = ref("month");
const type = ref("bar");
const allFields = ref([
  { label: "Other invoices", value: "revenue" },
  { label: "Instance start", value: "revenue_start" },
  { label: "Instance renew", value: "revenue_renew" },
  { label: "Top-up balance", value: "revenue_balance" },
]);
const fields = ref(["revenue"]);

const series = ref([]);
const categories = ref([]);
const summary = ref({});

const isDataLoading = ref(false);
const chartData = ref();
const comparable = ref(false);
const periods = ref({ first: [], second: [] });
const defaultCurrency = computed(() => store.getters["currencies/default"]);

function formatDate(date) {
  return date.toISOString().split("T")[0];
}

async function fetchData() {
  isDataLoading.value = true;

  try {
    const params = {
      entity: "revenue",
      params: {
        with_timeseries: true,
      },
    };

    if (!comparable.value) {
      params.params.start_date = formatDate(period.value[0]);
      params.params.end_date = formatDate(period.value[1]);

      chartData.value = [await store.dispatch("statistic/get", params)];
    } else {
      const params1 = {
        ...params,
        params: {
          ...params.params,
          start_date: formatDate(periods.value.first[0]),
          end_date: formatDate(periods.value.first[1]),
        },
      };
      const params2 = {
        ...params,
        params: {
          ...params.params,
          start_date: formatDate(periods.value.second[0]),
          end_date: formatDate(periods.value.second[1]),
        },
      };

      chartData.value = await Promise.all([
        store.dispatch("statistic/get", params1),
        store.dispatch("statistic/get", params2),
      ]);
    }
  } finally {
    isDataLoading.value = false;
  }
}

const fetchDataDebounced = debounce(fetchData, 1000);

watch([period, periods, comparable], () => {
  if (!chartData.value) {
    fetchData();
  } else {
    fetchDataDebounced();
  }
});

watch(comparable, (val) => {
  if (val) {
    fields.value = "revenue";
  } else {
    fields.value = ["revenue"];
  }
});

function getFormattedPrice(price) {
  return [price.toFixed(0), defaultCurrency.value.code].join("");
}

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
        serie.data.push(getFormattedPrice(timeseries[serie.id] || 0));
      });
    });

    newSeries.forEach((serie) => {
      summary.value[serie.name] = getFormattedPrice(
        tempData[0].summary?.[serie.id] || 0
      );
    });
  } else {
    Object.keys(periods.value).forEach((key) => {
      newSeries.push({
        name: `${formatDate(periods.value[key][0])}/${formatDate(
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

      newSeries[0].data.push(first?.[fields.value] || 0);
      newSeries[1].data.push(second?.[fields.value] || 0);
    }

    newSeries.forEach((serie) => {
      summary.value[serie.name] = getFormattedPrice(
        serie.data.reduce((acc, a) => acc + a, 0) || 0
      );

      serie.data = serie.data.map((el) => getFormattedPrice(el));
    });
  }

  series.value = newSeries;
  categories.value = newCategories;
});
</script>
