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
    item-key="id"
    @update:options="onUpdateOptions"
    show-expand
    :expanded.sync="expanded"
    no-hide-uuid
  >
    <template v-slot:[`item.ts`]="{ value }">
      {{ new Date(new Date(1970, 0, 1).setSeconds(value)).toLocaleString() }}
    </template>
    <template v-slot:[`item.requestor`]="{ value }">
      <router-link :to="{ name: 'Account', params: { accountId: value } }">
        {{ getAccount(value)?.title }}
      </router-link>
    </template>
    <template v-slot:[`item.uuid`]="{ item, value }">
      <router-link
        v-if="item.entity === 'Instances'"
        :to="{ name: 'Instance', params: { instanceId: value } }"
      >
        {{ getInstance(value)?.title }}
      </router-link>
      <router-link
        v-else
        :to="{ name: 'Service', params: { serviceId: value } }"
      >
        {{ getService(value)?.title }}
      </router-link>
    </template>
    <template v-slot:expanded-item="{ headers, item }">
      <td :colspan="headers.length" style="padding: 0">
        <nocloud-table
          :server-items-length="-1"
          hide-default-footer
          :headers="operationHeaders"
          :items="getDiffItems(item.snapshot.diff)"
        />
      </td>
    </template>
  </nocloud-table>
</template>

<script setup>
import { toRefs, ref, onMounted, computed } from "vue";
import nocloudTable from "@/components/table.vue";
import api from "@/api";
import { useStore } from "@/store";

const props = defineProps({
  tableName: {},
  accountId: {},
  uuid: {},
  hideRequestor: { type: Boolean, default: false },
    hideUuid: { type: Boolean, default: false },
});
const { tableName, accountId, uuid, hideRequestor,hideUuid } = toRefs(props);

const count = ref(10);
const logs = ref([]);
const page = ref(1);
const isFetchLoading = ref(false);
const isCountLoading = ref(false);
const fetchError = ref("");
const expanded = ref([]);

const store = useStore();

const headers = computed(() => [
  { text: "Id", value: "id" },
  !hideRequestor.value && { text: "Requestor", value: "requestor" },
  !hideUuid.value && { text: "Account or service", value: "uuid" },
  { text: "Entity", value: "entity" },
  { text: "Scope", value: "scope" },
  { text: "Action", value: "action" },
  { text: "TS", value: "ts" },
]);

const operationHeaders = ref([
  { text: "Operation", value: "op" },
  { text: "Path", value: "path" },
  { text: "Value", value: "value" },
]);

const getDiffItems = (diff) => {
  try {
    const replacedKeys = [
      { to: "}", from: "}," },
      { to: "'", from: '"' },
      { to: "`", from: '"' },
      { to: "'", from: '"' },
    ];

    let replaced = diff;
    replacedKeys.forEach(
      ({ to, from }) => (replaced = replaced.replaceAll(to, from))
    );
    return JSON.parse(`[${replaced.slice(0, -1)}]`);
  } catch (e) {
    return [];
  }
};

const onUpdateOptions = async (options) => {
  if (count.value === 0) {
    return;
  }

  isFetchLoading.value = true;
  try {
    logs.value = (
      await api.logging.list({
        requestor: accountId.value,
        uuid: uuid.value,
        page: options.page,
        limit: options.itemsPerPage,
      })
    ).events;
  } finally {
    isFetchLoading.value = false;
  }
};

const updateProps = async () => {
  await init();
  onUpdateOptions({});
};

const init = async () => {
  isCountLoading.value = true;
  try {
    count.value = +(
      await api.logging.count({ requestor: accountId.value, uuid: uuid.value })
    ).total;
  } finally {
    isCountLoading.value = false;
  }
};

const getAccount = (uuid) => {
  return accounts.value.find((acc) => acc.uuid === uuid);
};

const getInstance = (uuid) => {
  return instances.value.find((i) => i.uuid === uuid);
};

const getService = (uuid) => {
  return services.value.find((s) => s.uuid === uuid);
};

const isLoading = computed(() => {
  return isFetchLoading.value || isCountLoading.value;
});

const accounts = computed(() => store.getters["accounts/all"]);
const services = computed(() => store.getters["services/all"]);
const instances = computed(() => store.getters["services/getInstances"]);

onMounted(() => {
  init();
});

watch(accountId, () => updateProps());
watch(uuid, () => updateProps());
</script>
