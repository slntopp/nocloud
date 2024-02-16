<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Accounts' }">{{
        navTitle("Accounts")
      }}</router-link>
      / {{ groupTitle }}
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
        <v-progress-linear indeterminate class="pt-2" v-if="addonsLoading" />
        <component
          v-if="addonGroup && index === selectedTab"
          :is="tab.component"
          :addon-group="addonGroup"
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
import { getFullDate } from "@/functions";

import AddonCreate from "@/views/AddonCreate.vue";
import AddonProducts from "@/components/addons/products.vue";

const store = useStore();
const route = useRoute();

const selectedTab = ref(0);
const navTitles = ref(config.navTitles ?? {});

function navTitle(title) {
  if (title && navTitles.value[title]) {
    return navTitles.value[title];
  }

  return title;
}

const addonGroup = computed(() => {
  const addons = store.getters["addons/all"]
    .filter(({ group }) => group === groupTitle.value)?.map((a) => ({ ...a, period: getFullDate(a.period) }));

  return { addons, title: groupTitle.value };
});

const groupTitle = computed(() => route.params.title);

const addonsLoading = computed(() => {
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
]);

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    type: "addons/fetch",
  });
  selectedTab.value = route.query.tab || 0;
});

store.dispatch("addons/fetch").then(() => {
  document.title = `${groupTitle.value} | NoCloud`;
});
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
