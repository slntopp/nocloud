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
      <template v-slot:[`item.name`]="{ item }">
        <router-link
          :to="{ name: 'CategoriesPage', params: { uuid: item.name } }"
        >
          {{ getShortName(item.name, 45) }}
        </router-link>
      </template>

      <template v-slot:[`item.type`]="{ item }">
        <v-chip v-if="item.type">{{ item.type }}</v-chip>
      </template>

      <template v-slot:[`item.sorter`]="{ item }">
        <v-skeleton-loader v-if="updatedCategory === item.name" type="text" />
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
import { computed, onMounted, ref, watch } from "vue";
import { useStore } from "../store";
import NocloudTable from "@/components/table.vue";
import api from "../api";
import ConfirmDialog from "../components/confirmDialog.vue";
import { getShortName } from "../functions";

const SHOWCASE_CATEGORIES_SETTINGS_KEY = "showcase-categories";

const store = useStore();

const categories = ref([]);
const selectedCategories = ref([]);
const isCategoriesLoading = computed(() => store.getters["settings/isLoading"]);
const updatedCategory = ref(null);
const isDeleteLoading = ref(false);

const headers = ref([
  { text: "UUID", value: "uuid" },
  { text: "Name", value: "name" },
  { text: "Type", value: "type" },
  { text: "Sorter", value: "sorter" },
]);

const originalSettings = computed(() =>
  store.getters["settings/all"].find(
    (v) => v.key === SHOWCASE_CATEGORIES_SETTINGS_KEY
  )
);

onMounted(() => {
  setCategoriesFromSettings();
});

const setCategoriesFromSettings = () => {
  try {
    const settings =
      originalSettings.value && JSON.parse(originalSettings.value.value);

    if (Array.isArray(settings)) {
      categories.value = settings.map((c) => ({ ...c, uuid: c.name }));
    } else {
      categories.value = [];
    }
  } catch (e) {
    categories.value = [];
  }
};

const updateCategory = async (category, { key, value }) => {
  if (updatedCategory.value) {
    return;
  }

  updatedCategory.value = category.name;

  const newCategories = categories.value.map((c) => {
    if (c.name === category.name) {
      return { ...c, [key]: value };
    }

    return c;
  });

  try {
    let data = JSON.parse(JSON.stringify(originalSettings.value));

    data.value = JSON.stringify(newCategories);

    await api.settings.addKey(SHOWCASE_CATEGORIES_SETTINGS_KEY, data);

    categories.value = newCategories;
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

  const newCategories = categories.value.filter(
    (c) => !selectedCategories.value.find((s) => s.name === c.name)
  );

  try {
    let data = JSON.parse(JSON.stringify(originalSettings.value));

    data.value = JSON.stringify(newCategories);

    await api.settings.addKey(SHOWCASE_CATEGORIES_SETTINGS_KEY, data);

    categories.value = newCategories;
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

watch(originalSettings, () => {
  setCategoriesFromSettings();
});
</script>

<script>
export default {
  name: "CategoriesView",
};
</script>
