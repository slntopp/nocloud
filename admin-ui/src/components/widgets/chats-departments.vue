<template>
  <widget
    title="Chats departments"
    :loading="isLoading || isStatisticLoading"
    class="pa-0 ma-0"
    :more="{ name: 'Statistics', query: { tab: 'chats' } }"
  >
    <v-card color="background-light" flat>
      <div class="d-flex justify-space-between">
        <v-btn-toggle
          class="mt-2"
          dense
          :value="data.status"
          @change="
            emit('update:key', { value: $event || data.status, key: 'status' })
          "
          borderless
        >
          <v-btn x-small :value="item" :key="item" v-for="item in statuses">
            {{ item }}
          </v-btn>
        </v-btn-toggle>

        <v-btn-toggle
          class="mt-2"
          dense
          :value="data.period"
          @change="
            emit('update:key', { value: $event || data.period, key: 'period' })
          "
          borderless
        >
          <v-btn x-small :value="item" :key="item" v-for="item in periods">
            {{ item }}
          </v-btn>
        </v-btn-toggle>
      </div>

      <apexchart
        v-if="options && series.length"
        type="donut"
        :options="options"
        :series="series"
        height="300px"
      ></apexchart>
      <div v-else class="d-flex justify-center align-center">
        <v-card-title>Responsibles not found</v-card-title>
      </div>
    </v-card>
  </widget>
</template>

<script setup>
import widget from "@/components/widgets/widget.vue";
import { computed, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import apexchart from "vue-apexcharts";
import { getDatesPeriod } from "@/functions";

const props = defineProps(["data"]);
const { data } = toRefs(props);

const emit = defineEmits(["update", "update:key"]);

const store = useStore();

const periods = ref(["day", "week", "month"]);
const statuses = ref(["total", "closed", "created"]);
const statisticParams = ref({});
const statisticForPeriod = ref({});
const chatsResponsibleStatistic = ref(new Map());

const isLoading = computed(() => store.getters["chats/loading"]);

const isStatisticLoading = computed(() => store.getters["statistic/loading"]);

const options = computed(() => ({
  labels: [...chatsResponsibleStatistic.value.keys()]
    .filter((key) => !!chatsResponsibleStatistic.value.get(key))
    .map((key) => key),
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
const series = computed(() =>
  [...chatsResponsibleStatistic.value.values()].filter((key) => !!key)
);

const setDefaultData = () => {
  if (Object.keys(data.value || {}).length === 0) {
    emit("update", { period: "week", status: "total" });
  }
};

setDefaultData();

watch(
  () => data.value.period,
  (period) => {
    const [from, to] = getDatesPeriod(period);
    const dates = { from, to };

    dates.from = dates.from.toISOString().split("T")[0];
    dates.to = dates.to.toISOString().split("T")[0];

    statisticParams.value = {
      entity: "tickets",
      params: {
        start_date: dates.from,
        end_date: dates.to,
      },
    };

    store.dispatch("statistic/fetch", statisticParams.value);
  },
  { deep: true }
);

watch(
  [isStatisticLoading, () => data.value.period, () => data.value.status],
  () => {
    const response = store.getters["statistic/cached"](statisticParams.value);

    if (response instanceof Promise || !response) {
      return (statisticForPeriod.value = null);
    }

    statisticForPeriod.value = Object.keys(response.summary || {});

    chatsResponsibleStatistic.value = Object.keys(
      response.summary || {}
    ).reduce((acc, uuid) => {
      let count = 0;

      if (data.value.status === "closed") {
        count = response.summary[uuid].closed || 0;
      } else if (data.value.status === "created") {
        count = response.summary[uuid].created || 0;
      } else {
        count = response.summary[uuid].total || 0;
      }

      acc.set(uuid, count);
      return acc;
    }, new Map());
  }
);
</script>

<script>
export default {
  name: "chats-departments-widget",
};
</script>
