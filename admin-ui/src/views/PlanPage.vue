<template>
  <div class="pa-4 h-100">
    <h1 class="page__title mb-5">
      <router-link :to="{ name: 'Plans' }">{{
        navTitle("Price Models")
      }}</router-link>
      / {{ planTitle }}
      <plan-wiki-icon />
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
        <v-progress-linear indeterminate class="pt-2" v-if="planLoading" />
        <component
          v-else-if="plan"
          :is="tab.component"
          :isEdit="true"
          :item="plan"
          :template="plan"
        />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script>
import config from "@/config.js";
import PlanWikiIcon from "@/components/ui/planWikiIcon.vue";

export default {
  name: "plan-view",
  components: { PlanWikiIcon },
  data: () => ({
    tabsIndex: 0,
    navTitles: config.navTitles ?? {},
    planTitle: "Not found",
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
    plan() {
      return this.$store.getters["plans/one"];
    },
    planLoading() {
      return this.$store.getters["plans/isLoading"];
    },
    tabs() {
      return [
        {
          title: "Info",
          component: () => import("@/views/PlansCreate.vue"),
        },
        this.plan.type === "ione" && {
          title: "Configuration",
          component: () =>
            import("@/components/modules/ione/planConfiguration.vue"),
        },
        {
          title: "Instances",
          component: () => import("@/components/plan/instances.vue"),
        },
        {
          title: "Template",
          component: () => import("@/components/plan/template.vue"),
        },
      ].filter((t) => t);
    },
  },
  created() {
    const id = this.$route.params?.planId;
    this.$store.dispatch("plans/fetchItem", id).then(() => {
      this.planTitle = this.plan.title || this.planTitle;
      document.title = `${this.planTitle} | NoCloud`;
    });
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "plans/fetchItem",
      params: this.$route.params?.planId,
    });
  },
  watch: {
    plan() {
      const pricesComponents = require
        .context("@/components/plan/", true, /\.vue$/)
        .keys();

      if (
        !pricesComponents.includes(
          `./${this.plan.type.split(" ")[0]}Prices.vue`
        )
      )
        return;
      if (this.tabs.find(({ title }) => title === "Prices")) return;
      const type = this.plan.type.split(" ")[0];

      this.tabs.splice(this.tabs.length - 1, 0, {
        title: "Prices",
        component: () => import(`@/components/plan/${type}Prices.vue`),
      });
    },
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
