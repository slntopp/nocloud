<template>
  <widget
    title="Transactions"
    :loading="isLoading"
    class="pa-0 ma-0"
    :more="{ name: 'Statistics', query: { tab: 'transactions' } }"
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
        <v-card-subtitle class="ma-0 pa-0">
          {{ periodCount }}
        </v-card-subtitle>
      </div>

      <div class="d-flex justify-space-between align-center mb-2">
        <v-card-subtitle class="ma-0 my-2 pa-0">Total created</v-card-subtitle>
        <v-card-subtitle class="ma-0 pa-0">
          {{ totalCount }}
        </v-card-subtitle>
      </div>

      <v-divider></v-divider>
      <v-list dense color="transparent">
        <v-list-item v-for="report in fiveLast" :key="report.uuid" class="px-0">
          <v-list-item-content class="ma-0 pa-0">
            <div class="d-flex align-center justify-space-between">
              <span>{{ formatSecondsToDate(report.exec, true) }}</span>
              <router-link
                v-if="!isAccountsLoading"
                target="_blank"
                :to="{
                  name: 'Account',
                  params: { accountId: report.account },
                }"
              >
                {{ getShortName(getAccount(report.account)?.title) }}
              </router-link>
              <v-skeleton-loader type="text" v-else />

              <balance-display
                class="ml-3"
                small
                :currency="report.currency"
                :value="report.total.toFixed(2)"
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
import { computed, onMounted, ref, toRefs, watch } from "vue";
import api from "@/api";
import {
  formatSecondsToDate,
  getShortName,
  getDatesPeriod,
} from "../../functions";
import BalanceDisplay from "@/components/balance.vue";
import { useStore } from "@/store";

const props = defineProps(["data"]);
const { data } = toRefs(props);

const emit = defineEmits(["update", "update:key"]);

const store = useStore();

const isLoading = ref(false);
const isAccountsLoading = ref(false);
const periods = ref(["day", "week", "month"]);

const fiveLast = ref([]);
const totalCount = ref();
const periodCount = ref();
const accounts = ref({});
const statisticForPeriod = ref();
const statisticParams = ref({});

onMounted(async () => {
  isLoading.value = true;
  try {
    const { records } = await api.reports.list({
      page: 1,
      limit: 5,
      sort: "DESC",
      field: "exec",
      filters: {},
    });

    fiveLast.value = records;
  } finally {
    isLoading.value = false;
  }
});

const isStatisticLoading = computed(() => store.getters["statistic/loading"]);

const getAccount = (uuid) => {
  return accounts.value[uuid];
};

const setDefaultData = () => {
  if (Object.keys(data.value || {}).length === 0) {
    emit("update", { period: "week" });
  }
};

setDefaultData();

watch(fiveLast, () => {
  fiveLast.value.forEach(async ({ account: uuid }) => {
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
});

watch(statisticForPeriod, (summary) => {
  totalCount.value = (summary?.total || 0).toString();
  periodCount.value = (summary?.created || 0).toString();
});

watch(
  () => data.value.period,
  (period) => {
    const [from, to] = getDatesPeriod(period);
    const dates = { from, to };

    dates.from = dates.from.toISOString().split("T")[0];
    dates.to = dates.to.toISOString().split("T")[0];

    statisticParams.value = {
      entity: "transactions",
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
  name: "transactions-widget",
};
</script>

<style scoped></style>
