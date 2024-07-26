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
    >
      <template v-slot:[`item.totalPreview`]="{ item }">
        <v-chip>{{ `${item.total} ${item.currency?.title}` }}</v-chip>
      </template>
      <template v-slot:[`item.totalDefaultPreview`]="{ item }">
        <v-chip>{{
          item.totalDefault
            ? `${item.totalDefault} ${defaultCurrency?.title}`
            : item.totalDefault
        }}</v-chip>
      </template>
      <template v-slot:[`item.exec`]="{ value }">
        <span>{{ formatSecondsToDate(value, true) }}</span>
      </template>
      <template v-slot:[`item.meta.payment_date`]="{ value }">
        <div class="d-flex justify-center align-center">
          {{ value ? formatSecondsToDate(value, true) : "-" }}
        </div>
      </template>
      <template v-slot:[`item.status`]="{ item }">
        <div class="d-flex justify-center">
          <v-chip :color="getStatusColor(item)">
            {{ getStatus(item) }}
          </v-chip>
        </div>
      </template>
      <template v-slot:[`item.service`]="{ value }">
        <router-link :to="{ name: 'Service', params: { serviceId: value } }">
          {{ getShortName(getService(value)?.title || value) }}
        </router-link>
      </template>
      <template v-slot:[`item.instance`]="{ value }">
        <router-link
          v-if="!isInstancesLoading"
          :to="{ name: 'Instance', params: { instanceId: value } }"
        >
          {{ getShortName(getInstance(value)?.title || value) }}
        </router-link>

        <v-skeleton-loader type="text" v-else />
      </template>
      <template v-slot:[`item.meta.transactionType`]="{ item }">
        <span>{{ getReportType(item) }}</span>
      </template>
      <template v-slot:[`item.actions`]="{ item }">
        <div class="d-flex justify-start align-center">
          <v-btn
            v-for="action in getReportActions(item)"
            :key="action.title"
            @click="callAction(item, action.action)"
            small
            :disabled="
              !!runningActionName &&
              (runningActionName !== action.action ||
                runningActionReportUuid !== item.uuid)
            "
            :loading="
              !!runningActionName &&
              runningActionName === action.action &&
              runningActionReportUuid === item.uuid
            "
            class="mx-1"
            :icon="!!action.icon"
          >
            <span v-if="!action.icon">
              {{ action.title }}
            </span>
            <v-icon v-else>{{ action.icon }}</v-icon>
          </v-btn>
        </div>
      </template>
      <template v-slot:[`item.account`]="{ value }">
        <router-link
          v-if="!isAccountsLoading"
          :to="{ name: 'Account', params: { accountId: value } }"
        >
          {{ getShortName(getAccount(value)?.title || value) }}
        </router-link>
        <v-skeleton-loader type="text" v-else />
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
import { debounce, formatSecondsToDate, getShortName } from "@/functions";
import { useRouter } from "vue-router/composables";

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
const router = useRouter();
const { rates, convertFrom, defaultCurrency } = useCurrency();

const reports = ref([]);
const count = ref(10);
const page = ref(1);
const isFetchLoading = ref(false);
const isCountLoading = ref(false);
const runningActionName = ref("");
const runningActionReportUuid = ref("");
const fetchError = ref("");
const options = ref({});
const accounts = ref({});
const isAccountsLoading = ref(false);
const isInstancesLoading = ref(false);

const reportsHeaders = computed(() => {
  const headers = [
    { text: "Duration", value: "duration", sortable: false },
    { text: "Executed date", value: "exec" },
    { text: "Payment date", value: "meta.payment_date" },
    { text: "Status", value: "status", sortable: false },
    { text: "Total", value: "totalPreview" },
    {
      text: "Total in default currency",
      value: "totalDefaultPreview",
      sortable: false,
    },
    { text: "Product or resource", value: "item", sortable: false },
    { text: "Type", value: "meta.transactionType" },
    { text: "Actions", value: "actions", sortable: false },
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

const services = computed(() => store.getters["services/all"]);

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
  const status = getStatus(report);
  switch (status) {
    case "Paid": {
      if (report.meta.transactionType?.startsWith("invoice")) {
        actions.push({
          icon: "mdi-alpha-w-box",
          action: "invoice",
          handler: downloadInvoice,
        });
      }
      break;
    }
    case "Unpaid": {
      if (report.meta.transactionType?.startsWith("invoice")) {
        actions.push({
          icon: "mdi-alpha-w-box",
          action: "invoice",
          handler: downloadInvoice,
        });
        actions.push({ title: "Email", action: "email", handler: sendEmail });
      }
      break;
    }
  }

  if (report.meta.transactionType?.startsWith("invoice")) {
    actions.push({
      title: "Open",
      action: "open",
      handler: openInvoice,
    });
  }

  return actions;
};

const getReportType = (item) => {
  const { transactionType: type } = item.meta;

  if (!type) {
    return "Unknown";
  }

  return (
    transactionTypes.value.find((t) => t.key === type)?.title ||
    type.charAt(0).toUpperCase() + type.slice(1)
  );
};

const fetchReports = async () => {
  init();
  isFetchLoading.value = true;
  fetchError.value = "";
  try {
    const { records: result } = await api.reports.list(requestOptions.value);
    reports.value = result.map((r) => {
      return {
        total: -r.cost.toFixed(2),
        start: r.start,
        end: r.end,
        duration: `${formatSecondsToDate(
          r.start,
          true
        )} - ${formatSecondsToDate(r.end, true)}`,
        exec: r.exec,
        transactionUuid: r.meta?.transaction,
        currency: r.currency,
        item: r.product || r.resource,
        uuid: r.uuid,
        service: r.service,
        instance: r.instance,
        account: r.account,
        meta: {
          ...r.meta,
        },
        totalDefault: -convertFrom(r.cost, r.currency),
      };
    });
  } catch (e) {
    fetchError.value = e.message;
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

const getAccount = (value) => accounts.value[value];
const getInstance = (value) => store.getters["instances/cached"].get(value);
const getService = (value) => services.value.find((s) => s.uuid === value);

const getStatus = (item) => {
  if (item.meta.status) {
    return item.meta.status;
  }
  return item.meta.payment_date ? "Paid" : "Unpaid";
};

const getStatusColor = (item) => {
  return {
    Paid: "success",
    Unpaid: "error",
    Terminate: "blue-grey darken-2",
    Draft: "blue",
    Cancelled: "warning",
  }[getStatus(item)];
};

const callAction = async (report, action) => {
  runningActionName.value = action;
  runningActionReportUuid.value = report.uuid;
  try {
    const actions = getReportActions(report);
    await actions.find((a) => a.action === action).handler(report);
  } finally {
    runningActionName.value = "";
    runningActionReportUuid.value = "";
  }
};

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

const downloadInvoice = async (report) => {
  try {
    const response = await fetch(
      /https:\/\/(.+?\.?\/)/.exec(whmcsApi.value)[0] +
        `modules/addons/nocloud/api/index.php?run=download_invoice&account=${report.account}&invoiceid=${report.transactionUuid}`
    );
    const data = await response.json();
    window.open(data, "_blanc");
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Error while download invoice",
    });
  }
};

const openInvoice = (report) => {
  router.push({
    name: "Transaction edit",
    params: { uuid: report.meta.transaction },
  });
};

watch(rates, () => {
  reports.value = reports.value.map((r) => ({
    ...r,
    totalDefault: convertFrom(r.total, r.currency),
  }));
});

watch(filters, fetchReportsDebounced, { deep: true });
watch(options, fetchReportsDebounced);

watch(reports, async () => {
  reports.value.forEach(async ({ account: uuid }) => {
    isAccountsLoading.value = true;
    try {
      if (!accounts.value[uuid]) {
        accounts.value[uuid] = api.accounts.get(uuid);
        accounts.value[uuid] = await accounts.value[uuid];
      }
    } catch (err) {
      accounts.value[uuid] = undefined;
    } finally {
      isAccountsLoading.value = Object.values(accounts.value).some(
        (acc) => acc instanceof Promise
      );
    }
  });

  isInstancesLoading.value = true;
  try {
    await Promise.all(
      reports.value.map(({ instance: uuid }) =>
        store.dispatch("instances/fetchToCached", uuid)
      )
    );
  } finally {
    isInstancesLoading.value = false;
  }
});
</script>
