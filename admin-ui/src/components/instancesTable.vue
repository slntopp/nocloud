<template>
  <nocloud-table
    :table-name="tableName"
    class="mt-4"
    :value="value"
    @input="emit('input', $event)"
    sort-by="created"
    sort-desc
    :items="instances"
    :headers="headers"
    :loading="isLoading"
    :server-items-length="total"
    :server-side-page="page"
    @update:options="setOptions"
    editable
    :footer-error="fetchError"
    @update:edit-values="updateEditValues"
    :show-select="showSelect"
  >
    <template v-slot:[`item.account`]="{ value }">
      <router-link
        v-if="!isAccountsLoading"
        :to="{ name: 'Account', params: { accountId: value } }"
      >
        {{ getAccount(value)?.title }}
      </router-link>
      <v-skeleton-loader type="text" v-else />
    </template>

    <template v-slot:[`item.config.auto_renew`]="{ item }">
      <div class="d-flex justify-center align-center regular_payment">
        <v-switch
          dense
          hide-details
          readonly
          :input-value="item.config?.auto_renew"
        />
      </div>
    </template>

    <template v-slot:[`item.data.next_payment_date`]="{ item }">
      {{
        typeof getExpirationDate(item) === "number"
          ? formatSecondsToDate(getExpirationDate(item))
          : getExpirationDate(item)
      }}
    </template>

    <template v-slot:[`item.config.regular_payment`]="{ item }">
      <div class="d-flex justify-center align-center regular_payment">
        <v-switch
          dense
          hide-details
          :disabled="isChangeRegularPaymentLoading"
          :input-value="item.config?.regular_payment"
          @change="changeRegularPayment(item, $event)"
        />
      </div>
    </template>

    <template v-slot:[`item.state.meta.networking`]="{ item }">
      <template v-if="!item.state?.meta?.networking?.public">-</template>
      <instance-ip-menu v-else :item="item" ui="span" />
    </template>

    <template v-slot:[`item.config.location`]="{ value }">
      {{ value || "Unknown" }}
    </template>

    <template v-slot:[`item.period`]="{ item }">
      {{ getPeriod(item) }}
    </template>

    <template v-slot:[`item.created`]="{ item }">
      {{ formatSecondsToDate(+item.created || "Unknown") }}
    </template>

    <template v-slot:[`item.id`]="{ index }">
      {{ index + 1 }}
    </template>

    <template v-slot:[`item.title`]="{ item }">
      <div class="d-flex justify-space-between">
        <router-link
          :target="openInNewTab ? '_blank' : null"
          :to="{ name: 'Instance', params: { instanceId: item.uuid } }"
        >
          {{ getShortName(item.title, 45) }}
        </router-link>
        <login-in-account-icon
          :uuid="getAccount(item)?.uuid"
          :instanceId="item.uuid"
          :type="item.type"
        />
      </div>
    </template>

    <template v-slot:[`item.billingPlan.title`]="{ item, value }">
      <router-link
        :to="{ name: 'Plan', params: { planId: item.billingPlan.uuid } }"
      >
        {{ getShortName(value) }}
      </router-link>
    </template>

    <template v-slot:[`item.product`]="{ item }">
      <router-link
        :to="{
          name: 'Plan',
          params: {
            planId: item.billingPlan?.uuid,
          },
        }"
      >
        {{ getShortName(item.billingPlan.products[item.product]?.title) }}
      </router-link>
    </template>

    <template v-slot:[`item.email`]="{ item }">
      <span v-if="!isAccountsLoading">
        {{ getShortName(getAccount(item.account)?.data?.email ?? "-") }}
      </span>
      <v-skeleton-loader type="text" v-else />
    </template>

    <template v-slot:[`item.resources.cpu`]="{ value }">
      {{ value || 0 }}
      {{ value > 1 ? "cores" : "core" }}
    </template>

    <template v-slot:[`item.resources.ram`]="{ value }">
      {{ +(value / 1024).toFixed(2) || 0 }} GB
    </template>

    <template v-slot:[`item.resources.drive_size`]="{ value }">
      {{ +(value / 1024).toFixed(2) || 0 }} GB
    </template>

    <template v-slot:[`item.config.template_id`]="{ item }">
      {{ getOSName(item) }}
    </template>

    <template v-slot:[`item.state.state`]="{ item }">
      <instance-state small :template="item" />
    </template>

    <template v-slot:[`item.sp`]="{ value }">
      <router-link :to="{ name: 'ServicesProvider', params: { uuid: value } }">
        {{ getShortName(getServiceProvider(value)?.title) }}
      </router-link>
    </template>

    <template v-slot:[`item.estimate`]="{ item }">
      {{ item.estimate }} {{ defaultCurrency?.title }}
    </template>

    <template v-slot:[`item.accountPrice`]="{ item }">
      <span v-if="!isAccountsLoading">
        {{ convertTo(item.estimate, getAccount(item.account)?.currency) }}
        {{ getAccount(item.account)?.currency?.title }}
      </span>

      <v-skeleton-loader type="text" v-else />
    </template>
  </nocloud-table>
</template>

<script setup>
import nocloudTable from "@/components/table.vue";
import {
  debounce,
  formatDateToTimestamp,
  formatSecondsToDate,
  getBillingPeriod,
  getShortName,
  isInstancePayg,
} from "../functions";
import { ref, computed, watch, toRefs, onMounted } from "vue";
import { useStore } from "@/store";
import api from "@/api";
import instanceIpMenu from "./ui/instanceIpMenu.vue";
import LoginInAccountIcon from "@/components/ui/loginInAccountIcon.vue";
import InstanceState from "@/components/ui/instanceState.vue";
import useCurrency from "@/hooks/useCurrency";
import AccountsAutocomplete from "@/components/ui/accountsAutocomplete.vue";
import useSearch from "@/hooks/useSearch";
import { UpdateRequest } from "nocloud-proto/proto/es/instances/instances_pb";

const props = defineProps({
  tableName: { type: String, default: "instances-table" },
  value: { type: Array, required: false },
  refetch: { type: Boolean, default: false },
  headers: {},
  showSelect: { type: Boolean, default: true },
  openInNewTab: { type: Boolean, default: false },
  noSearch: { type: Boolean, default: false },
  customFilter: { type: Object, default: () => {} },
});
const { value, refetch, showSelect, openInNewTab, customFilter } =
  toRefs(props);

const emit = defineEmits(["input"]);

const store = useStore();
const { defaultCurrency, convertTo } = useCurrency();
useSearch({
  name: props.tableName,
  noSearch: props.noSearch,
  defaultLayout: {
    title: "Default",
    filter: {
      "state.state": [0, 1, 2, 3, 4, 6, 7, 8],
    },
  },
});

const page = ref(1);
const options = ref({});
const fetchError = ref("");

const isUniqueFetched = ref(false);
const uniqueProducts = ref([]);
const uniqueLocations = ref([]);
const uniquePlans = ref([]);
const uniquePeriods = ref([]);

const instancesTypes = ref([]);

const isAccountsLoading = ref(false);
const accounts = ref({});

const isChangeRegularPaymentLoading = ref(false);

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    event: () => {
      accounts.value = [];
      fetchInstances();
    },
  });

  const types = require.context(
    "@/components/modules/",
    true,
    /instanceCreate\.vue$/
  );
  types.keys().forEach((key) => {
    const matched = key.match(/\.\/([A-Za-z0-9-_,\s]*)\/instanceCreate\.vue/i);
    if (matched && matched.length > 1) {
      const type = matched[1];
      instancesTypes.value.push(type);
    }
  });
});

const headers = computed(() => {
  if (props.headers) return props.headers;
  const headers = [
    { text: "ID", value: "id" },
    { text: "Name", value: "title" },
    { text: "Account", value: "account" },
    {
      text: "Due date",
      value: "data.next_payment_date",
      editable: { type: "date" },
    },
    { text: "Status", value: "state.state" },
    { text: "Tariff", value: "product" },
    { text: "Service provider", value: "sp" },
    { text: "Location", value: "config.location" },
    { text: "Type", value: "type" },
    { text: "NCU price", value: "estimate" },
    { text: "Account price", value: "accountPrice" },
    { text: "Period", value: "period" },
    { text: "Email", value: "email" },
    { text: "Created date", value: "created", editable: { type: "date" } },
    { text: "UUID", value: "uuid" },
    { text: "Price model", value: "billingPlan.title" },
    { text: "IP", value: "state.meta.networking" },
    { text: "CPU", value: "resources.cpu" },
    { text: "RAM", value: "resources.ram" },
    { text: "Disk", value: "resources.drive_size" },
    { text: "OS", value: "config.template_id" },
    { text: "Domain", value: "resources.domain" },
    { text: "DCV", value: "resources.dcv" },
    { text: "Approver email", value: "resources.approver_email" },
    {
      text: "Invoice based",
      value: "config.regular_payment",
      editable: { type: "logic-select" },
    },
    { text: "Auto renew", value: "config.auto_renew" },
  ];
  return headers;
});

const instances = computed(() =>
  store.getters["instances/all"].map((i) => ({
    ...i.instance,
    ...i,
    instance: undefined,
  }))
);
const isLoading = computed(() => store.getters["instances/isLoading"]);
const total = computed(() => store.getters["instances/total"]);

const servicesProviders = computed(
  () => store.getters["servicesProviders/all"]
);
const filter = computed(() => store.getters["appSearch/filter"]);
const searchParam = computed(() => store.getters["appSearch/param"]);

const listOptions = computed(() => {
  const filters = {};
  const datekeys = ["created", "data.next_payment_date"];

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

  if (searchParam.value) {
    filters.search_param = searchParam.value;
  }

  //more priority
  for (const key in customFilter.value) {
    filters[key] = customFilter.value[key];
  }

  return {
    filters: filters,
    page: page.value,
    limit: options.value.itemsPerPage,
    field: options.value.sortBy?.[0],
    sort:
      options.value.sortBy?.[0] && options.value.sortDesc?.[0] ? "DESC" : "ASC",
  };
});

const searchFields = computed(() => [
  {
    key: "title",
    title: "Title",
    type: "input",
  },
  {
    key: "period",
    items: uniquePeriods.value.map((period) => ({
      text: getBillingPeriod(period),
      value: period,
    })),
    title: "Period",
    type: "select",
  },
  {
    key: "sp",
    items: servicesProviders.value.map((sp) => ({
      text: sp.title,
      value: sp.uuid,
    })),
    type: "select",
    title: "Service provider",
  },
  {
    key: "config.location",
    items: uniqueLocations.value,
    type: "select",
    title: "Location",
  },
  {
    key: "account",
    custom: true,
    multiple: true,
    title: "Account",
    component: AccountsAutocomplete,
  },
  {
    key: "product",
    items: uniqueProducts.value,
    type: "select",
    title: "Product",
  },
  {
    key: "state.state",
    items: [
      { text: "INIT", value: 0 },
      { text: "RUNNING", value: 3 },
      { text: "STOPPED", value: 2 },
      { text: "PENDING", value: 8 },
      { text: "OPERATION", value: 7 },
      { text: "SUSPENDED", value: 6 },
      { text: "DELETED", value: 5 },
      { text: "ERROR", value: 4 },
      { text: "UNKNOWN", value: 1 },
    ],
    type: "select",
    title: "State",
  },
  {
    key: "type",
    title: "Type",
    type: "select",
    items: instancesTypes.value,
  },
  {
    key: "billing_plan",
    type: "select",
    title: "Billing plan",
    items: uniquePlans.value.map((plan) => ({
      text: plan.title,
      value: plan.uuid,
    })),
  },
  { title: "Due date", key: "data.next_payment_date", type: "date" },
  { title: "NCU price", key: "estimate", type: "number-range" },
  { title: "Email", key: "email", type: "input" },
  { title: "Date", key: "created", type: "date" },
  { title: "IP", key: "state.meta.networking", type: "input" },
  {
    key: "resources.cpu",
    type: "number-range",
    title: "CPU",
  },
  { title: "RAM", key: "resources.ram", type: "number-range" },
  { title: "Disk", key: "resources.drive_size", type: "number-range" },
  { title: "OS", key: "config.template_id", type: "input" },
  { title: "Domain", key: "resources.domain", type: "input" },
  { title: "DCV", key: "resources.dcv", type: "input" },
  {
    title: "Approver email",
    key: "resources.approver_email",
    type: "input",
  },
]);

const getAccount = (uuid) => {
  return accounts.value[uuid];
};

const setOptions = (newOptions) => {
  const replaceSortKeys = { accountPrice: "estimate" };

  for (const key in replaceSortKeys) {
    if (newOptions.sortBy.includes(key)) {
      newOptions.sortBy = [replaceSortKeys[key]];
    }
  }

  page.value = newOptions.page;
  if (JSON.stringify(newOptions) !== JSON.stringify(options.value)) {
    options.value = newOptions;
  }
};

const fetchInstances = async () => {
  fetchError.value = "";
  try {
    if (!isUniqueFetched.value) {
      fetchUnique();
    }

    console.log(listOptions.value);
    await store.dispatch("instances/fetch", listOptions.value);
  } catch (e) {
    fetchError.value = e.message;
  }
};

const fetchUnique = async () => {
  try {
    const { unique } = await api.get("/instances/unique");
    uniqueLocations.value = unique.locations;
    uniqueProducts.value = unique.products;
    uniquePlans.value = unique.billing_plans;
    uniquePeriods.value = unique.periods;

    isUniqueFetched.value = true;
  } catch {
    isUniqueFetched.value = false;
  }
};

const fetchInstancesDebounce = debounce(fetchInstances, 100);

const getPeriod = (instance) => {
  if (isInstancePayg(instance)) {
    return "PayG";
  }

  const period = getBillingPeriod(Number(instance.period));

  return period || "Unknown";
};

const getExpirationDate = (instance) => {
  if (isInstancePayg(instance)) return "PayG";
  if (getPeriod(instance) === "One time") return "One time";
  return instance.data.next_payment_date || "Unknown";
};

const getOSName = (instance) => {
  const id = instance.config?.template_id;

  if (!id) return;
  return servicesProviders.value.find(({ uuid }) => uuid === instance.sp)
    ?.publicData.templates[id]?.name;
};

const getServiceProvider = (uuid) => {
  return servicesProviders.value.find((sp) => sp.uuid === uuid);
};

const changeRegularPayment = async (instance, value) => {
  isChangeRegularPaymentLoading.value = true;
  try {
    instance.config.regular_payment = value;

    const data = store.getters["instances/all"].find(
      (data) => data.instance.uuid === instance.uuid
    ).instance;
    data.config.regular_payment = value;

    await store.getters["instances/instancesClient"].update(
      UpdateRequest.fromJson({ instance: data })
    );
  } finally {
    isChangeRegularPaymentLoading.value = false;
  }
};

const updateEditValues = async (values) => {
  try {
    const promises = value?.value.map((instance) => {
      const data = store.getters["instances/all"].find(
        (data) => data.instance.uuid === instance.uuid
      ).instance;

      data.config.regular_payment = values["config.regular_payment"] === "True";
      data.created = formatDateToTimestamp(values["date"]);
      if (["ione", "empty"].includes(data.type)) {
        Object.keys(data.data).forEach((nextPaymentDateKey) => {
          if (nextPaymentDateKey.endsWith("next_payment_date")) {
            const lastMonitoringKey = nextPaymentDateKey.replace(
              "next_payment_date",
              "last_monitoring"
            );
            data.data[nextPaymentDateKey] = formatDateToTimestamp(
              values.dueDate
            );
            data.data[lastMonitoringKey] =
              data.data[lastMonitoringKey] +
              (formatDateToTimestamp(values.dueDate) -
                data.data[lastMonitoringKey]);
          }
        });
      }

      return store.getters["instances/instancesClient"].update(
        UpdateRequest.fromJson({ instance: data })
      );
    });
    await Promise.all(promises);
  } catch (e) {
    console.log("error while save", e);
  }
};

watch([filter, customFilter], fetchInstancesDebounce, { deep: true });
watch([options, refetch, searchParam], fetchInstancesDebounce);

watch(instances, () => {
  instances.value.forEach(async ({ account: uuid }) => {
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

watch(
  searchFields,
  (value) => {
    store.commit("appSearch/setFields", value);
  },
  { deep: true }
);
</script>

<script>
export default {
  name: "instances-table",
};
</script>
