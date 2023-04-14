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
      v-model="tabs"
    >
      <v-tab>Info</v-tab>
      <v-tab>Template</v-tab>
      <v-tab>Events</v-tab>
      <v-tab>History</v-tab>
    </v-tabs>
    <v-tabs-items
      class="rounded-b-lg"
      style="background: var(--v-background-light-base)"
      v-model="tabs"
    >
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="accountLoading" />
        <accounts-info v-if="account" :account="account" />
      </v-tab-item>
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="accountLoading" />
        <accounts-template v-if="account" :template="account" />
      </v-tab-item>
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="accountLoading" />
        <accounts-events
          :is-loading="accountLoading"
          v-if="account"
          :account="account"
        />
      </v-tab-item>
      <v-tab-item>
        <v-progress-linear indeterminate class="pt-2" v-if="accountLoading" />
        <accounts-history
          :is-loading="accountLoading"
          v-if="account"
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

export default {
  name: "account-view",
  components: {
    AccountsInfo,
    AccountsTemplate,
    AccountsEvents,
    AccountsHistory,
  },
  data: () => ({ tabs: 0, navTitles: config.navTitles ?? {} }),
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
    this.tabs = this.$route.query.tab || 0;
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
