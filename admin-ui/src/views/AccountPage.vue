<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Accounts' }">{{
        navTitle("Accounts")
      }}</router-link>
      / {{ accountTitle }}
      <account-grade :account="account" />
    </h1>
    <v-tabs
      show-arrows
      class="rounded-t-lg"
      bg-color="background-light"
      v-model="selectedTab"
    >
      <v-tab v-for="tab in tabItems" :key="tab.title"
        >{{ tab.title }}
        <v-chip class="ml-1" size="small" v-if="tab.title === 'notes'">{{
          account?.adminNotes?.length || 0
        }}</v-chip>
      </v-tab>
    </v-tabs>
    <v-window
      class="rounded-b-lg"
      style="background: rgb(var(--v-theme-background-light))"
      v-model="selectedTab"
    >
      <v-window-item v-for="(tab, index) in tabItems" :key="tab.title">
        <v-progress-linear indeterminate class="pt-2" v-if="accountLoading" />
        <component
          v-else-if="account && index === selectedTab && !isAccountNotFound"
          :is="tab.component"
          :account="account"
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

import AccountGrade from "@/components/ui/accountGrade.vue";
import AccountsInfo from "@/components/account/info.vue";
import AccountsTemplate from "@/components/account/template.vue";
import AccountsEvents from "@/components/account/events.vue";
import AccountsHistory from "@/components/account/history.vue";
import AccountReport from "@/components/account/reports.vue";
import AccountChats from "@/components/account/chats.vue";
import AccountCampaigns from "@/components/account/campaigns.vue";
import AccountNotes from "@/components/account/notes.vue";
import AccountSubaccounts from "@/components/account/subaccounts.vue";
import AccountInvoices from "@/components/account/invoices.vue";

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

const hasCampaigns = computed(() =>
  (store.getters["plugins/all"] || []).some((p) =>
    (p.url || "").includes("campaigns.ui")
  )
);

const tabItems = computed(() => [
  {
    component: AccountsInfo,
    title: "info",
  },
  {
    component: AccountInvoices,
    title: "invoices",
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
    component: AccountSubaccounts,
    title: "subaccounts",
  },
  {
    component: AccountChats,
    title: "helpdesk",
  },
  // Only when the campaigns plugin is installed: other installations of
  // nocloud do not have it, and an empty tab is worse than no tab.
  ...(hasCampaigns.value
    ? [
        {
          component: AccountCampaigns,
          title: "campaigns",
        },
      ]
    : []),
  {
    component: AccountsTemplate,
    title: "template",
  },
]);

onMounted(() => {
  store.commit("reloadBtn/setCallback", {
    event: fetchAccount,
  });
  if (route.query.tab) {
    selectedTab.value =
      +route.query.tab ||
      tabItems.value.findIndex((tab) => tab.title === route.query.tab) ||
      0;
  }

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
  color: rgb(var(--v-theme-primary));
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>
