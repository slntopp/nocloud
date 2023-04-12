<template>
  <nocloud-table
    :name="tableName"
    class="mt-4"
    :items="logs"
    :headers="headers"
    :loading="isLoading"
    :footer-error="fetchError"
    :server-items-length="count"
    :server-side-page="page"
    item-key="Id"
    @update:options="onUpdateOptions"
  >
    <template v-slot:[`item.ts`]="{ value }">
      {{ new Date(new Date(1970, 0, 1).setSeconds(value)).toLocaleString() }}
    </template>
  </nocloud-table>
</template>

<script setup>
import { defineProps, toRefs, ref, onMounted, computed } from "vue";
import nocloudTable from "@/components/table.vue";
import api from "@/api";

const props = defineProps(["tableName", "accountId", "uuid"]);
const { tableName, accountId, uuid } = toRefs(props);

const count = ref(10);
const logs = ref([]);
const page = ref(1);
const isFetchLoading = ref(false);
const isCountLoading = ref(false);
const fetchError = ref("");

const headers = ref([
  { text: "Id", value: "id" },
  { text: "UUID", value: "uuid" },
  { text: "Entity", value: "entity" },
  { text: "Scope", value: "scope" },
  { text: "Action", value: "action" },
  { text: "NC", value: "nc" },
  { text: "Requestor", value: "requestor" },
  { text: "TS", value: "ts" },
]);

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

const isLoading = computed(() => {
  return isFetchLoading.value || isCountLoading.value;
});

onMounted(() => {
  init();
});
</script>
