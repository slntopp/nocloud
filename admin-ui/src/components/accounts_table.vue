<template>
  <nocloud-table
      table-name="accounts"
      :headers="headers"
      :items="filteredAccounts"
      :value="selected"
      :loading="loading"
      :single-select="singleSelect"
      :footer-error="fetchError"
      @input="handleSelect"
  >
    <template v-slot:[`item.title`]="{ item }">
      <div class="d-flex justify-space-between">
        <router-link
            :to="{ name: 'Account', params: { accountId: item.uuid } }"
        >
          {{ item.title }}
        </router-link>
        <div>
          <v-icon
              @click="
              $router.push({
                name: 'Account',
                params: { accountId: item.uuid },
                query: { tab: 2 },
              })
            "
              class="ml-5"
          >mdi-calendar-multiple</v-icon
          >
          <login-in-account-icon
              class="ml-5"
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
import nocloudTable from "@/components/table.vue";
import Balance from "./balance.vue";
import LoginInAccountIcon from "@/components/ui/loginInAccountIcon.vue";
import {
  compareSearchValue,
  filterByKeysAndParam,
  formatSecondsToDate,
  getDeepObjectValue,
} from "@/functions";
import api from "@/api";
import { toRefs, ref, computed, onMounted, watch } from "vue";
import { useStore } from "@/store";
import { useRouter } from "vue-router/composables";

const props = defineProps({
  value: {
    type: Array,
    default: () => [],
  },
  singleSelect: {
    type: Boolean,
    default: false,
  },
  notFiltered: { type: Boolean, default: false },
  namespace: {
    type: String,
  },
});
const { value, singleSelect, notFiltered } = toRefs(props);

const emit = defineEmits(["input"]);

const store = useStore();
const router = useRouter();

const selected = ref([]);
const loading = ref(false);
const fetchError = ref("");
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
  { text: "Invoice based", value: "data.regular_payment" },
  { text: "Group(NameSpace)", value: "namespace" },
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
  store.dispatch("namespaces/fetch");
  store
      .dispatch("accounts/fetch")
      .then(() => {
        fetchError.value = "";
      })
      .finally(() => {
        loading.value = false;
      })
      .catch((err) => {
        console.error(err.toJSON());
        fetchError.value = "Can't reach the server";
        if (err.response && err.response.data.message) {
          fetchError.value += `: [ERROR]: ${err.response.data.message}`;
        } else {
          fetchError.value += `: [ERROR]: ${err.toJSON().message}`;
        }
      });
});

const searchParam = computed(() => store.getters["appSearch/param"]);
const filter = computed(() => store.getters["appSearch/filter"]);
const accounts = computed(() => {
  return store.getters["accounts/all"].map((a) => ({
    ...a,
    access: {
      ...a.access,
      color: colorChip(a.access.level),
    },
    balance: a.balance || 0,
    currency: a.currency || defaultCurrency.value,
    namespace: getNamespaceName(a.uuid),
    data: {
      ...a.data,
      regular_payment:
          a.data?.regular_payment === undefined ||
          a.data?.regular_payment === true,
    },
  }));
});
const filteredAccounts = computed(() => {
  if (notFiltered.value) {
    return accounts.value;
  }

  const filtred = accounts.value.filter((a) => {
    return Object.keys(filter.value).every((key) => {
      let data;
      if (key === "namespace") {
        data = getNamespaceName(a.uuid);
      } else {
        data = getDeepObjectValue(a, key);
      }

      return compareSearchValue(
          data,
          filter.value[key],
          searchFields.value.find((f) => f.key === key)
      );
    });
  });

  if (searchParam.value) {
    return filterByKeysAndParam(
        filtred,
        ["title", "uuid", "data.email"],
        searchParam.value
    );
  }
  return filtred;
});
const namespaces = computed(() => store.getters["namespaces/all"]);
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
    items: [...new Set(accounts.value.map((a) => a.status))],
  },
  { title: "Balance", key: "balance", type: "number-range" },
  { title: "Email", key: "data.email", type: "input" },
  { title: "Created date", key: "data.date_create", type: "date" },
  { title: "Country", key: "data.country", type: "input" },
  { title: "Address", key: "data.address", type: "input" },
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
    items: Object.keys(levelColorMap.value),
  },
  {
    title: "Invoice based",
    key: "data.regular_payment",
    type: "logic-select",
  },
  {
    title: "Group(NameSpace)",
    key: "namespace",
    type: "select",
    items: [...new Set(accounts.value.map((a) => getNamespaceName(a.uuid)))],
  },
]);

const handleSelect = (item) => {
  emit("input", item);
};
const getNamespaceName = (uuid) => {
  return (
      namespaces.value.find(({ access }) => access.namespace === uuid)?.title ??
      ""
  );
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
  store.commit("appSearch/setFields", searchFields.value);
});

watch(value, () => {
  selected.value = value.value;
});
</script>

<script>
import search from "@/mixins/search";

export default {
  name: "accounts-table",
  mixins: [search({ name: "accounts-table" })],
};
</script>

<style></style>
