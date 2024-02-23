<template>
  <div class="pa-4">
    <div class="d-flex align-center mb-5">
      <v-btn class="mr-2" :to="{ name: 'Addon create' }"> Create </v-btn>
      <v-btn
        class="mr-2"
        :loading="isDeleteLoading"
        :disabled="!selectedAddons.length"
        @click="deleteSelectedAddons"
      >
        Delete
      </v-btn>
    </div>

    <addons-table
      show-select
      :fetchError="fetchError"
      :loading="isLoading"
      v-model="selectedAddons"
      :items="addons"
      :server-items-length="count"
      :server-side-page="page"
      sort-by="exec"
      sort-desc
      @update:options="setOptions"
    />
  </div>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { useStore } from "@/store";
import api from "@/api";
import { debounce } from "@/functions";
import AddonsTable from "@/components/addonsTable.vue";

const store = useStore();

const selectedAddons = ref([]);
const isDeleteLoading = ref(false);
const count = ref(10);
const page = ref(1);
const isFetchLoading = ref(true);
const isCountLoading = ref(true);
const options = ref({});
const fetchError = ref("");
const allAddonsGroups = ref([]);

const requestOptions = computed(() => ({
  filters: filter.value,
  page: page.value,
  limit: options.value.itemsPerPage,
  field: options.value?.[0],
  sort:
    options.value.sortBy?.[0] && options.value.sortDesc?.[0] ? "DESC" : "ASC",
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
  ];
});

const isLoading = computed(() => store.getters["addons/isLoading"]);
const addons = computed(() => store.getters["addons/all"]);

const deleteSelectedAddons = async () => {
  try {
    isDeleteLoading.value = true;
    await Promise.all(
      selectedAddons.value.map((addon) =>
        api.delete("/billing/addons/" + addon.uuid)
      )
    );
    fetchAddonsDebounced();
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isDeleteLoading.value = false;
  }
};

const setOptions = (newOptions) => {
  page.value = newOptions.page;
  if (JSON.stringify(newOptions) !== JSON.stringify(options.value)) {
    options.value = newOptions;
  }
};

const init = async () => {
  isCountLoading.value = true;
  try {
    const { total, unique } = await store.dispatch(
      "addons/count",
      requestOptions.value
    );
    count.value = +total;
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

watch(
  searchFields,
  (value) => {
    store.commit("appSearch/setFields", value);
  },
  { deep: true }
);

watch(filter, fetchAddonsDebounced, { deep: true });
watch(options, fetchAddonsDebounced);
</script>

<script>
import searchMixin from "@/mixins/search";

export default {
  name: "AddonsView",
  mixins: [searchMixin({ name: "addons-table" })],
};
</script>

<style>
.change_public .v-input {
  margin-top: 0px !important;
}
</style>
