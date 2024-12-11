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
    plan: null,
    isDescriptionsLoading: false,
  }),
  methods: {
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
    async fetchPlan() {
      this.isDescriptionsLoading = true;
      const id = this.$route.params?.planId;

      try {
        await this.$store.dispatch("plans/fetchItem", id);
        this.plan = {
          resources: [],
          products: {},
          meta: {},
          ...this.$store.getters["plans/one"],
        };

        console.log({ ...this.plan });

        this.planTitle = this.plan.title;

        document.title = `${this.planTitle} | NoCloud`;

        const descriptionPromises = [
          ...this.plan.resources?.map((resource, index) => ({
            id: index,
            type: "resources",
            descriptionId: resource.descriptionId,
          })),
          ...Object.keys(this.plan.products || {}).map((key) => ({
            id: key,
            type: "products",
            descriptionId: this.plan.products[key].descriptionId,
          })),
        ];

        const descriptions = await Promise.all(
          descriptionPromises
            .filter((item) => !!item.descriptionId)
            .map(async (item) => ({
              ...item,
              data: await this.$store.dispatch(
                "descriptions/get",
                item.descriptionId
              ),
            }))
        );

        descriptions.forEach(({ type, id, data }) => {
          this.plan[type][id].description = data.text;
        });
      } finally {
        this.isDescriptionsLoading = false;
      }
    },
  },
  computed: {
    planLoading() {
      return (
        this.$store.getters["plans/isLoading"] || this.isDescriptionsLoading
      );
    },
    tabs() {
      return [
        {
          title: "Info",
          component: () => import("@/views/PlansCreate.vue"),
        },
        this.plan?.type === "ione" && {
          title: "Configuration",
          component: () =>
            import("@/components/modules/ione/planConfiguration.vue"),
        },
        {
          title: "Instances",
          component: () => import("@/components/plan/instances.vue"),
        },
        {
          title: "Promocodes",
          component: () => import("@/components/plan/promocodes.vue"),
        },
        {
          title: "Event log",
          component: () => import("@/components/plan/history.vue"),
        },
        {
          title: "Event overrides",
          component: () => import("@/components/plan/event-overrides.vue"),
        },
        {
          title: "Template",
          component: () => import("@/components/plan/template.vue"),
        },
      ].filter((t) => t);
    },
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      event: this.fetchPlan,
    });

    this.fetchPlan();
  },
  watch: {
    plan() {
      const pricesComponents = require
        .context("@/components/plan/", true, /\.vue$/)
        .keys();

      if (
        !pricesComponents.includes(
          `./${this.plan?.type.split(" ")[0]}Prices.vue`
        )
      )
        return;
      if (this.tabs.find(({ title }) => title === "Prices")) return;
      const type = this.plan?.type.split(" ")[0];

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
