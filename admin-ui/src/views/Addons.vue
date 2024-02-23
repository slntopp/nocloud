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
      :loading="isLoading"
      v-model="selectedAddons"
      :items="filteredAddons"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useStore } from "@/store";
import api from "@/api";
import { compareSearchValue } from "@/functions";
import AddonsTable from "@/components/addonsTable.vue";

const store = useStore();

const selectedAddons = ref([]);
const isDeleteLoading = ref(false);

onMounted(() => {
  fetchAddons();
});

const searchParam = computed(() => store.getters["appSearch/param"]);
const filter = computed(() => store.getters["appSearch/filter"]);
const searchFields = computed(() => {
  return [
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

const filteredAddons = computed(() => {
  const filtered = addons.value.filter((addon) => {
    return Object.keys(filter.value || {}).every((key) => {
      return compareSearchValue(
        addon[key],
        filter.value[key],
        searchFields.value?.find((f) => f.key === key)
      );
    });
  });

  if (searchParam.value) {
    return filtered.filter(
      (a) =>
        !searchParam.value ||
        a.title.toLowerCase().includes(searchParam.value.toLowerCase()) ||
        a.group.toLowerCase().includes(searchParam.value.toLowerCase())
    );
  }

  return filtered;
});

const allAddonsGroups = computed(() => [
  ...new Set(addons.value.map((a) => a.group)),
]);

const fetchAddons = () => {
  store.dispatch("addons/fetch");
};

const deleteSelectedAddons = async () => {
  try {
    isDeleteLoading.value = true;
    await Promise.all(
      selectedAddons.value.map((addon) =>
        api.delete("/billing/addons/" + addon.uuid)
      )
    );
    fetchAddons();
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isDeleteLoading.value = false;
  }
};
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
  name: "AddonsView",
  mixins: [searchMixin({ name: "addons-table" })],
};
</script>

<style>
.change_public .v-input {
  margin-top: 0px !important;
}
</style>
