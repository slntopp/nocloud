<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Addons' }">{{
        navTitle("Addons")
      }}</router-link>
      / {{ addonTitle || "Not found" }}
    </h1>
    <v-tabs
      class="rounded-t-lg"
      background-color="background-light"
      v-model="selectedTab"
    >
      <v-tab v-for="tab in tabItems" :key="tab.title">{{ tab.title }} </v-tab>
    </v-tabs>
    <v-tabs-items
      class="rounded-b-lg"
      style="background: var(--v-background-light-base)"
      v-model="selectedTab"
    >
      <v-tab-item v-for="(tab, index) in tabItems" :key="tab.title">
        <v-progress-linear indeterminate class="pt-2" v-if="isAddonLoading" />
        <component
          v-if="addon && index === selectedTab"
          :is="tab.component"
          :addon="addon"
          is-edit
        />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router/composables";
import { useStore } from "@/store";
import config from "@/config.js";

import AddonCreate from "@/views/AddonCreate.vue";
import AddonProducts from "@/components/addons/products.vue";
import AddonTemplate from "@/components/addons/template.vue";

const store = useStore();
const route = useRoute();

const selectedTab = ref(0);
const navTitles = ref(config.navTitles ?? {});

const addon = computed(() => store.getters["addons/one"]);
const addonTitle = computed(() => addon.value?.title);

onMounted(async () => {
  store.commit("reloadBtn/setCallback", {
    type: "addons/fetchById",
    params: route.params?.uuid,
  });
  selectedTab.value = route.query.tab || 0;

  try {
    await store.dispatch("addons/fetchById", route.params.uuid);
    document.title = `${addonTitle.value} | NoCloud`;

    const desc = await store.dispatch(
      "descriptions/get",
      addon.value.descriptionId
    );
    store.commit("addons/setOne", { ...addon.value, description: desc.text });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  }
});

const isAddonLoading = computed(() => {
  return store.getters["addons/isLoading"];
});

const tabItems = computed(() => [
  {
    component: AddonCreate,
    title: "info",
  },
  {
    component: AddonProducts,
    title: "products",
  },
  {
    component: AddonTemplate,
    title: "template",
  },
]);

function navTitle(title) {
  if (title && navTitles.value[title]) {
    return navTitles.value[title];
  }

  return title;
}
</script>

<script>
export default { name: "addon-view" };
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
