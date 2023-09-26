<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row v-if="showDates">
      <div style="max-width: 300px" class="mx-3">
        <date-picker
          label="from"
          :value="duration.from"
          @input="emit('input:duration', { ...duration, from: $event })"
        />
      </div>
      <div style="max-width: 300px" class="mx-3">
        <date-picker
          label="to"
          :value="duration.to"
          @input="emit('input:duration', { ...duration, to: $event })"
        />
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
        <router-link :to="{ name: 'Service', params: { serviceId: value } }">
          {{ getService(value)?.title || value }}
        </router-link>
      </template>
      <template v-slot:[`item.instance`]="{ value }">
        <router-link :to="{ name: 'Instance', params: { instanceId: value } }">
          {{ getInstance(value)?.title || value }}
        </router-link>
      </template>
      <template v-slot:[`item.account`]="{ value }">
        <router-link :to="{ name: 'Account', params: { accountId: value } }">
          {{ getAccount(value)?.title || value }}
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
import DatePicker from "@/components/ui/datePicker.vue";
import useCurrency from "@/hooks/useCurrency";

const props = defineProps({
  filters: { type: Object },
  hideInstance: { type: Boolean, default: false },
  hideAccount: { type: Boolean, default: false },
  hideService: { type: Boolean, default: false },
  selectRecord: { type: Function, default: () => {} },
  tableName: { type: String },
  showDates: { type: Boolean, default: false },
  duration: { type: Object, default: () => ({ from: null, to: null }) },
});
const { filters, hideInstance, hideService, hideAccount, duration, showDates } =
  toRefs(props);

const emit = defineEmits(["input:duration"]);

const store = useStore();
const { rates, convertFrom, defaultCurrency } = useCurrency();

const reports = ref([]);
const count = ref(10);
const page = ref(1);
const isFetchLoading = ref(false);
const isCountLoading = ref(false);
const fetchError = ref("");
const options = ref({});
const itemsPerPageOptions = ref([5, 10, 15, 25]);

const reportsHeaders = computed(() => {
  const headers = [
    { text: "Duration", value: "duration" },
    { text: "Executed date", value: "exec" },
    { text: "Total", value: "totalPreview" },
    { text: "Total in default currency", value: "totalDefaultPreview" },
    { text: "Product or resource", value: "item" },
    { text: "Type", value: "type" },
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

const instances = computed(() => store.getters["services/getInstances"]);
const services = computed(() => store.getters["services/all"]);
const accounts = computed(() => store.getters["accounts/all"]);

const isLoading = computed(() => {
  return isFetchLoading.value || isCountLoading.value;
});

const requestOptions = computed(() => ({
  ...filters.value,
  page: page.value,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy[0],
  sort: options.value.sortBy[0] && options.value.sortDesc[0] ? "DESC" : "ASC",
  from: duration.value.from
    ? new Date(duration.value.from).getTime() / 1000
    : undefined,
  to: duration.value.to
    ? new Date(duration.value.to).getTime() / 1000
    : undefined,
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
        type: r.meta.transactionType,
        totalDefault: convertFrom(r.total, r.currency),
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
    totalDefault: convertFrom(r.total, r.currency),
  }));
});

watch(
  duration,
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
