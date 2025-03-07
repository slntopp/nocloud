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
    description="Chats statistics for period"
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
const allFields = ref([
  { label: "Created", value: "created" },
  { label: "Closed", value: "closed" },
  { label: "Active", value: "active" },
  { label: "Total", value: "total" },
]);
const fields = ref(["created", "closed"]);

const seriesTypes = [
  { label: "By departments", value: "departments" },
  { label: "Amount", value: "amount" },
];

const isDataLoading = ref(false);

async function fetchData() {
  isDataLoading.value = true;

  try {
    const start_date = period.value[0].toISOString().split("T")[0];
    const end_date = period.value[1].toISOString().split("T")[0];

    const params = {
      entity: "tickets",
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

watch(period, () => {
  fetchDataDebounced();
});

watch(seriesType, (val) => {
  if (val === "departments") {
    fields.value = fields.value[0] || "created";
  } else {
    fields.value = [fields.value || ["created", "closed"]];
  }
});

watch([data, seriesType, fields], () => {
  if (Array.isArray(fields.value) && fields.value.length === 0) {
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

      for (const series of newSeries) {
        series.data.push(
          current.reduce((acc, c) => acc + (c[series.id] || 0), 0) || 0
        );
      }

      tempData.timeseries = tempData.timeseries.filter(
        (t) => t.ts !== timeseries.ts
      );
    });

    summary.value = {};
    newSeries.forEach((serie) => {
      summary.value[serie.name] =
        Object.keys(data.value.summary || {}).reduce(
          (acc, key) => acc + (data.value.summary[key][serie.id] || 0),
          0
        ) || 0;
    });
  } else {
    data.value.timeseries?.forEach((timeseries) => {
      const current = tempData.timeseries.filter((t) => t.ts === timeseries.ts);
      if (current.length <= 0) {
        return;
      }
      newCategories.push(timeseries.ts);
      current.map((ts) => {
        let index = newSeries.findIndex((series) => series.name === ts.dep);
        if (index == -1) {
          newSeries.push({ name: ts.dep, data: [] });
          index = newSeries.length - 1;
        }
        newSeries[index].data.push(ts[fields.value] || 0);
      });
      tempData.timeseries = tempData.timeseries.filter(
        (t) => t.ts !== timeseries.ts
      );
    });
    summary.value = Object.keys(data.value.summary || {}).reduce((acc, key) => {
      acc[key] = data.value.summary[key][fields.value] || 0;
      return acc;
    }, {});
  }

  series.value = newSeries;
  categories.value = newCategories;
});
</script>
