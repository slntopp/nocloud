<template>
  <div>
    <!-- Category Selection -->
    <v-card color="background-light" class="mb-4">
      <v-card-title>
        <v-icon left>mdi-folder</v-icon>
        Category Management
      </v-card-title>
      <v-card-text>
        <v-select
          v-model="selectedCategory"
          :items="categories"
          item-text="name"
          item-value="name"
          label="Select Category"
          clearable
          prepend-icon="mdi-magnify"
          @change="onCategorySelect"
        />
      </v-card-text>
    </v-card>

    <!-- Category Form - Always visible -->
    <v-card color="background-light">
      <v-card-title>
        <v-icon left>{{ isEditing ? "mdi-pencil" : "mdi-plus" }}</v-icon>
        {{
          isEditing ? `Edit Category ${selectedCategory}` : "Add New Category"
        }}
      </v-card-title>
      <v-card-text>
        <v-form ref="form" v-model="formValid">
          <!-- Category Name -->
          <v-row>
            <v-col cols="8">
              <v-text-field
                v-model="categoryForm.name"
                label="Category Name"
                :rules="nameRules"
                prepend-icon="mdi-tag"
                required
              />
            </v-col>
            <v-col cols="2">
              <v-select
                v-model="categoryForm.type"
                label="Category Type"
                :items="['others', 'hosting', 'vds', 'domains','ai']"
              />
            </v-col>

            <v-col cols="2">
              <v-text-field
                type="number"
                min="0"
                v-model="categoryForm.sorter"
                label="Category Sorter"
              />
            </v-col>
          </v-row>
          <!-- I18n Translations -->
          <v-card color="background-light" class="mt-4">
            <v-card-title class="text-h6">
              <v-icon left>mdi-translate</v-icon>
              Translations
            </v-card-title>
            <v-card-text>
              <!-- Existing translations -->
              <div
                v-for="(translation, locale) in categoryForm.i18n"
                :key="locale"
                class="mb-3"
              >
                <v-row align="center">
                  <v-col cols="1">
                    <v-chip>
                      {{ locale.toUpperCase() }}
                    </v-chip>
                  </v-col>
                  <v-col cols="10">
                    <v-text-field
                      v-model="categoryForm.i18n[locale]"
                      :label="`Translation (${locale})`"
                      dense
                      hide-details
                    />
                  </v-col>
                  <v-col cols="1">
                    <v-btn
                      size="small"
                      color="error"
                      icon
                      @click="removeTranslation(locale)"
                    >
                      <v-icon>mdi-delete</v-icon>
                    </v-btn>
                  </v-col>
                </v-row>
              </div>

              <!-- Add new translation - Always visible -->
              <v-divider class="mb-4" />
              <v-row align="center">
                <v-col cols="2">
                  <v-text-field
                    v-model.trim="newLocale"
                    label="Locale"
                    placeholder="e.g. en, ru, de"
                    dense
                    hide-details
                  />
                </v-col>
                <v-col cols="9">
                  <v-text-field
                    v-model="newTranslation"
                    label="Translation"
                    placeholder="Enter translation"
                    @keyup.enter="addTranslation"
                    dense
                    hide-details
                  />
                </v-col>
                <v-col cols="1">
                  <v-btn
                    icon
                    size="small"
                    color="primary"
                    :disabled="!newLocale || !newTranslation"
                    @click="addTranslation"
                  >
                    <v-icon>mdi-plus</v-icon>
                  </v-btn>
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>
        </v-form>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn
          color="primary"
          :disabled="!formValid"
          :loading="loading"
          @click="saveCategory"
        >
          {{ isEditing ? "Update" : "Create" }}
        </v-btn>
        <v-btn
          v-if="isEditing"
          color="error"
          :loading="deleting"
          @click="confirmDialog = true"
        >
          Delete
        </v-btn>
      </v-card-actions>
    </v-card>

    <!-- Confirmation Dialog -->
    <v-dialog v-model="confirmDialog" max-width="400">
      <v-card color="background-light">
        <v-card-title>Confirm Deletion</v-card-title>
        <v-card-text>
          Are you sure you want to delete the category "{{
            categoryForm.name
          }}"?
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn color="grey" @click="confirmDialog = false"> Cancel </v-btn>
          <v-btn color="error" @click="deleteCategory"> Delete </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, toRefs, onMounted } from "vue";
import { useStore } from "../../store";
import api from "../../api";

const SHOWCASE_CATEGORIES_SETTINGS_KEY = "showcase-categories";

const props = defineProps({ template: { type: Object, required: true } });
const { template } = toRefs(props);

const store = useStore();

const categories = ref([]);

const selectedCategory = ref(null);
const formValid = ref(false);
const loading = ref(false);
const deleting = ref(false);
const confirmDialog = ref(false);
const form = ref(null);

const categoryForm = ref({
  name: "",
  sorter: 0,
  type: "",
  i18n: {},
});

const newLocale = ref("");
const newTranslation = ref("");

const nameRules = ref([
  (v) => !!v || "Category name is required",
  (v) => v.length >= 2 || "Name must be at least 2 characters",
]);

onMounted(() => {
  setCategoriesFromSettings();
  setCategoryFromTemplate();
});

const originalSettings = computed(() =>
  store.getters["settings/all"].find(
    (v) => v.key === SHOWCASE_CATEGORIES_SETTINGS_KEY
  )
);

const isEditing = computed(() => !!selectedCategory.value);

const onCategorySelect = (categoryName) => {
  if (!categoryName) {
    clearForm();
    return;
  }

  const category = categories.value.find((c) => c.name === categoryName);
  if (category) {
    categoryForm.value.name = category.name;
    categoryForm.value.i18n = { ...category.i18n } || {};
    newLocale.value = "";
    newTranslation.value = "";
  }
};

const addTranslation = () => {
  if (newLocale.value && newTranslation.value) {
    if (categoryForm.value.i18n[newLocale.value]) {
      categoryForm.value.i18n[newLocale.value] = newTranslation.value;
    } else {
      categoryForm.value.i18n[newLocale.value] = newTranslation.value;
    }
    newLocale.value = "";
    newTranslation.value = "";
  }
};

const removeTranslation = (locale) => {
  delete categoryForm.value.i18n[locale];
};

const clearForm = () => {
  categoryForm.value.name = "";
  categoryForm.value.sorter = 0;
  categoryForm.value.type = "";
  categoryForm.value.i18n = {};
  newLocale.value = "";
  newTranslation.value = "";
  selectedCategory.value = null;
  if (form.value) {
    form.value.resetValidation();
  }
};

const saveCategory = async () => {
  if (!form.value.validate()) return;

  loading.value = true;

  try {
    const categoryData = {
      name: categoryForm.value.name,
      sorter: categoryForm.value.sorter || 0,
      type: categoryForm.value.type || "",
      i18n: { ...categoryForm.value.i18n },
    };

    if (isEditing.value) {
      const index = categories.value.findIndex(
        (c) => c.name === selectedCategory.value
      );
      if (index !== -1) {
        categories.value[index] = categoryData;
        selectedCategory.value = categoryData.name;
      }
    } else {
      const existingIndex = categories.value.findIndex(
        (c) => c.name === categoryData.name
      );
      if (existingIndex !== -1) {
        categories.value[existingIndex] = categoryData;
      } else {
        categories.value.push(categoryData);
      }
    }

    await updateSettingsAndShowcase(categories.value, categoryData);

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Category saved successfully",
    });
  } catch (error) {
    store.commit("snackbar/showSnackbarError", {
      message: "Error on save categories",
    });
  } finally {
    loading.value = false;
  }
};

const updateSettingsAndShowcase = async (categories, newCategory) => {
  let data =
    originalSettings.value &&
    JSON.parse(JSON.stringify(originalSettings.value));
  if (!data) {
    data = {
      key: SHOWCASE_CATEGORIES_SETTINGS_KEY,
      description: "Showcase categories",
    };
  }

  data.value = JSON.stringify(categories);

  await api.settings.addKey(SHOWCASE_CATEGORIES_SETTINGS_KEY, data);
  await updateShowcaseCategory(newCategory);
  await store.dispatch("settings/fetch");
};

const updateShowcaseCategory = async (newCategory) => {
  if (
    JSON.stringify(newCategory) ===
    JSON.stringify(template.value?.meta?.category)
  ) {
    return;
  }

  const newShowcase = {
    ...template.value,
    meta: { ...template.value.meta, category: newCategory },
  };
  await api.showcases.update(newShowcase);
  store.commit("showcases/replaceShowcase", newShowcase);
};

const deleteCategory = async () => {
  deleting.value = true;

  try {
    categories.value = categories.value.filter(
      (c) => c.name !== selectedCategory.value
    );

    await updateSettingsAndShowcase(categories.value, undefined);

    confirmDialog.value = false;
    clearForm();

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Category deleted successfully",
    });
  } catch (error) {
    store.commit("snackbar/showSnackbarError", {
      message: "Error on delete category",
    });
  } finally {
    deleting.value = false;
  }
};

const setCategoriesFromSettings = () => {
  try {
    const settings =
      originalSettings.value && JSON.parse(originalSettings.value.value);

    if (Array.isArray(settings)) {
      categories.value = settings;
    } else {
      categories.value = [];
    }
  } catch (e) {
    categories.value = [];
  }
};

const setCategoryFromTemplate = () => {
  if (template.value && template.value.meta && template.value.meta.category) {
    const cat = template.value.meta.category;
    selectedCategory.value = cat.name;
    categoryForm.value.name = cat.name;
    categoryForm.value.sorter = cat.sorter || 0;
    categoryForm.value.type = cat.type || "";
    categoryForm.value.i18n = { ...cat.i18n } || {};
  }
};

watch(
  () => categoryForm.value.name,
  () => {
    if (form.value) {
      form.value.validate();
    }
  }
);

watch(
  template,
  () => {
    setCategoryFromTemplate();
  },
  { deep: true }
);

watch(originalSettings, () => {
  setCategoriesFromSettings();
});

watch(selectedCategory, (newVal, prevVal) => {
  if ((!prevVal && !newVal) || categories.value.length === 0) return;

  updateShowcaseCategory(
    newVal ? categories.value.find((c) => c.name === newVal) : undefined
  );
});
</script>

<script>
export default {
  name: "ShowcaseCategory",
};
</script>

<style scoped>
.v-card {
  margin-bottom: 16px;
}
</style>
