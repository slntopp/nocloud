<template>
  <div class="pa-4 flex-wrap">
    <div class="d-flex justify-space-between align-center pb-8 pt-4">
      <div class="buttons__inline">
        <v-btn
          color="background-light"
          class="mr-2"
          :to="{ name: 'CategoriesCreate' }"
        >
          create
        </v-btn>
        <confirm-dialog
          :disabled="selectedCategories.length < 1"
          @confirm="deleteSelectedCategories"
        >
          <v-btn
            color="background-light"
            class="mr-8"
            :disabled="selectedCategories.length < 1"
            :loading="isDeleteLoading"
          >
            delete
          </v-btn>
        </confirm-dialog>
      </div>

      <div></div>
    </div>

    <nocloud-table
      v-model="selectedCategories"
      :headers="headers"
      item-key="uuid"
      table-name="categories_table"
      :items="categories"
      :loading="isCategoriesLoading"
    >
      <template v-slot:[`item.title`]="{ item }">
        <router-link
          :to="{ name: 'CategoriesPage', params: { uuid: item.uuid } }"
        >
          {{ getShortName(item.title, 45) }}
        </router-link>
      </template>

      <template v-slot:[`item.type`]="{ item }">
        <v-chip v-if="item.type">{{ item.type }}</v-chip>
      </template>

      <template v-slot:[`item.sorter`]="{ item }">
        <v-skeleton-loader v-if="updatedCategory === item.uuid" type="text" />
        <v-text-field
          v-else
          dense
          hide-details
          style="max-width: 50px"
          type="number"
          :disabled="!!updatedCategory"
          :value="item.sorter"
          @change="updateCategory(item, { key: 'sorter', value: +$event })"
        />
      </template>
    </nocloud-table>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useStore } from "../store";
import NocloudTable from "@/components/table.vue";
import api from "../api";
import ConfirmDialog from "../components/confirmDialog.vue";
import { getShortName } from "../functions";

const store = useStore();

const selectedCategories = ref([]);
const isCategoriesLoading = computed(
  () => store.getters["showcases/isLoading"]
);
const updatedCategory = ref(null);
const isDeleteLoading = ref(false);

const headers = ref([
  { text: "UUID", value: "uuid" },
  { text: "Title", value: "title" },
  { text: "Type", value: "type" },
  { text: "Sorter", value: "sorter" },
]);

const categories = computed(() => store.getters["showcases/categories"]);

onMounted(() => {
  store.dispatch("showcases/fetch");
});

const updateCategory = async (category, { key, value }) => {
  if (updatedCategory.value) {
    return;
  }

  updatedCategory.value = category.uuid;

  try {
    const newCategoryData = { ...category, [key]: value };
    await api.patch(`showcase_categories/${category.uuid}`, newCategoryData);

    category[key] = value;

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Done",
    });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response.data.message || "Error during update category",
    });
  } finally {
    updatedCategory.value = null;
  }
};

const deleteSelectedCategories = async () => {
  if (isDeleteLoading.value || !selectedCategories.value.length) {
    return;
  }

  isDeleteLoading.value = true;

  try {
    await Promise.all(
      selectedCategories.value.map((category) =>
        api.delete(`showcase_categories/${category.uuid}`)
      )
    );

    store.dispatch("showcases/fetch");
    selectedCategories.value = [];

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Done",
    });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response.data.message || "Error during delete category",
    });
  } finally {
    isDeleteLoading.value = false;
  }
};
</script>

<script>
export default {
  name: "CategoriesView",
};
</script>
