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
    :fields-multiple="seriesType === 'amount' && !comparable"
    description="Instances statistics for period"
    :comparable="comparable"
    :not-comparable="seriesType !== 'amount'"
    @input:comparable="comparable = $event"
    :periods="periods"
    @input:periods="periods = $event"
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
import { ref, watch } from "vue";
import { useStore } from "@/store";
import { debounce } from "@/functions";
import DefaultChart from "@/components/statistics/defaultChart.vue";

const store = useStore();

const period = ref([]);
const type = ref("bar");
const periodType = ref("month");

const data = ref({});
const series = ref([]);
const categories = ref([]);
const summary = ref({});
const seriesType = ref("amount");
const seriesTypes = [
  { label: "By types", value: "type" },
  { label: "Amount", value: "amount" },
];
const fields = ref(["created"]);
const allFields = ref([
  { label: "Created", value: "created" },
  { label: "Active", value: "active" },
  { label: "Total", value: "total" },
]);
const comparable = ref(false);
const periods = ref({ first: [], second: [] });
const isDataLoading = ref(false);

function formatDate(date) {
  return date.toISOString().split("T")[0];
}

function switchFields(type, comparable) {
  if (type !== "type" && !comparable) {
    fields.value = [fields.value || ["created", "closed"]];
  } else if (comparable) {
    fields.value = fields.value[0] || "created";
  } else {
    fields.value = "created";
  }
}

async function fetchData() {
  isDataLoading.value = true;

  try {
    const params = {
      entity: "services",
      params: {
        with_timeseries: true,
      },
    };

    if (!comparable.value) {
      params.params.start_date = formatDate(period.value[0]);
      params.params.end_date = formatDate(period.value[1]);

      data.value = [await store.dispatch("statistic/get", params)];
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

      data.value = await Promise.all([
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
  if (!data.value) {
    fetchData();
  } else {
    fetchDataDebounced();
  }
});

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

      newSeries[0].data.push(first?.[fields.value] || 0);
      newSeries[1].data.push(second?.[fields.value] || 0);
    }

    newSeries.forEach((serie) => {
      summary.value[serie.name] =
        serie.data.reduce((acc, a) => acc + a, 0) || 0;
    });
  }

  series.value = newSeries;
  categories.value = newCategories;
});
</script>
