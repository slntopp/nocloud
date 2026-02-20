<template>
  <v-card
    color="background-light"
    class="mx-5 px-5 pa-5"
    :loading="isFetchLoading"
  >
    <v-row>
      <v-col cols="3">
        <date-picker
          label="Trial End Date"
          v-model="trialEndDate"
          :clearable="false"
        />
      </v-col>

      <v-col class="d-flex align-center gap-2">
        <confirm-dialog
          :disabled="!isTrialAvailable"
          :loading="isTrialLoading"
          @confirm="setTrialEndDate"
        >
          <v-btn :disabled="!isTrialAvailable" :loading="isTrialLoading"
            >Set Trial End Date
          </v-btn>
        </confirm-dialog>
      </v-col>
    </v-row>

    <nocloud-table
      item-key="id"
      :loading="isLicencesLoading"
      no-hide-uuid
      v-model="selectedLicence"
      :headers="headers"
      :items="licences"
      table-name="sp-licences-table"
      :footer-error="fetchError"
      :options.sync="tableOptions"
      :server-items-length="totalLicences"
    >
      <template v-slot:[`item.app_key`]="{ value }">
        <router-link
          v-if="!isPlansLoading"
          :to="{ name: 'Plan', params: { planId: value } }"
        >
          {{ getPlan(value)?.title }}
        </router-link>
        <v-skeleton-loader type="text" v-else />
      </template>

      <template v-slot:[`item.tariff_key`]="{ value, item }">
        <span v-if="!isPlansLoading">
          {{ getProduct(item.app_key, value, item.licence_metadata) }}
        </span>
        <v-skeleton-loader type="text" v-else />
      </template>

      <template v-slot:[`item.account`]="{ item }">
        <router-link
          v-if="!isAccountsLoading && !isInstancesLoading"
          :to="{
            name: 'Account',
            params: {
              accountId: getAccount(item.licence_metadata?.nocloud_instance)
                ?.uuid,
            },
          }"
        >
          {{ getAccount(item.licence_metadata?.nocloud_instance)?.title }}
        </router-link>
        <v-skeleton-loader type="text" v-else />
      </template>

      <template v-slot:[`item.instance`]="{ item }">
        <router-link
          v-if="!isInstancesLoading"
          :to="{
            name: 'Instance',
            params: { instanceId: item.licence_metadata?.nocloud_instance },
          }"
        >
          {{
            getInstance(item.licence_metadata?.nocloud_instance)?.instance?.title
          }}
        </router-link>

        <v-skeleton-loader type="text" v-else />
      </template>

      <template v-slot:[`item.licence_expires_at`]="{ item }">
        {{ formatDate(item.licence_expires_at) }}
        <v-icon
          v-if="!item.read_only_is_trial && !item.suspended"
          color="green"
        >
          {{ "mdi-check-circle" }}
        </v-icon>

        <v-icon v-if="item.suspended" color="red">
          {{ "mdi-alert-circle" }}
        </v-icon>
      </template>

      <template v-slot:[`item.trial_ends_at`]="{ item }">
        {{ formatDate(item.trial_ends_at) }}
        <v-icon v-if="item.read_only_is_trial" color="green">
          {{ "mdi-check-circle" }}
        </v-icon>
      </template>
    </nocloud-table>
  </v-card>
</template>

<script>
export default {
  name: "licences-tab",
};
</script>

<script setup>
import confirmDialog from "@/components/confirmDialog.vue";
import { computed, onMounted, ref, defineProps, toRefs, watch } from "vue";
import api from "@/api";
import NocloudTable from "@/components/table.vue";
import { useStore } from "@/store";
import useSearch from "@/hooks/useSearch";
import plansAutoComplete from "@/components/ui/plansAutoComplete.vue";
import DatePicker from "@/components/ui/dateTimePicker.vue";
import { timestampToDateTimeLocal, formatDateToTimestamp } from "@/functions";

const props = defineProps(["template", "plan"]);
const { template, plan } = toRefs(props);

const store = useStore();
useSearch({
  name: "b2b-licences",
  defaultLayout: {
    title: "Default",
  },
});

const isFetchLoading = ref(false);
const isLicencesLoading = ref(false);
const isTrialLoading = ref(false);
const selectedLicence = ref([]);
const licences = ref([]);
const fetchError = ref("");
const trialEndDate = ref(
  timestampToDateTimeLocal(new Date(Date.now()).getTime() / 1000),
);
const totalLicences = ref(0);

const plans = ref({});
const isPlansLoading = ref(false);

const accounts = ref({});
const isAccountsLoading = ref(false);

const instances = ref({});
const isInstancesLoading = ref(false);

const tableOptions = ref({
  page: 1,
  itemsPerPage: 10,
  sortBy: [],
  sortDesc: [],
});

const headers = ref([
  { text: "Domain", value: "domain" },
  { text: "Instance", value: "instance" },
  { text: "Account", value: "account" },
  { text: "Plan", value: "app_key" },
  { text: "Tariff Key", value: "tariff_key" },
  { text: "License Expires", value: "licence_expires_at" },
  { text: "Trial Ends", value: "trial_ends_at" },
]);

const filter = computed(() => store.getters["appSearch/filter"]);
const searchParam = computed(() => store.getters["appSearch/param"]);

const sp = computed(() => template.value.uuid);

const searchFields = computed(() =>
  [
    {
      key: "domain",
      title: "Domain",
      type: "input",
    },
    !plan.value && {
      key: "app_key",
      custom: true,
      fetchValue: true,
      title: "Plan",
      label: "Plan",
      clearable: true,
      customParams: { filters: { type: ["bitrix24"] } },
      component: plansAutoComplete,
    },
    {
      key: "is_trial",
      title: "Trial Status",
      type: "select",
      items: [
        { text: "All", value: null },
        { text: "Trial", value: true },
        { text: "Not Trial", value: false },
      ],
    },
    {
      key: "suspended",
      title: "Status",
      type: "select",
      items: [
        { text: "All", value: null },
        { text: "Suspended", value: true },
        { text: "Active", value: false },
      ],
    },
    {
      key: "licence_expires",
      title: "License Expires",
      type: "date",
    },

    {
      key: "trial_expires",
      title: "Trial Expires",
      type: "date",
    },
  ].filter((v) => !!v),
);

const formatDate = (date) => {
  if (!date) return "-";
  if (date === "0001-01-01T00:00:00Z") return "-";
  return new Date(date).toLocaleString();
};

const getPlan = (uuid) => {
  return plans.value[uuid];
};

const getAccount = (instanceUuid) => {
  const accountId = getInstance(instanceUuid)?.account;

  return accounts.value[accountId];
};

const getInstance = (uuid) => {
  return instances.value[uuid];
};

const getProduct = (plan, product, meta) => {
  const planData = plans.value[plan];
  if (!planData) {
    return product;
  }

  const productData = planData.products[product];
  if (!productData) {
    return product;
  }

  return productData.title || product;
};

const buildPayload = () => {
  const payload = {
    page: tableOptions.value.page,
    limit: tableOptions.value.itemsPerPage,
    sort_field: tableOptions.value.sortBy?.[0] || "licence_started_at",
    sort_dir: tableOptions.value.sortDesc?.[0] ? "desc" : "asc",
  };

  if (filter.value.domain || searchParam.value) {
    payload.domain = filter.value.domain || searchParam.value;
  }

  if (filter.value.app_key || plan.value) {
    payload.app_key = filter.value.app_key || plan.value;
  }

  if (Array.isArray(filter.value.is_trial) && filter.value.is_trial.length) {
    if (
      filter.value.is_trial[0] !== null &&
      filter.value.is_trial[0] !== undefined
    ) {
      payload.is_trial = filter.value.is_trial[0];
    }
  }

  if (Array.isArray(filter.value.suspended) && filter.value.suspended.length) {
    if (
      filter.value.suspended[0] !== null &&
      filter.value.suspended[0] !== undefined
    ) {
      payload.suspended = filter.value.suspended[0];
    }
  }

  if (Array.isArray(filter.value.licence_expires)) {
    if (filter.value.licence_expires[0]) {
      payload.licence_expires_from =
        filter.value.licence_expires[0] + "T00:00:00Z";
    }
    if (filter.value.licence_expires[1]) {
      payload.licence_expires_to =
        filter.value.licence_expires[1] + "T23:59:59Z";
    }
  }

  if (Array.isArray(filter.value.trial_expires)) {
    if (filter.value.trial_expires[0]) {
      payload.trial_expires_from = filter.value.trial_expires[0] + "T00:00:00Z";
    }
    if (filter.value.trial_expires[1]) {
      payload.trial_expires_to = filter.value.trial_expires[1] + "T23:59:59Z";
    }
  }

  return payload;
};

const fetchLicences = async () => {
  isLicencesLoading.value = true;
  fetchError.value = "";
  try {
    const res = await api.servicesProviders.action({
      uuid: sp.value,
      action: "list_licences",
      params: {
        payload: buildPayload(),
      },
    });

    const payload = res.meta.payload;
    licences.value = (payload.items || []).map((item) => ({
      ...item,
    }));
    totalLicences.value = payload.total || 0;
  } catch (e) {
    fetchError.value = e.message;
    console.error("Error fetching licences:", e);
  } finally {
    isLicencesLoading.value = false;
  }
};

const isTrialAvailable = computed(
  () => sp.value && selectedLicence.value.length && trialEndDate.value,
);

const setTrialEndDate = async () => {
  if (!selectedLicence.value[0]) return;

  try {
    isTrialLoading.value = true;
    fetchError.value = "";

    await Promise.all(
      selectedLicence.value.map((licence) =>
        api.servicesProviders.action({
          uuid: sp.value,
          action: "set_trial_end_date",
          params: {
            domain: licence.domain,
            app_key: licence.app_key,
            end_date_ts: formatDateToTimestamp(trialEndDate.value),
          },
        }),
      ),
    );

    await fetchLicences();
    selectedLicence.value = [];
    trialEndDate.value = timestampToDateTimeLocal(
      new Date(Date.now()).getTime() / 1000,
    );
  } catch (e) {
    fetchError.value = `Error setting trial end date: ${e.message}`;
    console.error("Error setting trial end date:", e);
  } finally {
    isTrialLoading.value = false;
  }
};

watch(
  [filter, searchParam, tableOptions],
  async () => {
    await fetchLicences();
  },
  { deep: true },
);

watch(licences, () => {
  licences.value.forEach(async ({ app_key: uuid }) => {
    isPlansLoading.value = true;
    try {
      if (!plans.value[uuid]) {
        plans.value[uuid] = api.plans.get(uuid);
        plans.value[uuid] = await plans.value[uuid];
      }
    } catch (e) {
      console.error(`Error fetching plan ${uuid}:`, e);
      delete plans.value[uuid];
    } finally {
      isPlansLoading.value = Object.values(plans.value).some(
        (acc) => acc instanceof Promise,
      );
    }
  });

  licences.value.forEach(async (license) => {
    const instanceId = license.licence_metadata?.nocloud_instance;
    isInstancesLoading.value = true;
    try {
      if (!instances.value[instanceId]) {
        instances.value[instanceId] = store.getters[
          "instances/instancesClient"
        ].get({ uuid: instanceId });

        instances.value[instanceId] = await instances.value[instanceId];
      }
    } catch (e) {
      console.error(`Error fetching instance ${instanceId}:`, e);
      delete instances.value[instanceId];
    } finally {
      isInstancesLoading.value = Object.values(instances.value).some(
        (acc) => acc instanceof Promise,
      );
    }
  });
});

watch(
  isInstancesLoading,
  () => {
    Object.values(instances.value).forEach(async (instance) => {
      const accountId = instance.account;

      isAccountsLoading.value = true;
      try {
        if (!accounts.value[accountId]) {
          accounts.value[accountId] = api.accounts.get(accountId);

          accounts.value[accountId] = await accounts.value[accountId];
        }
      } catch (e) {
        console.error(`Error fetching account ${accountId}:`, e);
        delete accounts.value[accountId];
      } finally {
        isAccountsLoading.value = Object.values(accounts.value).some(
          (acc) => acc instanceof Promise,
        );
      }
    });
  },
  { deep: true },
);

onMounted(() => {
  fetchLicences();
  store.commit("appSearch/setFields", searchFields.value);
});

watch(
  searchFields,
  (value) => {
    store.commit("appSearch/setFields", value);
  },
  { deep: true },
);
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}

pre {
  background-color: rgba(0, 0, 0, 0.05);
  padding: 8px;
  border-radius: 4px;
  max-height: 200px;
  overflow: auto;
}
</style>
