<template>
  <widget
    title="Accounts"
    :loading="isLoading"
    class="pa-0 ma-0"
    :more="{ name: 'Statistics', query: { tab: 'accounts' } }"
  >
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
        <v-card-subtitle v-if="statisticForPeriod" class="ma-0 pa-0">
          {{ countForPeriod }}
        </v-card-subtitle>
      </div>

      <div class="d-flex justify-space-between align-center mb-2">
        <v-card-subtitle class="ma-0 my-2 pa-0"
          >Total active/Total</v-card-subtitle
        >
        <v-card-subtitle class="ma-0 pa-0">
          {{ activeAccountsCount }} / {{ accountsCount }}
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
                  query: { fullscreen: viewport > 768 ? false : true },
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
import { computed, onMounted, onUnmounted, ref, toRefs, watch } from "vue";
import BalanceDisplay from "@/components/balance.vue";
import api from "@/api";
import { getShortName, getDatesPeriod } from "@/functions";
import { useStore } from "@/store";

const props = defineProps(["data"]);
const { data } = toRefs(props);

const emit = defineEmits(["update", "update:key"]);

const store = useStore();

const isLoading = ref(false);
const periods = ref(["day", "week", "month"]);
const viewport = ref(window.innerWidth);
const statisticForPeriod = ref();
const countForPeriod = ref({});
const lastFiveAccounts = ref([]);
const activeAccountsCount = ref();
const accountsCount = ref();
const statisticParams = ref({});

onMounted(async () => {
  isLoading.value = true;
  try {
    const { pool } = await api.post("accounts", {
      page: 1,
      limit: 5,
      field: "data.date_create",
      sort: "DESC",
    });
    lastFiveAccounts.value = pool;
  } finally {
    isLoading.value = false;
  }

  window.addEventListener("resize", onResize);
});

onUnmounted(() => {
  window.removeEventListener("resize", onResize);
});

function onResize() {
  viewport.value = window.innerWidth;
}

const isStatisticLoading = computed(() => store.getters["statistic/loading"]);

const setDefaultData = () => {
  if (Object.keys(data.value || {}).length === 0) {
    emit("update", { period: "week" });
  }
};

setDefaultData();

watch(statisticForPeriod, (summary) => {
  countForPeriod.value = (summary?.created || 0).toString();
  activeAccountsCount.value = (summary?.active || 0).toString();
  accountsCount.value = (summary?.total || 0).toString();
});

watch(
  () => data.value.period,
  (period) => {
    const [from, to] = getDatesPeriod(period);
    const dates = { from, to };

    dates.from = dates.from.toISOString().split("T")[0];
    dates.to = dates.to.toISOString().split("T")[0];

    statisticParams.value = {
      entity: "accounts",
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

<script>
export default {
  name: "accounts-widget",
};
</script>

<style scoped></style>
