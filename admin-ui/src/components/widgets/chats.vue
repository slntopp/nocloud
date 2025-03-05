<template>
  <widget
    title="Chats"
    :loading="isLoading || isStatisticLoading"
    :more="{ name: 'Statistics', query: { tab: 'chats' } }"
    class="pa-0 ma-0"
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

      <div class="d-flex justify-space-between align-center">
        <v-card-subtitle class="ma-0 my-2 pa-0"
          >Created in last {{ data.period }}</v-card-subtitle
        >
        <v-card-subtitle class="ma-0 pa-0">
          {{ countForPeriod }}
        </v-card-subtitle>
      </div>

      <v-divider></v-divider>
      <v-card
        v-for="chat in lastActivityChats"
        :key="chat.uuid"
        dense
        color="transparent"
        class="d-flex justify-space-between pa-3"
        style="margin: 3px"
      >
        <div class="d-flex flex-column">
          <div>
            <span>Topic: </span>
            <router-link
              target="_blank"
              :to="{
                name: 'Chat',
                params: { uuid: chat.uuid },
              }"
            >
              {{ getShortName(chat.topic) }}
            </router-link>
          </div>

          <div class="d-flex">
            <span>Account: </span>
            <router-link
              v-if="!isAccountsLoading"
              target="_blank"
              :to="{
                name: 'Account',
                params: { accountId: chat.owner },
              }"
            >
              {{ accounts[chat.owner]?.title }}
            </router-link>
            <v-skeleton-loader type="text" v-else />
          </div>
        </div>
        <div class="d-flex flex-column">
          <span class="d-flex justify-end">
            {{
              formatSecondsToDate(
                Number(
                  chat.meta.lastMessage?.edited ||
                    chat.meta.lastMessage?.created ||
                    chat.created
                ) / 1000,
                true
              )
            }}
          </span>
          <span class="d-flex justify-end">
            {{ chat.department || "none" }}
          </span>
        </div>
      </v-card>
    </v-card>
  </widget>
</template>

<script setup>
import widget from "@/components/widgets/widget.vue";
import { computed, onMounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import api from "@/api";
import { formatSecondsToDate, getShortName, getDatesPeriod } from "@/functions";

const props = defineProps(["data"]);
const { data } = toRefs(props);

const emit = defineEmits(["update", "update:key"]);

const store = useStore();

const periods = ref(["day", "week", "month"]);
const statuses = ref(["total", "closed", "created"]);
const accounts = ref({});
const isAccountsLoading = ref(false);
const statisticForPeriod = ref();
const statisticParams = ref({});

onMounted(() => fetchAccounts());

const dayChats = computed(() => store.getters["chats/dayChats"]);
const isLoading = computed(() => store.getters["chats/loading"]);

const isStatisticLoading = computed(() => store.getters["statistic/loading"]);

const lastActivityChats = computed(() => {
  return dayChats.value.slice(0, 3);
});

const countForPeriod = computed(() => {
  if (!statisticForPeriod.value) {
    return 0;
  }

  if (data.value.status == "closed") {
    return statisticForPeriod.value.closed;
  }

  if (data.value.status == "open") {
    return statisticForPeriod.value.created;
  }

  return statisticForPeriod.value.total;
});

const setDefaultData = () => {
  if (Object.keys(data.value || {}).length === 0) {
    emit("update", { period: "week", status: "total" });
  }
};

const fetchAccounts = () => {
  lastActivityChats.value.forEach(async ({ owner: uuid }) => {
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

watch(lastActivityChats, fetchAccounts);

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

watch([isStatisticLoading, () => data.value.period], () => {
  const response = store.getters["statistic/cached"](statisticParams.value);

  if (response instanceof Promise || !response) {
    return (statisticForPeriod.value = null);
  }

  statisticForPeriod.value = Object.keys(response.summary).reduce(
    (acc, key) => {
      acc.created += response.summary[key].created || 0;
      acc.total += response.summary[key].total || 0;
      acc.closed += response.summary[key].closed || 0;
      return acc;
    },
    { created: 0, total: 0, closed: 0 }
  );
});
</script>

<script>
export default {
  name: "chats-widget",
};
</script>
