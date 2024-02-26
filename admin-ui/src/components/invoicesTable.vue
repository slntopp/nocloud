<template>
  <nocloud-table
    :table-name="tableName"
    class="mt-4"
    sort-by="created"
    sort-desc
    :show-select="false"
    :items="invoices"
    :headers="headers"
    :loading="isLoading"
    :server-items-length="count"
    :server-side-page="page"
    @update:options="setOptions"
  >
    <template v-slot:[`item.account`]="{ value }">
      <router-link :to="{ name: 'Account', params: { accountId: value } }">
        {{ account(value) }}
      </router-link>
    </template>

    <template v-slot:[`item.total`]="{ item }">
      <balance abs :currency="item.currency" :value="-item.total" />
    </template>
    <template v-slot:[`item.proc`]="{ item }">
      {{ formatSecondsToDate(item.proc, true) }}
    </template>
    <template v-slot:[`item.exec`]="{ item }">
      {{ formatSecondsToDate(item.exec, true) }}
    </template>
    <template v-slot:[`item.created`]="{ item }">
      {{ formatSecondsToDate(item.created, true) }}
    </template>
    <template v-slot:[`item.status`]="{ item }">
      <v-chip>{{ item.status }}</v-chip>
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
import balance from "@/components/balance.vue";
import { debounce, formatSecondsToDate } from "../functions";
import { ref, computed, watch, toRefs } from "vue";
import { useStore } from "@/store";

const props = defineProps({
  tableName: { type: String, default: "invoices-table" },
});
const { tableName } = toRefs(props);

const count = ref(10);
const page = ref(1);
const isFetchLoading = ref(true);
const isCountLoading = ref(true);
const options = ref({});
const fetchError = ref("");

const store = useStore();

const headers = ref([
  { text: "UUID ", value: "uuid" },
  { text: "Account ", value: "account" },
  { text: "Amount ", value: "total" },
  { text: "Payment date ", value: "proc" },
  { text: "Executed date ", value: "exec" },
  { text: "Created date ", value: "created" },
  { text: "Status ", value: "status" },
  { text: "Actions ", value: "actions" },
]);

const invoices = computed(() => store.getters["invoices/all"]);
const isLoading = computed(() => isFetchLoading.value || isCountLoading.value);

const filter = computed(() => store.getters["appSearch/filter"]);

const requestOptions = computed(() => ({
  filters: filter.value,
  page: page.value,
  limit: options.value.itemsPerPage,
  field: options.value?.[0],
  sort:
    options.value.sortBy?.[0] && options.value.sortDesc?.[0] ? "DESC" : "ASC",
}));

const account = (uuid) => {
  return accounts.value.find((acc) => acc.uuid === uuid)?.title;
};

const accounts = computed(() => store.getters["accounts/all"]);

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
      requestOptions.value
    );
    count.value = +total;
  } finally {
    isCountLoading.value = false;
  }
};

const fetchInvoices = async () => {
  init();
  isFetchLoading.value = true;
  fetchError.value = "";
  try {
    await store.dispatch("invoices/fetch", requestOptions.value);
  } catch (e) {
    fetchError.value = e.message;
  } finally {
    isFetchLoading.value = false;
  }
};

const fetchInvoicesDebounce = debounce(fetchInvoices, 100);

watch(filter, fetchInvoicesDebounce, { deep: true });
watch(options, fetchInvoicesDebounce);
</script>
