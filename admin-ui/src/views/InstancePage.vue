<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Instances' }">{{
        navTitle("Instances")
      }}</router-link>
      / {{ instanceTitle }}
    </h1>
    <v-tabs
      class="rounded-t-lg"
      background-color="background-light"
      v-model="tabsIndex"
    >
      <v-tab v-for="tab of tabs" :key="tab.title">{{ tab.title }}</v-tab>
    </v-tabs>
    <v-tabs-items
      class="rounded-b-lg"
      style="background: var(--v-background-light-base)"
      v-model="tabsIndex"
    >
      <v-tab-item v-for="tab of tabs" :key="tab.title">
        <v-progress-linear indeterminate class="pt-2" v-if="instanceLoading" />
        <component
          v-else-if="instance"
          :is="tab.component"
          :template="instance"
        />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script>
import config from "@/config.js";

export default {
  name: "instance-view",
  data: () => ({
    tabsIndex: 0,
    navTitles: config.navTitles ?? {},
    tabs: [
      {
        title: "Info",
        component: () => import("@/components/instance/info.vue"),
      },
      {
        title: "Template",
        component: () => import("@/components/instance/template.vue"),
      },
    ],
  }),
  methods: {
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
  },
  computed: {
    instance() {
      const id = this.$route.params?.instanceId;

      return this.$store.getters["services/getInstances"].find(
        ({ uuid }) => uuid === id
      );
    },
    instanceTitle() {
      return this.instance?.title ?? "not found";
    },
    instanceLoading() {
      return this.$store.getters["services/isLoading"];
    },
  },
  created() {
    this.$store.dispatch("services/fetch").then(() => {
      document.title = `${this.instanceTitle} | NoCloud`;
    });
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "services/fetch",
    });
    if (this.namespaces.length < 2) {
      this.$store.dispatch("namespaces/fetch");
    }
    if (this.accounts.length < 2) {
      this.$store.dispatch("accounts/fetch");
    }
    if (this.services.length < 2) {
      this.$store.dispatch("services/fetch");
    }
    if (this.servicesProviders.length < 2) {
      this.$store.dispatch("servicesProviders/fetch");
    }
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
