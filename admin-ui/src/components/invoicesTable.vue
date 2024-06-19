<template>
  <nocloud-table
    :table-name="tableName"
    class="mt-4"
    :value="value"
    @input="emit('input', $event)"
    sort-by="created"
    sort-desc
    :items="invoices"
    :headers="headers"
    :loading="isLoading"
    :server-items-length="count"
    :server-side-page="page"
    @update:options="setOptions"
  >
    <template v-slot:[`item.account`]="{ value }">
      <router-link
        v-if="!isAccountsLoading"
        :to="{ name: 'Account', params: { accountId: value } }"
      >
        {{ account(value) }}
      </router-link>
      <v-skeleton-loader type="text" v-else />
    </template>

    <template v-slot:[`item.total`]="{ item }">
      <v-chip :color="getTotalColor(item)" abs>
        {{ `${-item.total} ${item.currency?.title || defaultCurrency.title}` }}
      </v-chip>
    </template>

    <template v-slot:[`item.processed`]="{ item }">
      {{ formatSecondsToDate(item.processed, true) }}
    </template>

    <template v-slot:[`item.deadline`]="{ item }">
      {{ formatSecondsToDate(item.deadline, true) }}
    </template>

    <template v-slot:[`item.payment`]="{ item }">
      {{ formatSecondsToDate(item.payment, true) }}
    </template>

    <template v-slot:[`item.created`]="{ item }">
      {{ formatSecondsToDate(item.created, true) }}
    </template>

    <template v-slot:[`item.returned`]="{ item }">
      {{ formatSecondsToDate(item.returned, true) }}
    </template>

    <template v-slot:[`item.status`]="{ item }">
      <v-chip :color="getInvoiceStatusColor(item.status)">{{
        item.status
      }}</v-chip>
    </template>

    <template v-slot:[`item.actions`]="{ item }">
      <v-btn icon :to="{ name: 'Invoice page', params: { uuid: item.uuid } }">
        <v-icon>mdi-login</v-icon>
      </v-btn>
    </template>
  </nocloud-table>
</template>

<script setup>
import nocloudTable from "@/components/table.vue";
import { debounce, formatSecondsToDate } from "../functions";
import { ref, computed, watch, toRefs, onMounted } from "vue";
import { useStore } from "@/store";
import api from "@/api";
import { BillingStatus } from "nocloud-proto/proto/es/billing/billing_pb";

const props = defineProps({
  tableName: { type: String, default: "invoices-table" },
  value: {},
});
const { tableName, value } = toRefs(props);

const emit = defineEmits(["input"]);

const count = ref(10);
const page = ref(1);
const isFetchLoading = ref(true);
const isCountLoading = ref(true);
const options = ref({});
const fetchError = ref("");

const isAccountsLoading = ref(false);
const accounts = ref({});

const store = useStore();

const headers = ref([
  { text: "UUID ", value: "uuid" },
  { text: "Account ", value: "account" },
  { text: "Amount ", value: "total" },
  { text: "Type ", value: "type" },
  { text: "Created date ", value: "created" },
  { text: "Deadline date", value: "deadline" },
  { text: "Payment date", value: "payment" },
  { text: "Processed date", value: "processed" },
  { text: "Returned date", value: "returned" },
  { text: "Status ", value: "status" },
  { text: "Actions ", value: "actions" },
]);

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    event: () => {
      accounts.value = [];
      fetchInvoices();
    },
  });
});

const invoices = computed(() => store.getters["invoices/all"]);
const isLoading = computed(() => isFetchLoading.value || isCountLoading.value);

const defaultCurrency = computed(() => store.getters["currencies/default"]);

const filter = computed(() => store.getters["appSearch/filter"]);

const listOptions = computed(() => ({
  filters: filter.value,
  page: page.value,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy?.[0],
  sort:
    options.value.sortBy?.[0] && options.value.sortDesc?.[0] ? "DESC" : "ASC",
}));
const countOptions = computed(() => ({
  filters: filter.value,
}));

const account = (uuid) => {
  return accounts.value[uuid]?.title;
};

const setOptions = (newOptions) => {
  page.value = newOptions.page;
  if (JSON.stringify(newOptions) !== JSON.stringify(options.value)) {
    options.value = newOptions;
  }
};

const getInvoiceStatusColor = (status) => {
  switch (BillingStatus[status]) {
    case BillingStatus.CANCELED: {
      return "warning";
    }
    case BillingStatus.RETURNED: {
      return "blue";
    }
    case BillingStatus.DRAFT: {
      return "brown darked";
    }
    case BillingStatus.PAID: {
      return "green";
    }
    case BillingStatus.UNPAID: {
      return "gray";
    }
    case BillingStatus.BILLING_STATUS_UNKNOWN:
    case BillingStatus.TERMINATED:
    default: {
      return "red";
    }
  }
};

const getTotalColor = (item) => {
  if (
    BillingStatus[item.status] === BillingStatus.UNPAID &&
    item.deadline < Date.now() / 1000
  ) {
    return "red";
  }

  return "gray";
};

const init = async () => {
  isCountLoading.value = true;
  try {
    const { total } = await store.dispatch(
      "invoices/count",
      countOptions.value
    );
    count.value = Number(total);
  } finally {
    isCountLoading.value = false;
  }
};

const fetchInvoices = async () => {
  init();
  isFetchLoading.value = true;
  fetchError.value = "";
  try {
    await store.dispatch("invoices/fetch", listOptions.value);
  } catch (e) {
    fetchError.value = e.message;
  } finally {
    isFetchLoading.value = false;
  }
};

const fetchInvoicesDebounce = debounce(fetchInvoices, 100);

watch(filter, fetchInvoicesDebounce, { deep: true });
watch(options, fetchInvoicesDebounce);
watch(value, (newValue) => {
  if (newValue?.length === 0) {
    fetchInvoicesDebounce();
  }
});

watch(invoices, () => {
  invoices.value.forEach(async ({ account: uuid }) => {
    isAccountsLoading.value = true;
    try {
      if (!accounts.value[uuid]) {
        accounts.value[uuid] = api.accounts.get(uuid);
        accounts.value[uuid] = await accounts.value[uuid];
      }
    } finally {
      isAccountsLoading.value = Object.values(accounts.value).some(
        (acc) => acc instanceof Promise
      );
    }
  });
});
</script>
