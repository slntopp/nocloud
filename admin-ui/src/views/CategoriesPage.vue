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

const route = useRoute();
const store = useStore();

const navTitles = ref(config.navTitles ?? {});
const categorie = ref(null);
const tabs = ref(0);

onMounted(() => {
  store.dispatch("showcases/fetch");
});

const isFetchLoading = computed(() => store.getters["showcases/isLoading"]);
const categories = computed(() => store.getters["showcases/categories"]);

const categorieId = computed(() => {
  return route.params.uuid;
});

const categorieTitle = computed(() => {
  if (!categorie.value || !categorie.value?.title) {
    return "...";
  }

  return categorie.value.title;
});

const navTitle = (title) => {
  if (title && navTitles.value[title]) {
    return navTitles.value[title];
  }

  return title;
};

watch(categorie, (newVal) => {
  if (newVal && newVal.title) {
    document.title = `${newVal.title} | NoCloud`;
  }
});

watch(categories, () => {
  categorie.value = categories.value.find(
    (cat) => cat.uuid === categorieId.value
  );
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
