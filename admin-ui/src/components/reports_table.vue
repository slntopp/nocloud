<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row>
      <div style="max-width: 300px" class="mx-3">
        <date-picker label="from" v-model="durationFilter.from" />
      </div>
      <div style="max-width: 300px" class="mx-3">
        <date-picker label="to" v-model="durationFilter.to" />
      </div>
    </v-row>
    <nocloud-table
      :table-name="tableName"
      @input="selectRecord"
      :headers="reportsHeaders"
      :items="reports"
      :loading="isLoading"
      :footer-error="fetchError"
      :server-items-length="count"
      :server-side-page="page"
      :show-select="false"
      sort-by="exec"
      sort-desc
      @update:options="onUpdateOptions"
      no-hide-uuid
      :itemsPerPageOptions="itemsPerPageOptions"
    >
      <template v-slot:[`item.totalPreview`]="{ item }">
        <span>{{ `${item.total} ${item.currency}` }}</span>
      </template>
      <template v-slot:[`item.totalDefaultPreview`]="{ item }">
        <span>{{
          item.totalDefault
            ? `${item.totalDefault} ${defaultCurrency}`
            : item.totalDefault
        }}</span>
      </template>
      <template v-slot:[`item.exec`]="{ value }">
        <span>{{ new Date(value * 1000).toLocaleString() }}</span>
      </template>
      <template v-slot:[`item.service`]="{ value }">
        <router-link
          v-if="getService(value)"
          :to="{ name: 'Service', params: { serviceId: value } }"
        >
          {{ getService(value).title }}
        </router-link>
      </template>
      <template v-slot:[`item.instance`]="{ value }">
        <router-link
          v-if="getInstance(value)"
          :to="{ name: 'Instance', params: { instanceId: value } }"
        >
          {{ getInstance(value).title }}
        </router-link>
      </template>
      <template v-slot:[`item.account`]="{ value }">
        <router-link
          v-if="getAccount(value)"
          :to="{ name: 'Account', params: { accountId: value } }"
        >
          {{ getAccount(value).title }}
        </router-link>
      </template>
    </nocloud-table>
  </v-card>
</template>

<script setup>
import { toRefs, ref, computed, watch } from "vue";
import api from "@/api";
import NocloudTable from "@/components/table.vue";
import { useStore } from "@/store";
import useRate from "@/hooks/useRate";
import DatePicker from "@/components/ui/datePicker.vue";

const props = defineProps({
  filters: { type: Object },
  hideInstance: { type: Boolean, default: false },
  hideAccount: { type: Boolean, default: false },
  hideService: { type: Boolean, default: false },
  selectRecord: { type: Function, default: () => {} },
  tableName: { type: String },
});
const { filters, hideInstance, hideService, hideAccount } = toRefs(props);

const store = useStore();
const { convertTo, fetchRate } = useRate();

const reports = ref([]);
const count = ref(10);
const page = ref(1);
const isFetchLoading = ref(false);
const isCountLoading = ref(false);
const fetchError = ref("");
const options = ref({});
const itemsPerPageOptions = ref([5, 10, 15, 25]);
const durationFilter = ref({ to: "", from: "" });
const rates = ref({});

const reportsHeaders = computed(() => {
  const headers = [
    { text: "Duration", value: "duration" },
    { text: "Executed date", value: "exec" },
    { text: "Total", value: "totalPreview" },
    { text: "Total in default currency", value: "totalDefaultPreview" },
    { text: "Product or resource", value: "item" },
  ];

  if (!hideAccount.value) {
    headers.push({ text: "Account", value: "account" });
  }
  if (!hideInstance.value) {
    headers.push({ text: "Instance", value: "instance" });
  }
  if (!hideService.value) {
    headers.push({ text: "Service", value: "service" });
  }

  return headers;
});

const defaultCurrency = computed(() => store.getters["currencies/default"]);
const allCurrencies = computed(() => {
  const currencies = new Map(
    reports.value.map((r) => [r.currency, r.currency])
  );
  return [...currencies.values()];
});

const instances = computed(() => store.getters["services/getInstances"]);
const services = computed(() => store.getters["services/all"]);
const accounts = computed(() => store.getters["accounts/all"]);

const isLoading = computed(() => {
  return isFetchLoading.value || isCountLoading.value;
});

const requestOptions = computed(() => ({
  page: page.value,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy[0],
  sort: options.value.sortBy[0] && options.value.sortDesc[0] ? "DESC" : "ASC",
  from: durationFilter.value.from
    ? new Date(durationFilter.value.from).getTime() / 1000
    : undefined,
  to: durationFilter.value.to
    ? new Date(durationFilter.value.to).getTime() / 1000
    : undefined,
  filters: filters.value,
}));

const onUpdateOptions = async (newOptions) => {
  setOptions(newOptions);
  page.value = newOptions.page;
  init();
  isFetchLoading.value = true;
  try {
    const { records: result } = await api.reports.list(requestOptions.value);

    reports.value = result.map((r) => {
      return {
        total: r.total.toFixed(2),
        start: r.start,
        end: r.end,
        duration: `${new Date(r.start * 1000).toLocaleString()} - ${new Date(
          r.end * 1000
        ).toLocaleString()}`,
        exec: r.exec,
        currency: r.currency,
        item: r.product || r.resource,
        uuid: r.uuid,
        service: r.service,
        instance: r.instance,
        account: r.account,
        totalDefault: rates.value[r.currency]
          ? convertTo(r.total, rates.value[r.currency])
          : null,
      };
    });
  } finally {
    isFetchLoading.value = false;
  }
};

const setOptions = (newOptions) => {
  const sortByReplaceKeys = {
    totalPreview: "total",
    totalDefaultPreview: "total",
    duration: "start",
  };
  options.value = {
    ...newOptions,
    sortBy: newOptions.sortBy.map((k) => sortByReplaceKeys[k] || k),
  };
};

const init = async () => {
  isCountLoading.value = true;
  try {
    count.value = +(await api.reports.count(requestOptions.value)).total;
  } finally {
    isCountLoading.value = false;
  }
};

const getAccount = (value) => accounts.value.find((s) => s.uuid === value);
const getInstance = (value) => instances.value.find((s) => s.uuid === value);
const getService = (value) => services.value.find((s) => s.uuid === value);

watch(rates, () => {
  reports.value = reports.value.map((r) => ({
    ...r,
    totalDefault: rates.value[r.currency]
      ? convertTo(r.total, rates.value[r.currency])
      : null,
  }));
});

watch(allCurrencies, () => {
  allCurrencies.value.map(async (c) => {
    if (rates.value[c]) {
      return;
    }

    rates.value[c] = "blocked";
    const rate = await fetchRate(c);
    rates.value = { ...rates.value, [c]: rate };
  });
});

watch(
  durationFilter,
  () => {
    onUpdateOptions(options.value);
  },
  { deep: true }
);

watch(
  filters,
  () => {
    onUpdateOptions(options.value);
  },
  { deep: true }
);
</script>
