<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Invoices' }">{{
        navTitle("Invoices")
      }}</router-link>
      / {{ invoiceTitle || "Not found" }}
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
        <v-progress-linear indeterminate class="pt-2" v-if="isInvoiceLoading" />
        <component
          v-if="invoice && index === selectedTab"
          :is="tab.component"
          :invoice="invoice"
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
import InvoiceCreate from "@/views/InvoiceCreate.vue";
import InvoiceTemplate from "@/components/invoice/template.vue";

const store = useStore();
const route = useRoute();

const selectedTab = ref(0);
const navTitles = ref(config.navTitles ?? {});

const invoice = computed(() => store.getters["invoices/one"]);
const invoiceTitle = computed(() =>
  isInvoiceLoading.value ? "..." : "#" + invoice.value?.uuid
);

onMounted(async () => {
  store.commit("reloadBtn/setCallback", {
    type: "invoices/fetchById",
    params: route.params?.uuid,
  });
  selectedTab.value = route.query.tab || 0;

  try {
    await store.dispatch("invoices/fetchById", route.params.uuid);
    document.title = `${invoiceTitle.value} | NoCloud`;
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  }
});

const isInvoiceLoading = computed(() => {
  return store.getters["invoices/isLoading"];
});

const tabItems = computed(() => [
  {
    component: InvoiceCreate,
    title: "info",
  },
  {
    component: InvoiceTemplate,
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
export default { name: "invoice-view" };
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
