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
    @update:options="onUpdateOptions"
    show-expand
    :expanded.sync="expanded"
    no-hide-uuid
    :itemsPerPageOptions="itemsPerPageOptions"
  >
    <template v-slot:[`item.ts`]="{ value }">
      {{ new Date(new Date(1970, 0, 1).setSeconds(value)).toLocaleString() }}
    </template>
    <template v-slot:[`item.requestor`]="{ value }">
      <router-link :to="{ name: 'Account', params: { accountId: value } }">
        {{ getAccount(value)?.title }}
      </router-link>
    </template>
    <template v-slot:[`item.uuid`]="{ item }">
      <router-link :to="getEntityByUuid(item).route">
        {{
          `${
            getEntityByUuid(item).item?.title || getEntityByUuid(item).item
          } (${getEntityByUuid(item).type})`
        }}
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
import { toRefs, ref, onMounted, computed, watch } from "vue";
import nocloudTable from "@/components/table.vue";
import api from "@/api";
import { useStore } from "@/store";

const props = defineProps({
  tableName: {},
  accountId: {},
  uuid: {},
  path: {},
  hideRequestor: { type: Boolean, default: false },
  hideUuid: { type: Boolean, default: false },
});
const { tableName, accountId, uuid, hideRequestor, hideUuid, path } =
  toRefs(props);

const count = ref(10);
const logs = ref([]);
const page = ref(1);
const isFetchLoading = ref(false);
const isCountLoading = ref(false);
const fetchError = ref("");
const expanded = ref([]);
const options = ref({});
const itemsPerPageOptions = ref([5, 10, 15, 25]);

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

const onUpdateOptions = async (newOptions) => {
  options.value = newOptions;
  page.value = newOptions.page;
  init();
  isFetchLoading.value = true;
  try {
    logs.value = (await api.logging.list(requestOptions.value)).events;
  } finally {
    isFetchLoading.value = false;
  }
};

const updateProps = async () => {
  page.value = 1;
  try {
    await init();
    await onUpdateOptions(options.value);
  } catch (e) {
    fetchError.value = e.message;
  }
};

const init = async () => {
  isCountLoading.value = true;
  try {
    count.value = +(await api.logging.count(requestOptions.value)).total;
  } finally {
    isCountLoading.value = false;
  }
};

const getAccount = (uuid) => {
  return accounts.value.find((acc) => acc.uuid === uuid) || uuid;
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

const getFilterItems = async () => {
  const { unique } = await api.logging.count({});
  const actionItems = unique.actions.map((a) => ({ title: a, uuid: a }));
  const scopeItems = unique.scopes.map((a) => ({ title: a, uuid: a }));

  store.commit("appSearch/pushVariant", {
    key: "action",
    value: {
      items: actionItems,
      isArray: true,
      key: "action",
      title: "Action",
    },
  });
  store.commit("appSearch/pushVariant", {
    key: "scope",
    value: {
      isArray: true,
      items: scopeItems,
      key: "scope",
      title: "Scopes",
    },
  });
  if (Object.keys(store.getters["appSearch/customParams"]).length === 0) {
    const hiddenActions = ["monitoring", "regions"];

    const defaultCustomParams = [];
    actionItems.forEach(({ title, uuid }) => {
      if (!hiddenActions.includes(title)) {
        defaultCustomParams.push({ title, value: uuid });
      }
    });
    store.commit("appSearch/setCustomParams", { action: defaultCustomParams });
  }
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
    action: searchParams.value.action?.map(({ value }) => value) || undefined,
    scope: searchParams.value.scope?.map(({ value }) => value) || undefined,
    path: path.value || undefined,
  },
}));

const accounts = computed(() => store.getters["accounts/all"]);
const services = computed(() => store.getters["services/all"]);
const sps = computed(() => store.getters["servicesProviders/all"]);
const instances = computed(() => store.getters["services/getInstances"]);
const searchParams = computed(() => store.getters["appSearch/customParams"]);

onMounted(() => {
  getFilterItems();
});

watch(accountId, () => updateProps());
watch(uuid, () => updateProps());
watch(searchParams, () => updateProps(), { deep: true });
watch(path, () => updateProps());
</script>
