<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Accounts' }">{{
        navTitle("Accounts")
      }}</router-link>
      / {{ accountTitle }}
    </h1>
    <v-tabs
      class="rounded-t-lg"
      background-color="background-light"
      v-model="selectedTab"
    >
      <v-tab v-for="tab in tabItems" :key="tab.title"
        >{{ tab.title }}
        <v-chip class="ml-1" small v-if="tab.title === 'notes'">{{
          account?.adminNotes?.length || 0
        }}</v-chip>
      </v-tab>
    </v-tabs>
    <v-tabs-items
      class="rounded-b-lg"
      style="background: var(--v-background-light-base)"
      v-model="selectedTab"
    >
      <v-tab-item v-for="(tab, index) in tabItems" :key="tab.title">
        <v-progress-linear indeterminate class="pt-2" v-if="accountLoading" />
        <component
          v-else-if="account && index === selectedTab && !isAccountNotFound"
          :is="tab.component"
          :account="account"
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

import AccountsInfo from "@/components/account/info.vue";
import AccountsTemplate from "@/components/account/template.vue";
import AccountsEvents from "@/components/account/events.vue";
import AccountsHistory from "@/components/account/history.vue";
import AccountReport from "@/components/account/reports.vue";
import AccountChats from "@/components/account/chats.vue";
import AccountNotes from "@/components/account/notes.vue";

const store = useStore();
const route = useRoute();

const selectedTab = ref(0);
const isAccountNotFound = ref(false);
const navTitles = ref(config.navTitles ?? {});

function navTitle(title) {
  if (title && navTitles.value[title]) {
    return navTitles.value[title];
  }

  return title;
}

const account = computed(() => store.getters["accounts/one"]);

const accountTitle = computed(() => {
  return !isAccountNotFound.value ? account.value?.title : "not found";
});

const accountLoading = computed(() => {
  return store.getters["accounts/isLoading"];
});

const tabItems = computed(() => [
  {
    component: AccountsInfo,
    title: "info",
  },
  {
    component: AccountNotes,
    title: "notes",
  },
  {
    component: AccountsEvents,
    title: "events",
  },
  {
    component: AccountsHistory,
    title: "event log",
  },
  {
    component: AccountReport,
    title: "reports",
  },
  {
    component: AccountChats,
    title: "helpdesk",
  },
  {
    component: AccountsTemplate,
    title: "template",
  },
]);

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    event: fetchAccount,
  });
  selectedTab.value = route.query.tab || 0;

  fetchAccount();
});

const fetchAccount = async () => {
  try {
    await store.dispatch("accounts/fetchById", route.params?.accountId);
  } catch {
    isAccountNotFound.value = true;
  } finally {
    document.title = `${accountTitle.value} | NoCloud`;
  }
};
</script>

<script>
export default { name: "account-view" };
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
