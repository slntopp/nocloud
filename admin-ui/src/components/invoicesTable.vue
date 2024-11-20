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
    :show-select="!hideSelect"
  >
    <template v-slot:[`uuid-actions`]="{ item }">
      <v-btn @click="downloadInvoice(item)" icon>
        <v-icon>mdi-download</v-icon>
      </v-btn>
    </template>

    <template v-slot:[`item.account`]="{ value }">
      <router-link
        v-if="!isAccountsLoading"
        :to="{ name: 'Account', params: { accountId: value } }"
      >
        {{ account(value) }}
      </router-link>
      <v-skeleton-loader type="text" v-else />
    </template>

    <template v-slot:[`item.number`]="{ item }">
      <router-link :to="{ name: 'Invoice page', params: { uuid: item.uuid } }">
        {{ item.number }}
      </router-link>
    </template>

    <template v-slot:[`item.total`]="{ item }">
      <v-chip :color="getTotalColor(item)" abs>
        {{ `${item.total} ${item.currency?.title || defaultCurrency.title}` }}
      </v-chip>
    </template>

    <template v-slot:[`item.processed`]="{ item }">
      {{ formatSecondsToDate(item.processed, true) }}
    </template>

    <template v-slot:[`item.meta.whmcs_invoice_id`]="{ item }">
      <a
        style="text-decoration: underline"
        @click="() => downloadInvoice(item)"
        class="router-link-exact-active router-link-active"
      >
        {{ item?.meta?.whmcs_invoice_id }}
      </a>
    </template>

    <template v-slot:[`item.type`]="{ value }">
      {{ value?.replaceAll("_", " ") }}
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
  </nocloud-table>
</template>

<script setup>
import nocloudTable from "@/components/table.vue";
import { debounce, formatSecondsToDate } from "../functions";
import { ref, computed, watch, toRefs, onMounted } from "vue";
import { useStore } from "@/store";
import api from "@/api";
import useInvoices from "../hooks/useInvoices";
import AccountsAutocomplete from "@/components/ui/accountsAutocomplete.vue";
import {
  ActionType,
  BillingStatus,
} from "nocloud-proto/proto/es/billing/billing_pb";
import useSearch from "@/hooks/useSearch";

const props = defineProps({
  tableName: { type: String, default: "invoices-table" },
  value: {},
  customFilter: {},
  noSearch: { type: Boolean, default: false },
  hideSelect: { type: Boolean, default: false },
  refetch: { type: Boolean, default: false },
});
const { tableName, value, refetch, noSearch, customFilter } = toRefs(props);

const emit = defineEmits(["input"]);

const store = useStore();
const { getInvoiceStatusColor, getTotalColor } = useInvoices();
useSearch({
  name: props.tableName,
  noSearch: props.noSearch,
});

const count = ref(10);
const page = ref(1);
const isFetchLoading = ref(true);
const isCountLoading = ref(true);
const options = ref({});
const fetchError = ref("");

const isAccountsLoading = ref(false);
const accounts = ref({});

const headers = ref([
  { text: "UUID", value: "uuid" },
  { text: "Number", value: "number" },
  { text: "Account ", value: "account" },
  { text: "Total ", value: "total" },
  { text: "Type ", value: "type" },
  { text: "External ID", value: "meta.whmcs_invoice_id" },
  { text: "Created date ", value: "created" },
  { text: "Due date", value: "deadline" },
  { text: "Payment date", value: "payment" },
  { text: "Processed date", value: "processed" },
  { text: "Returned date", value: "returned" },
  { text: "Status ", value: "status" },
]);

onMounted(() => {
  if (!props.noSearch) {
    store.commit("appSearch/setFields", searchFields.value);
  }

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
const searchParam = computed(() => store.getters["appSearch/param"]);
const searchFields = computed(() => [
  {
    title: "Number",
    key: "number",
    type: "input",
  },
  {
    title: "Status",
    key: "status",
    type: "select",
    items: Object.keys(BillingStatus)
      .filter((value) => !Number.isInteger(+value))
      .map((key) => ({
        text: key,
        value: BillingStatus[key],
      })),
  },
  {
    title: "Type",
    key: "type",
    type: "select",
    items: Object.keys(ActionType)
      .filter((value) => !Number.isInteger(+value))
      .map((key) => ({
        text: key,
        value: ActionType[key],
      })),
  },
  {
    key: "account",
    custom: true,
    multiple: true,
    fetchValue: true,
    title: "Account",
    component: AccountsAutocomplete,
  },
  { title: "Total", key: "total", type: "number-range" },
  { title: "External ID", key: "whmcs_invoice_id", type: "input" },
  { title: "Due date", key: "deadline", type: "date" },
  { title: "Created date", key: "created", type: "date" },
  { title: "Payment date", key: "payment", type: "date" },
  { title: "Processed date", key: "processed", type: "date" },
  { title: "Returned date", key: "returned", type: "date" },
]);

const invoicesFilters = computed(() => {
  const filters = {};

  if (noSearch.value) {
    for (const key of Object.keys(customFilter.value)) {
      filters[key] = customFilter.value[key];
    }
  } else {
    const datekeys = [
      "created",
      "processed",
      "returned",
      "payment",
      "deadline",
    ];

    for (const key of Object.keys(filter.value)) {
      const value = filter.value[key];

      if (
        !value ||
        (Array.isArray(value) && !value.length) ||
        (typeof value === "object" && !Object.keys(value).length)
      ) {
        continue;
      }

      if (value?.to || value?.from) {
        const total = {};
        if (value?.to) {
          total.to = +value?.to;
        }
        if (value?.from) {
          total.from = +value?.from;
        }

        filters[key] = total;
        continue;
      }

      if (datekeys.includes(key)) {
        let dates = [];

        if (value[0]) {
          dates.push(new Date(value[0]).getTime() / 1000);
        }
        if (value[1]) {
          dates.push(new Date(value[1]).getTime() / 1000);
        }

        dates = dates.sort();

        const result = { from: dates[0] };
        if (dates[1]) {
          result.to = dates[1];
        }
        filters[key] = result;
        continue;
      }

      filters[key] = filter.value[key];
    }
  }

  if (searchParam.value) {
    filters.search_param = searchParam.value;
  }

  return filters;
});

const listOptions = computed(() => ({
  filters: invoicesFilters.value,
  page: page.value,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy?.[0],
  sort:
    options.value.sortBy?.[0] && options.value.sortDesc?.[0] ? "DESC" : "ASC",
}));
const countOptions = computed(() => ({
  filters: invoicesFilters.value,
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

const fetchInvoicesDebounce = debounce(fetchInvoices, 300);

const downloadInvoice = async (invoice) => {
  try {
    const { paymentLink } = await store.getters["invoices/invoicesClient"].pay({
      invoiceId: invoice.uuid,
    });
    if (!paymentLink) {
      throw new Error("No link");
    }
    window.open(paymentLink, "_blanc");
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: "Document not found",
    });
  }
};

watch(
  invoicesFilters,
  () => {
    page.value = 1;
    fetchInvoicesDebounce();
  },
  { deep: true }
);
watch(options, fetchInvoicesDebounce);
watch(refetch, fetchInvoicesDebounce);

watch(invoices, () => {
  invoices.value.forEach(async ({ account: uuid }) => {
    isAccountsLoading.value = true;
    try {
      if (!accounts.value[uuid]) {
        accounts.value[uuid] = api.accounts.get(uuid);
        accounts.value[uuid] = await accounts.value[uuid];
      }
    } catch (e) {
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
  name: "invoices-table",
  mixins: [],
};
</script>
