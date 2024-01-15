<template>
  <widget title="Instances" :loading="isLoading" class="pa-0 ma-0">
    <div style="color: #fff">
      <div class="d-flex justify-end">
        <v-btn-toggle class="mt-2" dense v-model="period" borderless>
          <v-btn x-small :value="item" :key="item" v-for="item in periods">
            {{ item }}
          </v-btn>
        </v-btn-toggle>
      </div>

      <div class="d-flex justify-space-between align-center">
        <v-card-subtitle class="ma-0 my-2 pa-0"
          >Created in last {{ period }}</v-card-subtitle
        >
        <v-card-subtitle class="ma-0 pa-0">
          {{ countForPeriod }}
        </v-card-subtitle>
      </div>

      <div class="d-flex justify-space-between align-center mb-2">
        <v-card-subtitle class="ma-0 my-2 pa-0">Total created</v-card-subtitle>
        <v-card-subtitle class="ma-0 pa-0">
          {{ instances.length }}
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
                }"
              >
                {{ instance.title }}
              </router-link>
              <instance-state small :template="instance" />
            </div>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </div>
  </widget>
</template>

<script setup>
import widget from "@/components/widgets/widget.vue";
import { computed, onMounted, ref } from "vue";
import { useStore } from "@/store";
import {
  endOfDay,
  endOfMonth,
  endOfWeek,
  startOfDay,
  startOfMonth,
  startOfWeek,
} from "date-fns";
import InstanceState from "@/components/ui/instanceState.vue";

const store = useStore();

const isLoading = ref(false);
const period = ref("day");
const periods = ref(["day", "week", "month"]);

onMounted(async () => {
  isLoading.value = true;
  try {
    await store.dispatch("services/fetch");
  } catch (e) {
    console.log(e);
  } finally {
    isLoading.value = false;
  }
});

const instances = computed(() =>
  store.getters["services/getInstances"].map((i) => ({
    ...i,
    data: {
      ...(i?.data || {}),
      creation: new Date(i.data?.creation || 0).getTime() / 1000,
    },
  }))
);

const lastInstances = computed(() => {
  const sorted = [...instances.value].sort(
    (a, b) => +(b.data?.creation || 0) - (a.data?.creation || 0)
  );

  return sorted.slice(0, 5);
});
const countForPeriod = computed(() => {
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
  }).length;
});
</script>

<script>
export default {
  name: "instances-widget",
};
</script>

<style scoped></style>