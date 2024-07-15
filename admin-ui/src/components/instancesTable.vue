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
          :input-value="item.config.auto_renew"
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
          :input-value="item.config.regular_payment"
          @change="changeRegularPayment(item, $event)"
        />
      </div>
    </template>

    <template v-slot:[`item.state.meta.networking`]="{ item }">
      <template v-if="!item.state?.meta.networking?.public">-</template>
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

const props = defineProps({
  tableName: { type: String, default: "instances-table" },
  value: { type: Array, required: false },
  refetch: { type: Boolean, default: false },
  headers: {},
  showSelect: { type: Boolean, default: true },
  openInNewTab: { type: Boolean, default: false },
  noSearch: { type: Boolean, default: false },
});
const { tableName, value, refetch, showSelect, openInNewTab } = toRefs(props);

const emit = defineEmits(["input"]);

const store = useStore();
const { defaultCurrency, convertTo } = useCurrency();

const page = ref(1);
const options = ref({});
const fetchError = ref("");

const isUniqueFetched = ref(false);
const uniqueProducts = ref([]);
const uniqueLocations = ref([]);

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

const listOptions = computed(() => ({
  filters: filter.value,
  page: page.value,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy?.[0],
  sort:
    options.value.sortBy?.[0] && options.value.sortDesc?.[0] ? "DESC" : "ASC",
}));

const searchFields = () =>
  computed(() => [
    {
      key: "title",
      title: "Title",
      type: "input",
    },
    {
      key: "period",
      items: uniqueProducts.value,
      title: "Period",
      type: "select",
    },
    {
      key: "sp",
      items: [],
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
      key: "access",
      type: "select",
      title: "Account",
      items: [],
    },
    {
      key: "product",
      items: [],
      type: "select",
      title: "Product",
    },
    {
      key: "state",
      items: [
        "INIT",
        "RUNNING",
        "STOPPED",
        "PENDING",
        "OPERATION",
        "SUSPENDED",
        "UNKNOWN",
        "DELETED",
        "ERROR",
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
      key: "billingPlan.title",
      type: "select",
      title: "Billing plan",
      items: [],
    },
    { title: "Due date", key: "dueDate", type: "date" },
    { title: "NCU price", key: "price", type: "number-range" },
    { title: "Email", key: "email", type: "input" },
    { title: "Date", key: "date", type: "date" },
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
  console.log(newOptions);
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
    const { unique } = await api.get("/instances/count");
    uniqueLocations.value = unique.locations;
    uniqueProducts.value = unique.products;

    isUniqueFetched.value = true;
  } catch {
    isUniqueFetched.value = false;
  }
};

const fetchInstancesDebounce = debounce(fetchInstances, 100);

const getPeriod = (instance) => {
  if (isInstancePayg(instance)) {
    return "PayG";
  } else if (
    instance.resources.period &&
    !["ovh", "opensrs"].includes(instance.type)
  ) {
    const text = instance.resources.period > 1 ? "months" : "month";
    return `${instance.resources.period} ${text}`;
  } else if (instance.type === "opensrs") {
    return getBillingPeriod(
      Object.values(instance.billingPlan.resources || {})[0]?.period || 0
    );
  }
  const period = getBillingPeriod(
    Object.values(instance.billingPlan.products || {})[0]?.period || 0
  );

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

//remake
const changeRegularPayment = async (instance, value) => {
  isChangeRegularPaymentLoading.value = true;
  try {
    const tempService = JSON.parse(
      JSON.stringify(this.services?.find((s) => s.uuid === instance.service))
    );
    const igIndex = tempService.instancesGroups.findIndex((ig) =>
      ig.instances?.find((i) => i.uuid === instance.uuid)
    );
    const instanceIndex = tempService.instancesGroups[
      igIndex
    ].instances.findIndex((i) => i.uuid === instance.uuid);

    instance.config.regular_payment = value;

    tempService.instancesGroups[igIndex].instances[instanceIndex] = instance;
    await api.services._update(tempService);
  } finally {
    isChangeRegularPaymentLoading.value = false;
  }
};

const updateEditValues = async (values) => {
  try {
    const promises = value?.value.map((instance) => {
      const tempService = JSON.parse(
        JSON.stringify(this.services?.find((s) => s.uuid === instance.service))
      );
      const igIndex = tempService.instancesGroups.findIndex((ig) =>
        ig.instances?.find((i) => i.uuid === instance.uuid)
      );
      const instanceIndex = tempService.instancesGroups[
        igIndex
      ].instances.findIndex((i) => i.uuid === instance.uuid);

      instance.config.regular_payment =
        values["config.regular_payment"] === "True";
      instance.created = formatDateToTimestamp(values["date"]);
      if (["ione", "empty"].includes(instance.type)) {
        Object.keys(instance.data).forEach((nextPaymentDateKey) => {
          if (nextPaymentDateKey.endsWith("next_payment_date")) {
            const lastMonitoringKey = nextPaymentDateKey.replace(
              "next_payment_date",
              "last_monitoring"
            );
            instance.data[nextPaymentDateKey] = formatDateToTimestamp(
              values.dueDate
            );
            instance.data[lastMonitoringKey] =
              instance.data[lastMonitoringKey] +
              (formatDateToTimestamp(values.dueDate) -
                instance.data[lastMonitoringKey]);
          }
        });
      }

      tempService.instancesGroups[igIndex].instances[instanceIndex] = instance;

      return api.services._update(tempService);
    });
    await Promise.all(promises);
  } catch (e) {
    console.log("error while save", e);
  }
};

watch(filter, fetchInstancesDebounce, { deep: true });
watch(options, fetchInstancesDebounce);
watch(refetch, fetchInstancesDebounce);

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
import searchMixin from "@/mixins/search";

export default {
  name: "instances-table",
  mixins: [
    searchMixin({
      name: "instances-table",
      defaultLayout: {
        title: "Default",
        filter: {
          state: [
            "INIT",
            "RUNNING",
            "STOPPED",
            "PENDING",
            "OPERATION",
            "SUSPENDED",
            "UNKNOWN",
            "ERROR",
          ],
        },
      },
    }),
  ],
};
</script>
