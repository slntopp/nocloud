<template>
  <widget title="Instances count" :loading="isLoading" class="pa-0 ma-0">
    <v-card color="background-light" flat>
      <div class="d-flex justify-end">
        <v-btn-toggle
          class="mt-2"
          dense
          :value="period"
          @change="period = $event || period"
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
      ></apexchart>
      <div v-else class="d-flex justify-center align-center">
        <v-card-title>Instances not found</v-card-title>
      </div>
    </v-card>
  </widget>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import widget from "@/components/widgets/widget.vue";
import apexchart from "vue-apexcharts";
import { useStore } from "@/store";
import {
  endOfDay,
  endOfMonth,
  endOfWeek,
  startOfDay,
  startOfMonth,
  startOfWeek,
} from "date-fns";

const store = useStore();

const period = ref("day");
const periods = ref(["day", "week", "month"]);
const typesMap = ref(new Map());

const options = computed(() => ({
  labels: [...typesMap.value.keys()].map(
    (key) => `${key} - ${typesMap.value.get(key)}`
  ),
  theme: {
    palette: "palette8",
  },
}));
const series = computed(() => [...typesMap.value.values()]);

const isLoading = computed(() => store.getters["services/isLoading"]);

const instances = computed(() =>
  store.getters["services/getInstances"].map((inst) => ({
    ...inst,
    data: {
      ...(inst?.data || {}),
      creation: new Date(inst.data?.creation || 0).getTime() / 1000,
    },
  }))
);

const instancesForPeriod = computed(() => {
  const dates = { from: null, to: null };

  switch (period.value) {
    case "day": {
      dates.from = startOfDay(new Date());
      dates.to = endOfDay(new Date());
      break;
    }
    case "month": {
      dates.from = startOfMonth(new Date());
      dates.to = endOfMonth(new Date());
      break;
    }
    case "week": {
      dates.from = startOfWeek(new Date());
      dates.to = endOfWeek(new Date());
      break;
    }
  }

  dates.from = dates.from.getTime() / 1000;
  dates.to = dates.to.getTime() / 1000;

  return instances.value.filter((ac) => {
    const createDate = +ac.data?.creation || 0;

    return dates.from <= createDate && dates.to >= createDate;
  });
});

watch(
  instancesForPeriod,
  () => {
    const map = new Map();
    instancesForPeriod.value.forEach((inst) => {
      const key = inst.type;
      if (map.has(key)) {
        map.set(key, map.get(key) + 1);
      } else {
        map.set(key, 1);
      }
    });

    typesMap.value = map;
  },
  { deep: true }
);
</script>

<style>
span.apexcharts-legend-text {
  color: var(--v-primary-base) !important;
}
</style>