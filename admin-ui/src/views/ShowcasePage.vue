<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Showcases' }">{{
        navTitle("Showcases")
      }}</router-link>
      / {{ showcaseTitle }}
    </h1>
    <v-tabs
      class="rounded-t-lg"
      background-color="background-light"
      v-model="tabs"
    >
      <v-tab>Info</v-tab>
      <v-tab>Promo</v-tab>
      <v-tab>Promocodes</v-tab>
      <v-tab>Template</v-tab>
    </v-tabs>
    <v-tabs-items
      class="rounded-b-lg"
      style="background: var(--v-background-light-base)"
      v-model="tabs"
    >
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="isFetchLoading" />
        <showcase-create
          v-if="showcase && showcaseTitle"
          :real-showcase="showcase"
          is-edit
        />
      </v-tab-item>
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="isFetchLoading" />
        <promo-tab :template="showcase" />
      </v-tab-item>
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="isFetchLoading" />
        <promocodes-tab :template="showcase" />
      </v-tab-item>
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="isFetchLoading" />
        <showcase-template :template="showcase" />
        />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script setup>
import config from "@/config.js";
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router/composables";
import { useStore } from "@/store";
import ShowcaseCreate from "@/views/ShowcaseCreate.vue";
import PromoTab from "@/components/showcase/promo.vue";
import PromocodesTab from "@/components/showcase/promocodes.vue";
import ShowcaseTemplate from "@/components/showcase/template.vue";

const route = useRoute();
const store = useStore();

const navTitles = ref(config.navTitles ?? {});
const isFetchLoading = ref(false);
const tabs = ref(0);
const showcase = ref({});

const navTitle = (title) => {
  if (title && navTitles.value[title]) {
    return navTitles.value[title];
  }

  return title;
};
const showcases = computed(() => store.getters["showcases/all"]);
const showcaseId = computed(() => {
  return route.params.uuid;
});
const showcaseTitle = computed(() => {
  if (!showcase.value || !showcase.value?.title) {
    return "...";
  }

  return showcase.value.title;
});

onMounted(async () => {
  try {
    isFetchLoading.value = true;
    await store.dispatch("showcases/fetchById", showcaseId.value);
  } catch (e) {
    console.log(e);
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during fetch showcase",
    });
  } finally {
    isFetchLoading.value = false;
  }

  showcase.value = showcases.value?.find((n) => n.uuid == showcaseId.value);
  document.title = `${showcase.value.title} | NoCloud`;
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
