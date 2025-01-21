<template>
  <widget title="Chats resposibles" :loading="isLoading" class="pa-0 ma-0">
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

      <v-divider></v-divider>

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
import { computed, onMounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import api from "@/api";
import apexchart from "vue-apexcharts";

const props = defineProps(["data"]);
const { data } = toRefs(props);

const emit = defineEmits(["update", "update:key"]);

const store = useStore();

const periods = ref(["day", "week", "month"]);
const statuses = ref(["all", "close", "open"]);
const accounts = ref({});
const isAccountsLoading = ref(false);

onMounted(() => fetchAccounts());

const dayChats = computed(() => store.getters["chats/dayChats"]);
const monthChats = computed(() => store.getters["chats/monthChats"]);
const weekChats = computed(() => store.getters["chats/weekChats"]);
const isLoading = computed(() => store.getters["chats/loading"]);

const chatsResponsibleStatistic = computed(() => {
  if (!data.value.period) {
    return new Map();
  }
  let chats;

  switch (data.value.period) {
    case "day": {
      chats = dayChats.value;
      break;
    }
    case "month": {
      chats = monthChats.value;
      break;
    }
    case "week": {
      chats = weekChats.value;
      break;
    }
  }

  const chatResponsibles = new Map();

  chats.forEach((chat) => {
    const { responsible } = chat;

    if (!responsible) {
      return;
    }

    let count = chatResponsibles.get(responsible) || 0;
    if (data.value.status === "all" || !data.value.status) {
      count++;
    } else if (data.value.status === "close" && chat.status === 3) {
      count++;
    } else if (data.value.status === "open" && chat.status !== 3) {
      count++;
    }

    chatResponsibles.set(responsible, count);
  });

  return chatResponsibles;
});

const options = computed(
  () =>
    !isAccountsLoading.value && {
      labels: [...chatsResponsibleStatistic.value.keys()]
        .filter((key) => !!chatsResponsibleStatistic.value.get(key))
        .map((key) => `${accounts.value[key]?.title}`),
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
    }
);
const series = computed(() =>
  [...chatsResponsibleStatistic.value.values()].filter((key) => !!key)
);

const setDefaultData = () => {
  if (Object.keys(data.value || {}).length === 0) {
    emit("update", { period: "week", status: "all" });
  }
};

const fetchAccounts = () => {
  [...chatsResponsibleStatistic.value.keys()].forEach(async (uuid) => {
    isAccountsLoading.value = true;
    try {
      if (!accounts.value[uuid]) {
        accounts.value[uuid] = api.accounts.get(uuid);
        accounts.value[uuid] = await accounts.value[uuid];
      }
    } catch {
      accounts.value[uuid] = undefined;
    } finally {
      isAccountsLoading.value = Object.values(accounts.value).some(
        (acc) => acc instanceof Promise
      );
    }
  });
};

watch(chatsResponsibleStatistic, fetchAccounts, { deep: true });

setDefaultData();
</script>

<script>
export default {
  name: "chats-responsibles-widget",
};
</script>
