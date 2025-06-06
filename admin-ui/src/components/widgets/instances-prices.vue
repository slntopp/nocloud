<template>
  <widget
    title="Instances income"
    :loading="isStatisticLoading"
    :more="{ name: 'Statistics', query: { tab: 'instances-income' } }"
    class="pa-0 ma-0"
  >
    <v-card color="background-light" flat>
      <div class="d-flex justify-space-between">
        <v-btn-toggle
          class="mt-2"
          dense
          :value="data.type"
          @change="
            emit('update:key', { value: $event || data.type, key: 'type' })
          "
          borderless
        >
          <v-btn x-small :value="item" :key="item" v-for="item in types">
            {{ item }}
          </v-btn>
        </v-btn-toggle>

        <v-btn-toggle
          class="mt-2"
          dense
          :value="data.period"
          @change="
            emit('update:key', { key: 'period', value: $event || data.period })
          "
          borderless
        >
          <v-btn x-small :value="item" :key="item" v-for="item in periods">
            {{ item }}
          </v-btn>
        </v-btn-toggle>
      </div>
      <apexchart
        v-if="options.labels.length"
        type="donut"
        :options="options"
        :series="series"
        height="300px"
      ></apexchart>
      <div v-else class="d-flex justify-center align-center">
        <v-card-title>Instances not found</v-card-title>
      </div>
    </v-card>
  </widget>
</template>

<script setup>
import { computed, ref, toRefs, watch } from "vue";
import widget from "@/components/widgets/widget.vue";
import apexchart from "vue-apexcharts";
import { useStore } from "@/store";
import { getDatesPeriod } from "@/functions";

const props = defineProps(["data"]);
const { data } = toRefs(props);

const emit = defineEmits(["update", "update:key"]);

const store = useStore();

const periods = ref(["day", "week", "month"]);
const types = ref(["total", "renew", "new"]);
const typesMap = ref(new Map());
const statisticForPeriod = ref();
const statisticParams = ref({});

const options = computed(() => ({
  labels: [...typesMap.value.keys()].map(
    (key) =>
      `${key} - ${typesMap.value.get(key)} ${defaultCurrency.value?.title}`
  ),
  theme: {
    palette: "palette8",
  },
  plotOptions: {
    pie: {
      donut: {
        size: "0%",
      },
    },
  },
}));

const series = computed(() => [...typesMap.value.values()]);

const defaultCurrency = computed(() => store.getters["currencies/default"]);

const isStatisticLoading = computed(() => store.getters["statistic/loading"]);

const setDefaultData = () => {
  if (Object.keys(data.value || {}).length === 0) {
    emit("update", { period: "week", type: "new" });
  }
};

setDefaultData();

watch([statisticForPeriod, () => data.value.type], ([summary]) => {
  if (!summary) {
    return;
  }

  typesMap.value = Object.keys(summary || {}).reduce((acc, key) => {
    let price = 0;

    if (data.value.type === "new") {
      price = summary[key].revenue_new || 0;
    } else if (data.value.type === "renew") {
      price = summary[key].revenue || 0;
    } else {
      price = (summary[key].revenue || 0) + (summary[key].revenue_new || 0);
    }

    acc.set(key, +price.toFixed(0));
    return acc;
  }, new Map());
});

watch(
  () => data.value.period,
  (period) => {
    const [from, to] = getDatesPeriod(period);
    const dates = { from, to };

    dates.from = dates.from.toISOString().split("T")[0];
    dates.to = dates.to.toISOString().split("T")[0];

    statisticParams.value = {
      entity: "services_revenue",
      params: {
        start_date: dates.from,
        end_date: dates.to,
      },
    };

    store.dispatch("statistic/fetch", statisticParams.value);
  },
  { deep: true }
);

watch([isStatisticLoading, () => data.value.period], () => {
  const response = store.getters["statistic/cached"](statisticParams.value);

  if (response instanceof Promise || !response) {
    return (statisticForPeriod.value = null);
  }

  statisticForPeriod.value = response.summary;
});
</script>

<style>
span.apexcharts-legend-text {
  color: var(--v-primary-base) !important;
}
</style>
