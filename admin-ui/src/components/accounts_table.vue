<template>
  <nocloud-table
    :table-name="tableName"
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
          <div
            class="color-box"
            :style="{ backgroundColor: getColorByGroup(item) }"
          ></div>

          {{ getShortName(item.title) }}
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

    <template v-slot:[`item.data.email`]="{ item }">
      {{ getShortName(item.data.email) }}
    </template>

    <template v-slot:[`item.is_phone_verified`]="{ item }">
      <div class="d-flex justify-center">
        <v-icon :color="item.isPhoneVerified ? 'green' : 'red'">{{
          item.isPhoneVerified ? "mdi-check-circle" : "mdi-close-circle"
        }}</v-icon>
      </div>
    </template>

    <template v-slot:[`item.data.tax_rate`]="{ item }">
      {{ (item.data.tax_rate || 0) * 100 }}
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
  </nocloud-table>
</template>

<script setup>
import Balance from "./balance.vue";
import LoginInAccountIcon from "@/components/ui/loginInAccountIcon.vue";
import { debounce, formatSecondsToDate, getShortName } from "@/functions";
import { toRefs, ref, computed, onMounted, watch } from "vue";
import { useStore } from "@/store";
import { useRouter } from "vue-router/composables";
import NocloudTable from "@/components/table.vue";
import whmcsBtn from "@/components/ui/whmcsBtn.vue";
import useSearch from "@/hooks/useSearch";

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
  customFilter: { type: Object },
  tableName: { type: String, default: "accounts" },
});
const { value, singleSelect, customSearchParam, noSearch, customFilter } =
  toRefs(props);

const emit = defineEmits(["input"]);

const store = useStore();
const router = useRouter();
useSearch({
  name: props.tableName,
  noSearch: props.noSearch,
});

const selected = ref([]);
const loading = ref(false);
const fetchError = ref("");
const options = ref({});
const headers = ref([
  { text: "Title", value: "title" },
  { text: "UUID", value: "uuid" },
  { text: "Status", value: "status" },
  { text: "Balance", value: "balance" },
  { text: "Email", value: "data.email" },
  { text: "Phone verified", value: "is_phone_verified" },
  { text: "Created date", value: "data.date_create" },
  { text: "Country", value: "data.country" },
  { text: "Address", value: "address" },
  { text: "Client currency", value: "currency.code" },
  { text: "Access level", value: "access.level" },
  { text: "WHMCS ID", value: "data.whmcs_id" },
  { text: "Tax rate", value: "data.tax_rate" },
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

  store.dispatch("accountGroups/fetch");
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

  filter["status"] = filter["status"]?.map((val) => +val);

  return {
    ...filter,
    ...dates,
    title: undefined,
    search_param: filter.title || searchParam.value || undefined,
    balance: Object.keys(total).length ? total : undefined,
  };
});
const accounts = computed(() =>
  store.getters["accounts/all"].map((a) => {
    const currency = a.currency || defaultCurrency.value;
    return {
      ...a,
      access: {
        ...a.access,
        color: colorChip(a.access.level),
      },
      balance: a.balance,
      currency,
      data: {
        ...a.data,
      },
    };
  })
);
const total = computed(() => store.getters["accounts/total"]);

const accountGroups = computed(() => store.getters["accountGroups/all"]);

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
    : customFilter.value
    ? customFilter.value
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
      { title: "ACTIVE", uuid: "0" },
      { title: "LOCK", uuid: "1" },
      { title: "PERMANENT_LOCK", uuid: "2" },
    ],
  },
  { title: "Balance", key: "balance", type: "number-range" },
  { title: "Email", key: "data.email", type: "input" },
  { title: "Phone verified", key: "is_phone_verified", type: "logic-select" },
  { title: "Created date", key: "data.date_create", type: "date" },
  { title: "Company", key: "data.company", type: "input" },
  { title: "Country", key: "data.country", type: "input" },
  { title: "City", key: "data.city", type: "input" },
  { title: "Address", key: "data.address", type: "input" },
  { title: "Phone", key: "data.phone", type: "input" },
  { title: "WHMCS ID", key: "data.whmcs_id", type: "input" },
  {
    title: "Client currency",
    key: "currency",
    type: "select",
    items: store.getters["currencies/all"]
      .filter((c) => c.title !== "NCU")
      .map((c) => ({ text: c.code, value: c.id })),
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
]);

const setOptions = (newOptions) => {
  if (JSON.stringify(newOptions) !== JSON.stringify(options.value)) {
    options.value = newOptions;
  }
};

const fetchAccounts = async () => {
  loading.value = true;
  try {
    await store.dispatch("accounts/fetch", {
      ...requestOptions.value,
      silent: true,
    });
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
const getColorByGroup = (account) => {
  const group = accountGroups.value.find(
    (g) => g.uuid == account.accountGroup
  );
  return group?.color || "#CCCCCC";
};
const goToBalance = (uuid) => {
  router.push({ name: "Transactions", query: { account: uuid } });
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
export default {
  name: "accounts-table",
};
</script>

<style scoped lang="scss">
.color-box {
  width: 12px;
  height: 12px;
  border-radius: 2px;
  display: inline-block;
  margin-right: 6px;
}
</style>
