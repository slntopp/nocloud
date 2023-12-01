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
      @update:options="setOptions"
      no-hide-uuid
      :itemsPerPageOptions="itemsPerPageOptions"
    >
      <template v-slot:[`item.totalPreview`]="{ item }">
        <v-chip :color="+item.total > 0 ? 'success' : 'error'">{{
          `${item.total} ${item.currency}`
        }}</v-chip>
      </template>
      <template v-slot:[`item.totalDefaultPreview`]="{ item }">
        <v-chip :color="+item.totalDefault > 0 ? 'success' : 'error'">{{
          item.totalDefault
            ? `${item.totalDefault} ${defaultCurrency}`
            : item.totalDefault
        }}</v-chip>
      </template>
      <template v-slot:[`item.exec`]="{ value }">
        <span>{{ new Date(value * 1000).toLocaleString() }}</span>
      </template>
      <template v-slot:[`item.paymentDate`]="{ value }">
        <div class="d-flex justify-center align-center">
          {{ value ? new Date(value * 1000).toLocaleString() : "-" }}
        </div>
      </template>
      <template v-slot:[`item.status`]="{ item }">
        <div class="d-flex justify-center align-center">
          {{ item.paymentDate ? "Paid" : "Unpaid" }}
        </div>
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
      <template v-slot:[`item.type`]="{ item }">
        <span>{{getReportType(item)}}</span>
      </template>
      <template v-slot:[`item.actions`]="{ item }">
        <div class="d-flex justify-center align-center">
          <v-btn
            v-for="action in getReportActions(item)"
            :key="action.title"
            @click="action.handler(item)"
            small
          >
            {{ action.title }}
          </v-btn>
        </div>
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
import { debounce } from "@/functions";

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

const emit = defineEmits(["input:unique", "input:duration"]);

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
    { text: "Payment date", value: "paymentDate" },
    { text: "Status", value: "status" },
    { text: "Total", value: "totalPreview" },
    { text: "Total in default currency", value: "totalDefaultPreview" },
    { text: "Product or resource", value: "item" },
    { text: "Type", value: "type" },
    { text: "Actions", value: "actions" },
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
  filters: filters.value,
  page: page.value,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy[0],
  sort: options.value.sortBy[0] && options.value.sortDesc[0] ? "DESC" : "ASC",
}));

const whmcsApi = computed(() => store.getters["settings/whmcsApi"]);
const transactionTypes = computed(() => store.getters["transactions/types"]);

const getReportActions = (report) => {
  const actions = [];

  if (report.type?.startsWith("invoice")) {
    actions.push({ title: "Email", handler: sendEmail });
  }
  return actions;
};

const getReportType=(item)=>{
  return transactionTypes.value.find(t=>t.key===item.type)?.title
}

const fetchReports = async () => {
  init();
  isFetchLoading.value = true;
  try {
    const { records: result } = await api.reports.list(requestOptions.value);

    reports.value = result.map((r) => {
      return {
        total: -r.total.toFixed(2),
        start: r.start,
        end: r.end,
        duration: `${new Date(r.start * 1000).toLocaleString()} - ${new Date(
          r.end * 1000
        ).toLocaleString()}`,
        exec: r.exec,
        transactionUuid: r.meta?.transaction,
        currency: r.currency,
        item: r.product || r.resource,
        uuid: r.uuid,
        service: r.service,
        instance: r.instance,
        account: r.account,
        type: r.meta?.transactionType,
        paymentDate: r.meta?.payment_date,
        totalDefault: -convertFrom(r.total, r.currency),
      };
    });
  } finally {
    isFetchLoading.value = false;
  }
};

const fetchReportsDebounced = debounce(fetchReports);

const setOptions = (newOptions) => {
  const sortByReplaceKeys = {
    totalPreview: "total",
    totalDefaultPreview: "total",
    duration: "start",
  };
  newOptions = {
    ...newOptions,
    sortBy: newOptions.sortBy.map((k) => sortByReplaceKeys[k] || k),
  };
  page.value = newOptions.page;
  if (JSON.stringify(newOptions) !== JSON.stringify(options.value)) {
    options.value = newOptions;
  }
};

const init = async () => {
  isCountLoading.value = true;
  try {
    const { total, unique } = await api.reports.count(requestOptions.value);
    count.value = +total;
    emit("input:unique", unique);
  } finally {
    isCountLoading.value = false;
  }
};

const getAccount = (value) => accounts.value.find((s) => s.uuid === value);
const getInstance = (value) => instances.value.find((s) => s.uuid === value);
const getService = (value) => services.value.find((s) => s.uuid === value);

const sendEmail = async (report) => {
  try {
    await fetch(
      /https:\/\/(.+?\.?\/)/.exec(whmcsApi.value)[0] +
        `modules/addons/nocloud/api/index.php?run=send_email&account=${report.account}&invoiceid=${report.transactionUuid}`
    );
    store.commit("snackbar/showSnackbarSuccess", {
      message: "Email resend successfully",
    });
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Error while try resend email",
    });
  }
};

watch(rates, () => {
  reports.value = reports.value.map((r) => ({
    ...r,
    totalDefault: convertFrom(r.total, r.currency),
  }));
});

watch(filters, fetchReportsDebounced, { deep: true });
watch(options, fetchReportsDebounced);
</script>
