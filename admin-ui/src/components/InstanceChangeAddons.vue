<template>
  <v-card style="min-height: 60vh" color="background-light" class="pa-5">
    <v-text-field v-model="searchParam" placeholder="Search..." />

    <nocloud-table
      @update:options="setOptions"
      :loading="isLoading"
      :items="addons"
      :headers="headers"
      :server-items-length="count"
      :server-side-page="page"
      :footer-error="fetchError"
    >
      <template v-slot:[`item.title`]="{ item }">
        <router-link :to="{ name: 'Addon page', params: { uuid: item.uuid } }">
          {{ item.title }}
        </router-link>
      </template>

      <template v-slot:[`item.enabled`]="{ item }">
        <div class="enabled_column">
          <v-switch
            dense
            hide-details
            :input-value="instanceAddons.find((a) => a.uuid === item.uuid)"
            @change="toggleAddon(item)"
          />
        </div>
      </template>
    </nocloud-table>
  </v-card>
</template>

<script setup>
import NocloudTable from "@/components/table.vue";
import { computed, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";
import { debounce } from "@/functions";

const props = defineProps({
  instanceAddons: {},
  instance: {},
});
const { instanceAddons, instance } = toRefs(props);

const emit = defineEmits(["update:options"]);

const store = useStore();

const headers = ref([
  { text: "Title", value: "title" },
  { text: "Group", value: "group" },
  { text: "Enabled", value: "enabled" },
]);
const isFetchLoading = ref(true);
const isCountLoading = ref(true);
const options = ref({});
const fetchError = ref("");
const allAddonsGroups = ref([]);
const page = ref(1);
const count = ref(10);
const searchParam = ref("");

const isLoading = computed(() => store.getters["addons/isLoading"]);
const addons = computed(() => store.getters["addons/all"]);

const requestOptions = computed(() => ({
  filters: { search_param: searchParam.value },
  page: page.value,
  limit: options.value.itemsPerPage,
  field: options.value.sortBy?.[0],
  sort:
    options.value.sortBy?.[0] && options.value.sortDesc?.[0] ? "DESC" : "ASC",
}));

const countOptions = computed(() => ({
  filters: { search_param: searchParam.value },
}));

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

const toggleAddon = (item) => {
  let newAddons = [...instance.value.addons];
  if (instanceAddons.value.find((a) => a.uuid === item.uuid)) {
    newAddons = newAddons.filter((a) => a !== item.uuid);
  } else {
    newAddons.push(item.uuid);
  }

  emit("update", newAddons);
};

watch(searchParam, fetchAddonsDebounced);
watch(options, fetchAddonsDebounced);
</script>

<style scoped></style>
