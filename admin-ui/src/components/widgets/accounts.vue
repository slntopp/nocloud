<template>
  <widget title="Accounts" :loading="isLoading" class="pa-0 ma-0">
    <v-card color="background-light" flat>
      <div class="d-flex justify-end">
        <v-btn-toggle
          class="mt-2"
          dense
          :value="data.period"
          @change="
            $emit('update:key', { key: 'period', value: $event || data.period })
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
          >Created this {{ data.period }}</v-card-subtitle
        >
        <v-card-subtitle v-if="!isCountForPeriodLoading" class="ma-0 pa-0">
          {{ countsForPeriod[data.period] }}
        </v-card-subtitle>
      </div>

      <div class="d-flex justify-space-between align-center mb-2">
        <v-card-subtitle class="ma-0 my-2 pa-0">Total active</v-card-subtitle>
        <v-card-subtitle v-if="activeAccountsCount" class="ma-0 pa-0">
          {{ activeAccountsCount }}
        </v-card-subtitle>
      </div>

      <v-divider></v-divider>
      <v-list dense color="transparent">
        <v-list-item
          v-for="account in lastFiveAccounts"
          :key="account.uuid"
          class="px-0"
        >
          <v-list-item-content class="ma-0 pa-0">
            <div class="d-flex justify-space-between align-center">
              <router-link
                target="_blank"
                :to="{
                  name: 'Account',
                  params: { accountId: account?.uuid },
                  query: { fullscreen: (viewport > 768) ? false : true }
                }"
              >
                {{ getShortName(account.title) }}
              </router-link>
              <balance-display
                small
                :currency="account.currency"
                :value="account.balance"
              />
            </div>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-card>
  </widget>
</template>

<script setup>
import widget from "@/components/widgets/widget.vue";
import { onMounted, onUnmounted, ref, toRefs, watch } from "vue";
import {
  endOfDay,
  endOfMonth,
  endOfWeek,
  startOfDay,
  startOfMonth,
  startOfWeek,
} from "date-fns";
import BalanceDisplay from "@/components/balance.vue";
import api from "@/api";
import { getShortName } from "@/functions";

const props = defineProps(["data"]);
const { data } = toRefs(props);

const emit = defineEmits(["update", "update:key"]);

const isLoading = ref(false);
const isCountForPeriodLoading = ref(false);
const periods = ref(["day", "week", "month"]);
const viewport = ref(window.innerWidth);

onMounted(async () => {
  isLoading.value = true;
  try {
    const { pool, active } = await api.post("accounts", {
      page: 1,
      limit: 5,
      field: "data.date_create",
      sort: "DESC",
    });
    lastFiveAccounts.value = pool;
    activeAccountsCount.value = active;
  } finally {
    isLoading.value = false;
  }

  window.addEventListener("resize", onResize)
});

onUnmounted(() => {
  window.removeEventListener("resize", onResize)
})

function onResize() {
  viewport.value = window.innerWidth
}

const countsForPeriod = ref({});
const lastFiveAccounts = ref([]);
const activeAccountsCount = ref();

const getCountForPeriod = async (period) => {
  const dates = { from: null, to: null };
  switch (period) {
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

  dates.from = Math.floor(dates.from.getTime() / 1000);
  dates.to = Math.floor(dates.to.getTime() / 1000);

  try {
    isCountForPeriodLoading.value = true;
    const { count } = await api.post("accounts", {
      limit: 0,
      page: 1,
      filters: {
        "data.date_create": dates,
      },
    });
    countsForPeriod.value[period] = count;
  } finally {
    isCountForPeriodLoading.value = false;
  }
};

const setDefaultData = () => {
  if (Object.keys(data.value || {}).length === 0) {
    emit("update", { period: "week" });
  }
};

setDefaultData();

watch(
  () => data.value.period,
  () => {
    if (!countsForPeriod.value[data.value.period]) {
      getCountForPeriod(data.value.period);
    }
  }
);
</script>

<script>
export default {
  name: "accounts-widget",
};
</script>

<style scoped></style>
