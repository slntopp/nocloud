<template>
  <widget
    title="Instances"
    :loading="isLoading"
    :more="{ name: 'Statistics', query: { tab: 'instances' } }"
    class="pa-0 ma-0"
  >
    <v-card color="background-light" flat>
      <div class="d-flex justify-end">
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

      <div class="d-flex justify-space-between align-center">
        <v-card-subtitle class="ma-0 my-2 pa-0"
          >Created in last {{ data.period }}</v-card-subtitle
        >
        <v-card-subtitle v-if="countForPeriod" class="ma-0 pa-0">
          {{ countForPeriod }}
        </v-card-subtitle>
      </div>

      <div class="d-flex justify-space-between align-center mb-2">
        <v-card-subtitle class="ma-0 my-2 pa-0">Total</v-card-subtitle>
        <v-card-subtitle class="ma-0 pa-0">
          {{ instancesCount }}
        </v-card-subtitle>
      </div>

      <v-divider></v-divider>
      <v-list dense color="transparent">
        <v-list-item
          v-for="instance in lastInstances"
          :key="instance.uuid"
          class="px-0"
        >
          <v-list-item-content class="ma-0 pa-0">
            <div class="d-flex justify-space-between align-center">
              <router-link
                target="_blank"
                :to="{
                  name: 'Instance',
                  params: { instanceId: instance.uuid },
                  query: { fullscreen: viewport > 768 ? false : true },
                }"
              >
                {{ getShortName(instance.title) }}
              </router-link>
              <instance-state small :template="instance" />
            </div>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-card>
  </widget>
</template>

<script setup>
import widget from "@/components/widgets/widget.vue";
import { computed, onMounted, onUnmounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import { getShortName, getDatesPeriod } from "@/functions";
import InstanceState from "@/components/ui/instanceState.vue";

const props = defineProps(["data"]);
const { data } = toRefs(props);

const emit = defineEmits(["update", "update:key"]);

const store = useStore();

const isLoading = ref(false);
const periods = ref(["day", "week", "month"]);
const viewport = ref(window.innerWidth);
const statisticForPeriod = ref();
const countForPeriod = ref();
const lastInstances = ref([]);
const instancesCount = ref();
const statisticParams = ref({});

onMounted(async () => {
  isLoading.value = true;
  try {
    lastInstances.value = await store.dispatch("instances/fetch", {
      field: "created",
      limit: "5",
      page: "1",
      sort: "DESC",
    });
  } finally {
    isLoading.value = false;
  }

  window.addEventListener("resize", onResize);
});

onUnmounted(() => {
  window.removeEventListener("resize", onResize);
});

const isStatisticLoading = computed(() => store.getters["statistic/loading"]);

function onResize() {
  viewport.value = window.innerWidth;
}

const setDefaultData = () => {
  if (Object.keys(data.value || {}).length === 0) {
    emit("update", { period: "week" });
  }
};

setDefaultData();

watch(statisticForPeriod, (summary) => {
  countForPeriod.value = (summary?.created || 0).toString();
  instancesCount.value = (summary?.total || 0).toString();
});

watch(
  () => data.value.period,
  (period) => {
    const [from, to] = getDatesPeriod(period);
    const dates = { from, to };

    dates.from = dates.from.toISOString().split("T")[0];
    dates.to = dates.to.toISOString().split("T")[0];

    statisticParams.value = {
      entity: "services",
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

  statisticForPeriod.value = Object.keys(response.summary).reduce(
    (acc, key) => {
      acc.created += response.summary[key].created || 0;
      acc.total += response.summary[key].total || 0;
      return acc;
    },
    { created: 0, total: 0 }
  );
});
</script>

<script>
export default {
  name: "instances-widget",
};
</script>

<style scoped></style>
