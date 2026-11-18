<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Promocodes' }">{{
        navTitle("Promocodes")
      }}</router-link>
      / {{ promocodeTitle || "Not found" }}
    </h1>
    <v-tabs
      class="rounded-t-lg"
      bg-color="background-light"
      v-model="selectedTab"
    >
      <v-tab v-for="tab in tabItems" :key="tab.title">{{ tab.title }} </v-tab>
    </v-tabs>
    <v-window
      class="rounded-b-lg"
      style="background: rgb(var(--v-theme-background-light))"
      v-model="selectedTab"
    >
      <v-window-item v-for="(tab, index) in tabItems" :key="tab.title">
        <v-progress-linear
          indeterminate
          class="pt-2"
          v-if="isPromocodeLoading"
        />
        <component
          v-else-if="promocode && index === selectedTab"
          :is="tab.component"
          :promocode="promocode"
          is-edit
          @refresh="refreshPromocode"
        />
      </v-window-item>
    </v-window>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { useStore } from "@/store";
import config from "@/config.js";
import PromocodeCreate from "./PromocodeCreate.vue";
import PromocodeActivations from "@/components/promocode/activation.vue";

const store = useStore();
const route = useRoute();

const selectedTab = ref(0);
const navTitles = ref(config.navTitles ?? {});

const promocode = computed(() => store.getters["promocodes/one"]);
const promocodeTitle = computed(() =>
  isPromocodeLoading.value ? "..." : promocode.value?.title
);

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    type: "promocodes/get",
    params: route.params?.uuid,
  });
  selectedTab.value = route.query.tab || 0;

  refreshPromocode();
});

const isPromocodeLoading = computed(() => {
  return store.getters["promocodes/isLoading"];
});

const tabItems = computed(() => [
  {
    component: PromocodeCreate,
    title: "info",
  },
  {
    component: PromocodeActivations,
    title: "activations",
  },
]);

function navTitle(title) {
  if (title && navTitles.value[title]) {
    return navTitles.value[title];
  }

  return title;
}

const refreshPromocode = async () => {
  try {
    await store.dispatch("promocodes/get", route.params.uuid);
    document.title = `${promocodeTitle.value} | NoCloud`;
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  }
};
</script>

<script>
export default { name: "promocode-view" };
</script>

<style scoped lang="scss">
.page__title {
  color: rgb(var(--v-theme-primary));
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>
