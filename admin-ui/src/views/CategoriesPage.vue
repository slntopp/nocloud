<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Categories' }">{{
        navTitle("Categories")
      }}</router-link>
      / {{ categorieTitle }}
    </h1>
    <v-tabs
      class="rounded-t-lg"
      background-color="background-light"
      v-model="tabs"
    >
      <v-tab>Info</v-tab>
      <v-tab>Showcases</v-tab>
    </v-tabs>
    <v-tabs-items
      class="rounded-b-lg"
      style="background: var(--v-background-light-base)"
      v-model="tabs"
    >
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="isFetchLoading" />
        <categories-create
          v-if="categorie && categorieTitle"
          :category="categorie"
          is-edit
        />
      </v-tab-item>

      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="isFetchLoading" />
        <categories-showcases :category="categorie" />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script setup>
import config from "@/config.js";
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router/composables";
import { useStore } from "@/store";
import CategoriesCreate from "@/views/CategoriesCreate.vue";
import CategoriesShowcases from "../components/categories/showcases.vue";

const SHOWCASE_CATEGORIES_SETTINGS_KEY = "showcase-categories";

const route = useRoute();
const store = useStore();

const navTitles = ref(config.navTitles ?? {});
const isFetchLoading = ref(false);
const categories = ref([]);
const categorie = ref(null);
const tabs = ref(0);

onMounted(() => {
  setCategoriesFromSettings();
});

const originalSettings = computed(() =>
  store.getters["settings/all"].find(
    (v) => v.key === SHOWCASE_CATEGORIES_SETTINGS_KEY
  )
);

const categorieId = computed(() => {
  return route.params.uuid;
});

const categorieTitle = computed(() => {
  if (!categorie.value || !categorie.value?.name) {
    return "...";
  }

  return categorie.value.name;
});

const setCategoriesFromSettings = () => {
  try {
    const settings =
      originalSettings.value && JSON.parse(originalSettings.value.value);

    if (Array.isArray(settings)) {
      categories.value = settings;
    } else {
      categories.value = [];
    }

    categorie.value = categories.value.find(
      (c) => c.name === categorieId.value
    );
  } catch (e) {
    categories.value = [];
  }
};

const navTitle = (title) => {
  if (title && navTitles.value[title]) {
    return navTitles.value[title];
  }

  return title;
};

watch(originalSettings, () => {
  setCategoriesFromSettings();
});

watch(categorie, (newVal) => {
  if (newVal && newVal.title) {
    document.title = `${newVal.title} | NoCloud`;
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
  margin-bottom: 10px;
}
</style>
