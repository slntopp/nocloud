<template>
  <nocloud-table
    table-name="namespaces"
    :loading="isLoading"
    :items="namespaces"
    :headers="headers"
    :value="value"
    @input="emit('input', $event)"
    :single-select="singleSelect"
    :footer-error="fetchError"
    :server-items-length="total"
    :server-side-page="options.page"
    @update:options="setOptions"
  >
    <template v-slot:[`item.access`]="{ item }">
      <v-chip color="info" v-if="!isAccountsLoading">
        {{ accounts[item.access.namespace]?.title }}
        ({{ item.access.level }})
      </v-chip>
      <v-skeleton-loader type="text" v-else />
    </template>
    <template v-slot:[`item.title`]="{ item }">
      <router-link
        :to="{ name: 'NamespacePage', params: { namespaceId: item.uuid } }"
      >
        {{ getShortName(item.title, 50) }}
      </router-link>
    </template>
  </nocloud-table>
</template>

<script setup>
import nocloudTable from "@/components/table.vue";
import { useStore } from "@/store/";
import { debounce, getShortName } from "@/functions";
import api from "@/api";

const props = defineProps({
  value: {
    type: Array,
    default: () => [],
  },
  singleSelect: {
    type: Boolean,
    default: false,
  },
  refetch: {
    type: Boolean,
    default: false,
  },
});
const { value, singleSelect, refetch } = toRefs(props);

const emit = defineEmits(["input"]);

const store = useStore();

const headers = ref([
  { text: "Title", value: "title" },
  { text: "UUID", value: "uuid" },
  { text: "Access", value: "access" },
]);
const fetchError = ref("");
const total = ref(0);
const options = ref({});
const accounts = ref({});
const isAccountsLoading = ref(false);

onMounted(() => {
  store.commit("appSearch/setFields", searchFields.value);
});

const filter = computed(() => store.getters["appSearch/filter"]);
const searchParam = computed(() => store.getters["appSearch/param"]);
const isLoading = computed(() => store.getters["namespaces/isLoading"]);
const namespaces = computed(() => store.getters["namespaces/all"]);

const requestOptions = computed(() => ({
  filters: {
    ...filter.value,
    search_param: searchParam.value || filter.value.title || undefined,
  },
  page: options.value.page,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy[0],
  sort: options.value.sortBy[0] && options.value.sortDesc[0] ? "DESC" : "ASC",
}));

const searchFields = computed(() => [
  {
    title: "Title",
    key: "title",
    type: "input",
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

const fetchNamespaces = async () => {
  fetchError.value = "";
  try {
    const { count } = await store.dispatch(
      "namespaces/fetch",
      requestOptions.value
    );
    total.value = +count;
  } catch (err) {
    fetchError.value = "Can't reach the server";
    if (err.response && err.response.data.message) {
      fetchError.value += `: [ERROR]: ${err.response.data.message}`;
    } else {
      fetchError.value += `: [ERROR]: ${err.toJSON().message}`;
    }
  }
};

const fetchNamespacesDebounce = debounce(fetchNamespaces);

const fetchAccounts = () => {
  namespaces.value.forEach(async ({ access: { namespace: uuid } }) => {
    if (!uuid) {
      return;
    }

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
};

watch(filter, fetchNamespacesDebounce, { deep: true });
watch(searchParam, fetchNamespacesDebounce);
watch(options, fetchNamespacesDebounce);
watch(refetch, fetchNamespacesDebounce);

watch(namespaces, fetchAccounts);
</script>

<script>
import search from "@/mixins/search";
import { computed, onMounted, ref, toRefs, watch } from "vue";

export default {
  name: "namespaces-table",
  mixins: [search({ name: "namespaces-table" })],
};
</script>

<style></style>
