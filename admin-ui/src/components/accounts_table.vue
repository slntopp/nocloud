<template>
  <nocloud-table
    table-name="accounts"
    :headers="headers"
    :items="accounts"
    :value="selected"
    :loading="loading"
    :single-select="singleSelect"
    :footer-error="fetchError"
    @input="handleSelect"
    :server-items-length="total"
    :server-side-page="options.page"
    @update:options="setOptions"
  >
    <template v-slot:[`item.title`]="{ item }">
      <div class="d-flex justify-space-between">
        <router-link
          :to="{ name: 'Account', params: { accountId: item.uuid } }"
        >
          {{ item.title }}
        </router-link>
        <div>
          <whmcs-btn :account="item" />
          <login-in-account-icon
            v-if="['ROOT', 'ADMIN'].includes(item.access.level)"
            :uuid="item.uuid"
          />
        </div>
      </div>
    </template>
    <template v-slot:[`item.balance`]="{ item }">
      <balance
        :hide-currency="true"
        :currency="item.currency"
        @click="goToBalance(item.uuid)"
        :value="item.balance"
      />
    </template>

    <template v-slot:[`item.data.date_create`]="{ value }">
      {{ formatSecondsToDate(value) }}
    </template>

    <template v-slot:[`item.address`]="{ item }">
      <v-tooltip bottom>
        <template v-slot:activator="{ on, attrs }">
          <span v-bind="attrs" v-on="on">
            {{ item.data?.city || item.data?.address }}
          </span>
        </template>
        <span>{{ item.data?.address }}</span>
      </v-tooltip>
    </template>
    <template v-slot:[`item.access.level`]="{ item }">
      <v-chip :color="item.access.color">
        {{ item.access.level }}
      </v-chip>
    </template>
    <template v-slot:[`item.data.regular_payment`]="{ value, item }">
      <v-switch
        :disabled="
          !!changeRegularPaymentUuid && changeRegularPaymentUuid !== item.uuid
        "
        :loading="
          !!changeRegularPaymentUuid && changeRegularPaymentUuid === item.uuid
        "
        @change="changeRegularPayment(item, $event)"
        :input-value="value"
      >
      </v-switch>
    </template>
  </nocloud-table>
</template>

<script setup>
import Balance from "./balance.vue";
import LoginInAccountIcon from "@/components/ui/loginInAccountIcon.vue";
import { debounce, formatSecondsToDate } from "@/functions";
import api from "@/api";
import { toRefs, ref, computed, onMounted, watch } from "vue";
import { useStore } from "@/store";
import { useRouter } from "vue-router/composables";
import NocloudTable from "@/components/table.vue";
import whmcsBtn from "@/components/ui/whmcsBtn.vue";

const props = defineProps({
  value: {
    type: Array,
    default: () => [],
  },
  singleSelect: {
    type: Boolean,
    default: false,
  },
  noSearch: { type: Boolean, default: false },
  customSearchParam: { type: String, default: "" },
});
const { value, singleSelect, customSearchParam, noSearch } = toRefs(props);

const emit = defineEmits(["input"]);

const store = useStore();
const router = useRouter();

const selected = ref([]);
const loading = ref(false);
const fetchError = ref("");
const options = ref({});
const changeRegularPaymentUuid = ref("");
const headers = ref([
  { text: "Title", value: "title" },
  { text: "UUID", value: "uuid" },
  { text: "Status", value: "status" },
  { text: "Balance", value: "balance" },
  { text: "Email", value: "data.email" },
  { text: "Created date", value: "data.date_create" },
  { text: "Country", value: "data.country" },
  { text: "Address", value: "address" },
  { text: "Client currency", value: "currency" },
  { text: "Access level", value: "access.level" },
  { text: "WHMCS ID", value: "data.whmcs_id" },
  { text: "Invoice based", value: "data.regular_payment" },
]);
const levelColorMap = ref({
  ROOT: "info",
  ADMIN: "success",
  MGMT: "warning",
  READ: "gray",
  NONE: "error",
});

onMounted(() => {
  loading.value = true;
  store.commit("reloadBtn/setCallback", {
    event: fetchAccounts,
  });
});

const searchParam = computed(() => store.getters["appSearch/param"]);
const filter = computed(() => {
  const filter = store.getters["appSearch/filter"];
  const total = {};
  if (filter.total?.to) {
    total.to = +filter.total.to;
  }
  if (filter.total?.from) {
    total.from = +filter.total.from;
  }

  const dates = {};
  const dateKeys = ["data.date_create"];
  dateKeys.forEach((key) => {
    if (!filter[key]) {
      return;
    }
    dates[key] = {};

    if (filter[key][0]) {
      dates[key].from = new Date(filter[key][0]).getTime() / 1000;
    }
    if (filter[key][1]) {
      dates[key].to = new Date(filter[key][1]).getTime() / 1000;
    }
  });

  return {
    ...filter,
    ...dates,
    title: undefined,
    search_param: filter.title || searchParam.value || undefined,
    balance: Object.keys(total).length ? total : undefined,
  };
});
const accounts = computed(() => {
  return store.getters["accounts/all"].map((a) => ({
    ...a,
    access: {
      ...a.access,
      color: colorChip(a.access.level),
    },
    balance: a.balance || 0,
    currency: a.currency || defaultCurrency.value,
    data: {
      ...a.data,
      regular_payment:
        a.data?.regular_payment === undefined ||
        a.data?.regular_payment === true,
    },
  }));
});
const total = computed(() => store.getters["accounts/total"]);

const requestOptions = computed(() => ({
  filters: !noSearch.value
    ? {
        ...filter.value,
        "data.whmcs_id": +filter.value["data.whmcs_id"] || undefined,
        balance: filter.value?.balance && {
          from: filter.value?.balance.from && +filter.value?.balance.from,
          to: filter.value?.balance.to && +filter.value?.balance.to,
        },
        search_param:
          searchParam.value || filter.value.search_param || undefined,
      }
    : { search_param: customSearchParam.value || undefined },
  page: options.value.page,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy[0],
  sort: options.value.sortBy[0] && options.value.sortDesc[0] ? "DESC" : "ASC",
}));

const defaultCurrency = computed(() => store.getters["currencies/default"]);
const searchFields = computed(() => [
  {
    title: "Title",
    key: "title",
    type: "input",
  },
  {
    title: "Status",
    key: "status",
    type: "select",
    item: { value: "uuid", title: "title" },
    items: [
      { title: "ACTIVE", uuid: 0 },
      { title: "LOCK", uuid: 1 },
      { title: "PERMANENT_LOCK", uuid: 2 },
    ],
  },
  { title: "Balance", key: "balance", type: "number-range" },
  { title: "Email", key: "data.email", type: "input" },
  { title: "Created date", key: "data.date_create", type: "date" },
  { title: "Country", key: "data.country", type: "input" },
  { title: "Address", key: "data.address", type: "input" },
  { title: "WHMCS ID", key: "data.whmcs_id", type: "input" },
  {
    title: "Client currency",
    key: "currency",
    type: "select",
    items: store.getters["currencies/all"].filter((c) => c !== "NCU"),
  },
  {
    title: "Access level",
    key: "access.level",
    type: "select",
    item: { value: "id", title: "title" },
    items: [
      { id: 0, title: "NONE" },
      { id: 1, title: "READ" },
      { id: 2, title: "MGMT" },
      { id: 3, title: "ADMIN" },
    ],
  },
  {
    title: "Invoice based",
    key: "data.regular_payment",
    type: "logic-select",
  },
]);

const setOptions = (newOptions) => {
  if (JSON.stringify(newOptions) !== JSON.stringify(options.value)) {
    options.value = newOptions;
  }
};

const fetchAccounts = async () => {
  loading.value = true;
  try {
    await store.dispatch("accounts/fetch", requestOptions.value);
  } catch (err) {
    fetchError.value = "Can't reach the server";
    if (err.response && err.response.data.message) {
      fetchError.value += `: [ERROR]: ${err.response.data.message}`;
    } else {
      fetchError.value += `: [ERROR]: ${err.toJSON().message}`;
    }
  } finally {
    loading.value = false;
  }
};

const fetchAccountsDebounce = debounce(fetchAccounts);

const handleSelect = (item) => {
  emit("input", item);
};
const colorChip = (level) => {
  return levelColorMap.value[level];
};
const goToBalance = (uuid) => {
  router.push({ name: "Transactions", query: { account: uuid } });
};
const changeRegularPayment = async (item, value) => {
  changeRegularPaymentUuid.value = item.uuid;
  try {
    if (!item.data) {
      item.data = {};
    }
    item.data.regular_payment = value;
    await api.accounts.update(item.uuid, item);
    store.commit("accounts/pushAccount", item);
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Error while change invoice based",
    });
  } finally {
    changeRegularPaymentUuid.value = "";
  }
};

watch(accounts, () => {
  fetchError.value = "";
});

watch(searchFields, () => {
  if (!noSearch.value) {
    store.commit("appSearch/setFields", searchFields.value);
  }
});

watch(value, () => {
  selected.value = value.value;
});

watch(filter, fetchAccountsDebounce, { deep: true });
watch(options, fetchAccountsDebounce);
watch(customSearchParam, fetchAccountsDebounce);
</script>

<script>
import search from "@/mixins/search";

export default {
  name: "accounts-table",
  mixins: [search({ name: "accounts-table" })],
};
</script>

<style></style>
