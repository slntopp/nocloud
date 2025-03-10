<template>
  <statistic-item
    :period="period"
    :periodType="periodType"
    @input:period="period = $event"
    @input:period-type="periodType = $event"
    :loading="isDataLoading"
    :type="type"
    @input:type="type = $event"
    :not-comparable="seriesType === 'users'"
    :comparable="comparable"
    @input:comparable="comparable = $event"
    :periods="periods"
    @input:periods="periods = $event"
  >
    <template v-slot:content>
      <default-chart
        description="Chat responses statistics"
        :type="type"
        :series="series"
        :categories="categories"
        :summary="summary"
        :custom-legend-formater="legendFomatter"
        :options="chartOptions"
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
import api from "@/api";
import DefaultChart from "@/components/statistics/defaultChart.vue";

const store = useStore();

const period = ref([]);
const periodType = ref("month");
const type = ref("bar");

const series = ref([]);
const categories = ref([]);
const summary = ref({});
const accounts = ref({});
const chartData = ref();
const seriesType = ref("amount");
const seriesTypes = [
  { label: "By users", value: "users" },
  { label: "Amount", value: "amount" },
];
const comparable = ref(false);
const periods = ref({ first: [], second: [] });

const isDataLoading = ref(false);

const chartOptions = computed(() => {
  if (seriesType.value === "users") {
    return {
      tooltip: {
        y: {
          title: {
            formatter: (seriesName) => accounts.value[seriesName]?.title,
          },
        },
      },
    };
  }
  return {};
});

function formatDate(date) {
  return date.toISOString().split("T")[0];
}

function legendFomatter(val, opts) {
  const account = accounts.value[val] ?? { title: val };

  return `${account.title} ${
    summary.value[series.value[opts.seriesIndex]?.name]
      ? summary.value[series.value[opts.seriesIndex]?.name]
      : ""
  }`;
}

async function fetchData() {
  isDataLoading.value = true;

  try {
    const params = {
      entity: "ticket-responses",
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
          datas[0]?.timeseries?.length || 0,
          datas[1]?.timeseries?.length || 0
        );
        index++
      ) {
        const first = datas[0]?.timeseries?.[index];
        const second = datas[1]?.timeseries?.[index];

        if (!newCategories.includes(index + 1)) {
          newCategories.push(index + 1);
        }

        newSeries[0].data.push(first?.responses || 0);
        newSeries[1].data.push(second?.responses || 0);
      }

      newSeries.forEach((serie) => {
        summary.value[serie.name] =
          serie.data.reduce((acc, a) => acc + a, 0) || 0;
      });
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
  categories.value = newCategories;
});
</script>
