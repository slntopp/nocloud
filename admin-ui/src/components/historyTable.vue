<template>
  <nocloud-table
    :table-name="tableName"
    class="mt-4"
    :items="logs"
    :headers="headers"
    :loading="isLoading"
    :footer-error="fetchError"
    :server-items-length="count"
    :server-side-page="page"
    sort-by="ts"
    sort-desc
    item-key="id"
    @update:options="setOptions"
    show-expand
    :expanded.sync="expanded"
    no-hide-uuid
  >
    <template v-slot:[`item.ts`]="{ value }">
      {{ formatSecondsToDate(value, true) }}
    </template>
    <template v-slot:[`item.requestor`]="{ value }">
      <router-link
        v-if="!isAccountsLoading"
        :to="{ name: 'Account', params: { accountId: value } }"
      >
        {{ getShortName(getAccount(value)?.title) }}
      </router-link>
      <v-skeleton-loader type="text" v-else />
    </template>
    <template v-slot:[`item.uuid`]="{ item }">
      <router-link
        v-if="getEntityByUuid(item).route"
        :to="getEntityByUuid(item).route"
      >
        {{
          getShortName(
            `${getEntityByUuid(item).item?.title || getEntityByUuid(item).item}`
          )
        }}
        {{ `(${getEntityByUuid(item).type})` }}
      </router-link>
    </template>
    <template
      v-if="services.length && instances.length"
      v-slot:expanded-item="{ headers, item }"
    >
      <td :colspan="headers.length" style="padding: 0">
        <nocloud-table
          table-name="log-operations"
          :server-items-length="-1"
          hide-default-footer
          :headers="operationHeaders"
          :items="getDiffItems(item)"
        />
      </td>
    </template>
  </nocloud-table>
</template>
<script setup>
import { toRefs, ref, computed, watch } from "vue";
import nocloudTable from "@/components/table.vue";
import api from "@/api";
import { useStore } from "@/store";
import { debounce, formatSecondsToDate, getShortName } from "@/functions";

const props = defineProps({
  tableName: {},
  accountId: {},
  uuid: {},
  hideRequestor: { type: Boolean, default: false },
  hideUuid: { type: Boolean, default: false },
});
const { tableName, accountId, uuid, hideRequestor, hideUuid } = toRefs(props);

const count = ref(10);
const logs = ref([]);
const page = ref(1);
const isFetchLoading = ref(false);
const isCountLoading = ref(false);
const fetchError = ref("");
const expanded = ref([]);
const options = ref({});
const scopeItems = ref([]);
const actionItems = ref([]);
const isAccountsLoading = ref(false);
const accounts = ref({});

const store = useStore();

const headers = computed(() => [
  { text: "Id", value: "id" },
  !hideRequestor.value && { text: "Account (Requestor)", value: "requestor" },
  !hideUuid.value && { text: "Entity", value: "uuid" },
  { text: "Scope", value: "scope" },
  { text: "Action", value: "action" },
  { text: "Timestamp", value: "ts" },
]);

const operationHeaders = ref([
  { text: "Operation", value: "op" },
  { text: "Path", value: "path" },
  { text: "Old value", value: "oldValue" },
  { text: "New value", value: "newValue" },
]);

const getDiffItems = (item) => {
  try {
    const replacedKeys = [
      { to: "}", from: "}," },
      { to: "'", from: '"' },
      { to: "`", from: '"' },
      { to: "'", from: '"' },
    ];

    let replaced = item.snapshot.diff;
    replacedKeys.forEach(
      ({ to, from }) => (replaced = replaced.replaceAll(to, from))
    );
    const operations = JSON.parse(`[${replaced.slice(0, -1)}]`);
    return operations.map((op, index, arr) => {
      op.oldValue = op.value;
      if (op.value && index !== arr.length - 1) {
        op.newValue = arr[index + 1]?.value;
      } else if (op.value) {
        op.newValue = getEntityByUuid(item).item;
        op.path
          .split("/")
          .filter((k) => !!k)
          .forEach((subKey) => {
            op.newValue = op.newValue[subKey];
          });
      }
      return op;
    });
  } catch (e) {
    return [];
  }
};

const getEntityByUuid = (item) => {
  switch (item.entity) {
    case "Instances": {
      return {
        route: { name: "Instance", params: { instanceId: item.uuid } },
        item: getInstance(item.uuid),
        type: "Instance",
      };
    }
    case "Services": {
      return {
        route: { name: "Service", params: { serviceId: item.uuid } },
        item: getService(item.uuid),
        type: "Service",
      };
    }
    case "ServicesProviders": {
      return {
        route: { name: "ServicesProvider", params: { uuid: item.uuid } },
        item: getServiceProvider(item.uuid),
        type: "Service provider",
      };
    }
    default: {
      return { route: null, item: null };
    }
  }
};

const setOptions = (newOptions) => {
  if (JSON.stringify(newOptions) !== JSON.stringify(options.value)) {
    options.value = newOptions;
    page.value = newOptions.page;
  }
};

const fetchLogs = async () => {
  init();
  isFetchLoading.value = true;
  try {
    logs.value = (await api.logging.list(requestOptions.value)).events;
  } finally {
    isFetchLoading.value = false;
  }
};

const fetchLogsDebounced = debounce(fetchLogs);

const updateProps = async () => {
  page.value = 1;
  try {
    await fetchLogsDebounced(options.value);
  } catch (e) {
    fetchError.value = e.message;
  }
};

const init = async () => {
  isCountLoading.value = true;
  try {
    const { total, unique } = await api.logging.count(requestOptions.value);
    count.value = +total;

    if (!actionItems.value.length || !scopeItems.value.length) {
      actionItems.value = unique.actions;
      scopeItems.value = unique.scopes;

      store.commit("appSearch/pushFields", searchFields.value);
    }
  } finally {
    isCountLoading.value = false;
  }
};

const getAccount = (uuid) => {
  return accounts.value[uuid] || uuid;
};

const getInstance = (uuid) => {
  return instances.value.find((i) => i.uuid === uuid) || uuid;
};

const getService = (uuid) => {
  return services.value.find((s) => s.uuid === uuid) || uuid;
};

const getServiceProvider = (uuid) => {
  return sps.value.find((s) => s.uuid === uuid) || uuid;
};

const isLoading = computed(() => {
  return isFetchLoading.value || isCountLoading.value;
});

const requestOptions = computed(() => ({
  page: page.value,
  limit: options.value.itemsPerPage,
  requestor: accountId.value,
  uuid: uuid.value,
  field: options.value.sortBy[0],
  sort: options.value.sortBy[0] && options.value.sortDesc[0] ? "DESC" : "ASC",
  filters: {
    action: action.value,
    scope: scope.value,
    path: path.value,
  },
}));

const searchFields = computed(() => {
  return [
    {
      items: actionItems.value,
      type: "select",
      key: "action",
      title: "Action",
    },
    { type: "select", items: scopeItems.value, key: "scope", title: "Scopes" },
    { type: "input", key: "path", title: "Path" },
  ];
});

const services = computed(() => store.getters["services/all"]);
const sps = computed(() => store.getters["servicesProviders/all"]);
const instances = computed(() => store.getters["services/getInstances"]);
const filter = computed(() => store.getters["appSearch/filter"]);
const path = computed(() => filter.value.path || undefined);
const action = computed(() =>
  filter.value.action?.length ? filter.value.action : undefined
);
const scope = computed(() =>
  filter.value.scope?.length ? filter.value.scope : undefined
);

watch(accountId, () => updateProps());
watch(uuid, () => updateProps());
watch(filter, fetchLogsDebounced, { deep: true });
watch(options, fetchLogsDebounced);
watch(logs, () => {
  logs.value.forEach(async ({ requestor: uuid }) => {
    isAccountsLoading.value = true;
    try {
      if (!accounts.value[uuid]) {
        accounts.value[uuid] = api.accounts.get(uuid);
        accounts.value[uuid] = await accounts.value[uuid];
      }
    } catch {
      accounts.value[uuid] = undefined;
    } finally {
      isAccountsLoading.value = Object.values(accounts.value).some(
        (acc) => acc instanceof Promise
      );
    }
  });
});
</script>
