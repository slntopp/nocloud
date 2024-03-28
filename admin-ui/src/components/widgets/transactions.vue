<template>
  <widget title="Transactions" :loading="isLoading" class="pa-0 ma-0">
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
          {{ dates[data.period] }}
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
                {{ getAccount(report.account)?.title }}
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
import { onMounted, ref, toRefs, watch } from "vue";
import api from "@/api";
import {
  endOfDay,
  endOfMonth,
  endOfWeek,
  startOfDay,
  startOfMonth,
  startOfWeek,
} from "date-fns";
import { formatSecondsToDate } from "../../functions";
import BalanceDisplay from "@/components/balance.vue";

const props = defineProps(["data"]);
const { data } = toRefs(props);

const emit = defineEmits(["update", "update:key"]);

const isLoading = ref(false);
const isAccountsLoading = ref(false);
const periods = ref(["day", "week", "month"]);

const fiveLast = ref([]);
const totalCount = ref(0);
const dates = ref({});
const accounts = ref({});

onMounted(async () => {
  isLoading.value = true;
  try {
    const [{ total }, { records }] = await Promise.all([
      api.reports.count({}),
      api.reports.list({
        page: 1,
        limit: 5,
        sort: "DESC",
        field: "exec",
        filters: {},
      }),
    ]);

    const data = await Promise.all(
      periods.value.map(async (period) => {
        return { period, data: await getCountForPeriod(period) };
      })
    );

    data.forEach(({ period, data: { total } }) => {
      dates.value[period] = total;
    });

    totalCount.value = total;
    fiveLast.value = records;
  } catch (e) {
    console.log(e);
  } finally {
    isLoading.value = false;
  }
});

const getCountForPeriod = (period) => {
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

  dates.from = Math.round(dates.from.getTime() / 1000);
  dates.to = Math.round(dates.to.getTime() / 1000);

  return api.reports.count({
    limit: 5,
    page: 1,
    field: "exec",
    sort: "DESC",
    filters: { exec: { ...dates } },
  });
};

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
</script>

<script>
export default {
  name: "transactions-widget",
};
</script>

<style scoped></style>
