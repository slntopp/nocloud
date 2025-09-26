<template>
  <div class="pa-10 h-100 w-100">
    <h1 v-if="!isEdit" class="page__title">{{ "Create Category" }}</h1>

    <v-form ref="categoryForm" v-model="formValid">
      <v-row>
        <v-col cols="6">
          <v-text-field
            v-model="categoryData.name"
            label="Category Name"
            :rules="nameRules"
            prepend-icon="mdi-tag"
            required
          />
        </v-col>

        <v-col cols="3">
          <v-select
            v-model="categoryData.type"
            label="Category Type"
            :items="categoryTypes"
            prepend-icon="mdi-shape"
          />
        </v-col>

        <v-col cols="3">
          <v-text-field
            v-model.number="categoryData.sorter"
            label="Sort Order"
            type="number"
            min="0"
            prepend-icon="mdi-sort-numeric-variant"
          />
        </v-col>
      </v-row>

      <v-card color="background-light" class="mt-6" variant="outlined">
        <v-card-title class="text-h6 bg-primary">
          <v-icon left>mdi-translate</v-icon>
          Translations
        </v-card-title>

        <v-card-text class="pa-4">
          <v-row class="mb-4">
            <v-col cols="2">
              <v-text-field
                v-model="newLocale"
                label="Language Code"
                placeholder="en, ru, de..."
                density="compact"
                :rules="localeRules"
                maxlength="2"
                dense
                @input="formatLocale"
              />
            </v-col>
            <v-col cols="9">
              <v-text-field
                dense
                v-model="newTranslation"
                :label="`Translation${newLocale ? ' (' + newLocale + ')' : ''}`"
                density="compact"
                @keyup.enter="addTranslation"
              />
            </v-col>
            <v-col cols="1" class="d-flex align-center">
              <v-btn
                color="primary"
                variant="outlined"
                size="small"
                @click="addTranslation"
                :disabled="
                  !newLocale || !newTranslation || newLocale.length !== 2
                "
              >
                Add
              </v-btn>
            </v-col>
          </v-row>

          <v-divider class="mb-4" />

          <div v-if="Object.keys(categoryData.i18n).length > 0">
            <h4 class="mb-3">Current Translations:</h4>
            <div
              v-for="(translation, locale) in categoryData.i18n"
              :key="locale"
              class="mb-2"
            >
              <v-row align="center">
                <v-col cols="1">
                  <v-chip color="primary" variant="outlined" size="small">
                    {{ locale.toUpperCase() }}
                  </v-chip>
                </v-col>
                <v-col cols="10">
                  <v-text-field
                    dense
                    v-model="categoryData.i18n[locale]"
                    :label="`Translation (${locale})`"
                    density="compact"
                    hide-details
                  />
                </v-col>
                <v-col cols="1">
                  <v-btn
                    size="small"
                    color="error"
                    variant="text"
                    @click="removeTranslation(locale)"
                    icon
                  >
                    <v-icon>mdi-delete</v-icon>
                  </v-btn>
                </v-col>
              </v-row>
            </div>
          </div>

          <div v-else class="text-center text-grey">
            <v-icon size="48" color="grey-lighten-2">mdi-translate-off</v-icon>
            <p class="mt-2">No translations added yet</p>
          </div>
        </v-card-text>
      </v-card>

      <v-row class="mt-6">
        <v-col cols="12" class="d-flex justify-end gap-3">
          <v-btn variant="outlined" @click="resetForm"> Reset </v-btn>

          <v-btn
            color="primary"
            :disabled="!formValid || !categoryData.name"
            :loading="isSaveLoading"
            @click="saveCategory"
          >
            <v-icon left>{{
              isEdit ? "mdi-pencil" : "mdi-content-save"
            }}</v-icon>
            {{ isEdit ? "Update Category" : "Create Category" }}
          </v-btn>

          <confirm-dialog :loading="isDeleteLoading" @confirm="deleteCategory">
            <v-btn
              v-if="isEdit"
              color="error"
              variant="outlined"
              :loading="isDeleteLoading"
            >
              <v-icon left>mdi-delete</v-icon>
              Delete
            </v-btn>
          </confirm-dialog>
        </v-col>
      </v-row>
    </v-form>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from "vue";
import { useStore } from "@/store";
import api from "@/api";
import { useRouter } from "vue-router/composables";
import ConfirmDialog from "../components/confirmDialog.vue";

const props = defineProps({
  category: {},
  isEdit: { type: Boolean, default: false },
});

const store = useStore();
const router = useRouter();

const SHOWCASE_CATEGORIES_SETTINGS_KEY = "showcase-categories";

const categoryData = ref({
  name: "",
  sorter: 0,
  type: "",
  i18n: {},
});

const formValid = ref(false);
const isSaveLoading = ref(false);
const isDeleteLoading = ref(false);
const categoryForm = ref(null);

const newLocale = ref("");
const newTranslation = ref("");

const categoryTypes = ["others", "hosting", "vds", "domains", "ai"];

const nameRules = [
  (v) => !!v || "Category name is required",
  (v) => v.length >= 2 || "Name must be at least 2 characters",
];

const localeRules = [
  (v) => !v || v.length === 2 || "Language code must be 2 characters",
  (v) =>
    !v ||
    /^[a-z]{2}$/.test(v) ||
    "Language code must contain only lowercase letters",
];

const existingCategories = computed(() => {
  const settings = store.getters["settings/all"].find(
    (v) => v.key === SHOWCASE_CATEGORIES_SETTINGS_KEY
  );

  try {
    return settings ? JSON.parse(settings.value) : [];
  } catch (e) {
    return [];
  }
});

const initializeForm = () => {
  if (props.isEdit && props.category) {
    categoryData.value = {
      name: props.category.name || "",
      sorter: props.category.sorter || 0,
      type: props.category.type || "",
      i18n: { ...props.category.i18n } || {},
    };
  } else {
    categoryData.value = {
      name: "",
      sorter: 0,
      type: "",
      i18n: {},
    };
  }
};

const formatLocale = (event) => {
  const value = event.target ? event.target.value : event;
  newLocale.value = value.toLowerCase().substring(0, 2);
};

const addTranslation = () => {
  if (
    !newLocale.value ||
    !newTranslation.value ||
    newLocale.value.length !== 2
  ) {
    store.commit("snackbar/showSnackbarError", {
      message: "Please enter a valid 2-character language code and translation",
    });
    return;
  }

  if (categoryData.value.i18n[newLocale.value]) {
    store.commit("snackbar/showSnackbarError", {
      message: `Translation for "${newLocale.value}" already exists. Use the existing field to edit it.`,
    });
    return;
  }

  categoryData.value.i18n[newLocale.value] = newTranslation.value;

  newLocale.value = "";
  newTranslation.value = "";
};

const removeTranslation = (locale) => {
  delete categoryData.value.i18n[locale];
  categoryData.value.i18n = { ...categoryData.value.i18n };
};

const resetForm = () => {
  initializeForm();
  newLocale.value = "";
  newTranslation.value = "";

  if (categoryForm.value) {
    categoryForm.value.resetValidation();
  }
};

const saveCategory = async () => {
  if (!categoryForm.value.validate()) {
    store.commit("snackbar/showSnackbarError", {
      message: "Please fill in all required fields correctly",
    });
    return;
  }

  if (!props.isEdit) {
    const existingCategory = existingCategories.value.find(
      (cat) => cat.name.toLowerCase() === categoryData.value.name.toLowerCase()
    );

    if (existingCategory) {
      store.commit("snackbar/showSnackbarError", {
        message: "Category with this name already exists",
      });
      return;
    }
  }

  isSaveLoading.value = true;

  try {
    const newCategoryData = {
      name: categoryData.value.name,
      sorter: categoryData.value.sorter || 0,
      type: categoryData.value.type || "others",
      i18n: { ...categoryData.value.i18n },
    };

    let updatedCategories;

    if (props.isEdit) {
      updatedCategories = existingCategories.value.map((cat) =>
        cat.name === props.category.name ? newCategoryData : cat
      );
    } else {
      updatedCategories = [...existingCategories.value, newCategoryData];
    }

    const settingsData = {
      key: SHOWCASE_CATEGORIES_SETTINGS_KEY,
      description: "Showcase categories",
      value: JSON.stringify(updatedCategories),
    };

    await api.settings.addKey(SHOWCASE_CATEGORIES_SETTINGS_KEY, settingsData);

    await store.dispatch("settings/fetch");

    store.commit("snackbar/showSnackbarSuccess", {
      message: props.isEdit
        ? "Category updated successfully!"
        : "Category created successfully!",
    });

    if (!props.isEdit) {
      router.push({ name: "Categories" });
    }
  } catch (error) {
    store.commit("snackbar/showSnackbarError", {
      message:
        error.response?.data?.message ||
        `Error ${props.isEdit ? "updating" : "creating"} category`,
    });
  } finally {
    isSaveLoading.value = false;
  }
};

const deleteCategory = async () => {
  if (!props.isEdit || !props.category) return;

  isDeleteLoading.value = true;

  try {
    const updatedCategories = existingCategories.value.filter(
      (cat) => cat.name !== props.category.name
    );

    const settingsData = {
      key: SHOWCASE_CATEGORIES_SETTINGS_KEY,
      description: "Showcase categories",
      value: JSON.stringify(updatedCategories),
    };

    await api.settings.addKey(SHOWCASE_CATEGORIES_SETTINGS_KEY, settingsData);

    await store.dispatch("settings/fetch");

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Category deleted successfully!",
    });

    router.push({ name: "Categories" });
  } catch (error) {
    store.commit("snackbar/showSnackbarError", {
      message: error.response?.data?.message || "Error deleting category",
    });
  } finally {
    isDeleteLoading.value = false;
  }
};

watch(
  () => props.category,
  () => {
    initializeForm();
  },
  { deep: true }
);

onMounted(async () => {
  try {
    await store.dispatch("settings/fetch");
    initializeForm();
  } catch (error) {
    store.commit("snackbar/showSnackbarError", {
      message: "Error loading categories data",
    });
  }
});
</script>

<style scoped lang="scss">
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 20px;
}

.gap-3 {
  gap: 12px;
}

.v-card {
  border-radius: 8px;
}

.bg-primary {
  background-color: rgb(var(--v-theme-primary)) !important;
  color: white;
}
</style>
