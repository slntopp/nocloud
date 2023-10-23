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
      <v-tab v-for="tab in tabItems" :key="tab.title">{{ tab.title }}</v-tab>
    </v-tabs>
    <v-tabs-items
      class="rounded-b-lg"
      style="background: var(--v-background-light-base)"
      v-model="selectedTab"
    >
      <v-tab-item v-for="(tab, index) in tabItems" :key="tab.title">
        <v-progress-linear indeterminate class="pt-2" v-if="accountLoading" />
        <component
          v-if="account && index === selectedTab"
          :is="tab.component"
          :account="account"
        />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script>
import config from "@/config.js";
import AccountsInfo from "@/components/account/info.vue";
import AccountsTemplate from "@/components/account/template.vue";
import AccountsEvents from "@/components/account/events.vue";
import AccountsHistory from "@/components/account/history.vue";
import AccountReport from "@/components/account/reports.vue";

export default {
  name: "account-view",
  components: {
    AccountReport,
    AccountsInfo,
    AccountsTemplate,
    AccountsEvents,
    AccountsHistory,
  },
  data: () => ({ selectedTab: 0, navTitles: config.navTitles ?? {} }),
  methods: {
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
  },
  computed: {
    account() {
      const id = this.$route.params?.accountId;

      return this.$store.getters["accounts/all"].find(
        ({ uuid }) => uuid === id
      );
    },
    accountTitle() {
      return this?.account?.title ?? "not found";
    },
    accountLoading() {
      return this.$store.getters["accounts/isLoading"];
    },
    accountId() {
      return this.$route.params.accountId;
    },
    tabItems() {
      return [
        {
          component: AccountsInfo,
          title: "info",
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
          component: AccountsTemplate,
          title: "template",
        },
      ];
    },
  },
  created() {
    this.$store.dispatch("accounts/fetchById", this.accountId).then(() => {
      document.title = `${this.accountTitle} | NoCloud`;
    });
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "accounts/fetchById",
      params: this.accountId,
    });
    this.selectedTab = this.$route.query.tab || 0;
  },
};
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
