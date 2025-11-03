<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'AccountGroups' }">{{
        navTitle("Account Groups")
      }}</router-link>
      / {{ accountGroupTitle || "Not found" }}
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
        <v-progress-linear
          indeterminate
          class="pt-2"
          v-if="isAccountGroupLoading"
        />
        <component
          v-if="accountGroup && index === selectedTab"
          :is="tab.component"
          :account-group="accountGroup"
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

import AccountGroupCreate from "@/views/AccountGroupCreate.vue";
import AccountGroupAccounts from "@/components/account-groups/accounts.vue";

const store = useStore();
const route = useRoute();

const selectedTab = ref(0);
const navTitles = ref(config.navTitles ?? {});

const accountGroup = computed(() => store.getters["accountGroups/one"]);
const accountGroupTitle = computed(() => accountGroup.value?.title);

onMounted(async () => {
  store.commit("reloadBtn/setCallback", {
    type: "accountGroups/fetchById",
    params: route.params?.uuid,
  });
  selectedTab.value = route.query.tab || 0;

  try {
    await store.dispatch("accountGroups/fetchById", route.params.uuid);
    document.title = `${accountGroupTitle.value} | NoCloud`;
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: "Failed to load Account Group",
    });
  }
});

const isAccountGroupLoading = computed(() => {
  return store.getters["accountGroups/isLoading"];
});

const tabItems = computed(() => [
  {
    component: AccountGroupCreate,
    title: "info",
  },
  {
    component: AccountGroupAccounts,
    title: "accounts",
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
export default { name: "account-group-page" };
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
