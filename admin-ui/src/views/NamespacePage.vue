<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Namespaces' }">{{
        navTitle("Namespaces")
      }}</router-link>
      / {{ namespaceTitle }}
    </h1>
    <v-tabs
      class="rounded-t-lg"
      bg-color="background-light"
      v-model="tabs"
    >
      <v-tab>Info</v-tab>
      <v-tab>Accounts</v-tab>
    </v-tabs>
    <v-window
      class="rounded-b-lg"
      style="background: rgb(var(--v-theme-background-light))"
      v-model="tabs"
    >
      <v-window-item>
        <v-progress-linear indeterminate class="pt-2" v-if="isLoading" />
        <namespace-info
          :loading="isLoading"
          v-if="namespace && namespaceTitle"
          :namespace="namespace"
          @input:title="namespace.title = $event"
        />
      </v-window-item>
      <v-window-item>
        <v-progress-linear indeterminate class="pt-2" v-if="isLoading" />
        <namespace-accounts
          v-if="namespace && namespaceTitle"
          :namespace="namespace"
        />
      </v-window-item>
    </v-window>
  </div>
</template>

<script>
import config from "@/config.js";
import namespaceInfo from "../components/namespace/info.vue";
import NamespaceAccounts from "@/components/namespace/accounts.vue";

export default {
  name: "account-view",
  data: () => ({
    navTitles: config.navTitles ?? {},
    tabs: 0,
  }),
  components: { NamespaceAccounts, namespaceInfo },
  methods: {
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
  },
  computed: {
    namespace() {
      return this.$store.getters["namespaces/one"] || {};
    },
    isLoading() {
      return this.$store.getters["namespaces/isLoading"];
    },
    namespaceId() {
      return this.$route.params.namespaceId;
    },
    namespaceTitle() {
      if (this.isLoading) {
        return "...";
      } else if (!this.isLoading && !Object.keys(this.namespace).length) {
        return "Not found";
      }

      return this.namespace.title;
    },
  },
  async mounted() {
    await this.$store.dispatch("namespaces/fetchById", this.namespaceId);
  },
  watch: {
    namespaceTitle(value) {
      document.title = `${value} | NoCloud`;
    },
  },
};
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
