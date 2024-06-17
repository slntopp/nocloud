<template>
  <nocloud-table
    :value="value"
    @input="emit('input', $event)"
    @update:options="setOptions"
    :loading="isLoading"
    :items="addons"
    :headers="headers"
    :show-select="showSelect"
    :table-name="tableName"
    :server-items-length="count"
    :server-side-page="page"
    :sort-desc="sortDesc"
    :sort-by="sortBy"
    :footer-error="fetchError"
  >
    <template v-slot:[`item.title`]="{ item }">
      <router-link :to="{ name: 'Addon page', params: { uuid: item.uuid } }">
        {{ item.title }}
      </router-link>
    </template>

    <template v-slot:[`item.public`]="{ item }">
      <div class="change_public">
        <v-switch
          :readonly="!editable"
          :loading="updatingAddonUuid === item.uuid"
          dense
          hide-details
          :disabled="!!updatingAddonUuid"
          :input-value="item.public"
          @change="updateAddon(item, { key: 'public', value: $event })"
        />
      </div>
    </template>
  </nocloud-table>
</template>

<script setup>
import NocloudTable from "@/components/table.vue";
import { computed, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import { debounce } from "@/functions";

const props = defineProps({
  value: {},
  tableName: { type: String, default: "addons-table" },
  showSelect: { type: Boolean, default: false },
  sortBy: {},
  sortDesc: {},
  editable: { type: Boolean, default: false },
  refetch: { type: Boolean, default: false },
});
const { showSelect, tableName, value, sortBy, sortDesc, refetch } =
  toRefs(props);

const emit = defineEmits(["input", "update:options"]);

const store = useStore();

const headers = ref([
  { text: "UUID", value: "uuid" },
  { text: "Title", value: "title" },
  { text: "Group", value: "group" },
  { text: "Public", value: "public" },
]);
const updatingAddonUuid = ref(false);
const isFetchLoading = ref(true);
const isCountLoading = ref(true);
const options = ref({});
const fetchError = ref("");
const allAddonsGroups = ref([]);
const page = ref(1);
const count = ref(10);

const isLoading = computed(() => store.getters["addons/isLoading"]);
const addons = computed(() => store.getters["addons/all"]);

const requestOptions = computed(() => ({
  filters: Object.keys(filter.value).reduce((newFilter, key) => {
    if (!filter.value[key]) {
      newFilter[key] = undefined;
    } else {
      newFilter[key] = filter.value[key];
    }

    return newFilter;
  }, {}),
  page: page.value,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy?.[0],
  sort:
    options.value.sortBy?.[0] && options.value.sortDesc?.[0] ? "DESC" : "ASC",
}));

const countOptions = computed(() => ({
  filters: filter.value,
}));

const filter = computed(() => ({
  ...store.getters["appSearch/filter"],
}));
const searchFields = computed(() => {
  return [
    {
      key: "title",
      title: "Title",
      type: "input",
    },
    {
      key: "group",
      items: allAddonsGroups.value,
      title: "Group",
      type: "select",
    },
    {
      key: "system",
      title: "System",
      type: "logic-select",
    },
  ];
});

const setOptions = (newOptions) => {
  page.value = newOptions.page;
  if (JSON.stringify(newOptions) !== JSON.stringify(options.value)) {
    options.value = newOptions;
  }
};

const init = async () => {
  isCountLoading.value = true;
  try {
    const data = await store.dispatch("addons/count", countOptions.value);
    const { unique, total } = data.toJson();

    count.value = Number(total);
    allAddonsGroups.value = unique.groups;
  } finally {
    isCountLoading.value = false;
  }
};

const fetchAddons = async () => {
  init();
  isFetchLoading.value = true;
  fetchError.value = "";
  try {
    await store.dispatch("addons/fetch", requestOptions.value);
  } catch (e) {
    fetchError.value = e.message;
  } finally {
    isFetchLoading.value = false;
  }
};

const fetchAddonsDebounced = debounce(fetchAddons, 100);

const updateAddon = async (item, { key, value }) => {
  try {
    updatingAddonUuid.value = item.uuid;
    await store.getters["addons/addonsClient"].update({
      ...item,
      [key]: value,
    });
    item.public = value;
  } catch (e) {
    console.log(e);
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    updatingAddonUuid.value = "";
  }
};

watch(
  searchFields,
  (value) => {
    store.commit("appSearch/setFields", value);
  },
  { deep: true }
);

watch(filter, fetchAddonsDebounced, { deep: true });
watch(options, fetchAddonsDebounced);
watch(refetch, fetchAddonsDebounced);
</script>

<script>
import searchMixin from "@/mixins/search";

export default {
  name: "AddonsView",
  mixins: [searchMixin({ name: "addons-table" })],
};
</script>

<style scoped></style>
