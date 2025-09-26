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

    <categories-create
      v-if="selectedCategory"
      :category="categoryForm"
      is-edit
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, toRefs, onMounted } from "vue";
import { useStore } from "../../store";
import api from "../../api";
import CategoriesCreate from "../../views/CategoriesCreate.vue";

const SHOWCASE_CATEGORIES_SETTINGS_KEY = "showcase-categories";

const props = defineProps({ template: { type: Object, required: true } });
const { template } = toRefs(props);

const store = useStore();

const categories = ref([]);

const selectedCategory = ref(null);

const categoryForm = ref({
  name: "",
  sorter: 0,
  type: "",
  i18n: {},
});

onMounted(() => {
  setCategoriesFromSettings();
  setCategoryFromTemplate();
});

const originalSettings = computed(() =>
  store.getters["settings/all"].find(
    (v) => v.key === SHOWCASE_CATEGORIES_SETTINGS_KEY
  )
);

const onCategorySelect = (categoryName) => {
  if (!categoryName) {
    return;
  }

  const category = categories.value.find((c) => c.name === categoryName);
  if (category) {
    categoryForm.value = JSON.parse(JSON.stringify(category));
  }
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
