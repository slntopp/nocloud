<template>
  <div class="servicesProviders pa-4 flex-wrap">
    <div class="page__title mb-5">
      <router-link :to="{ name: 'ServicesProviders' }">{{
        navTitle("Services Providers")
      }}</router-link>
      /
      {{ title }}
    </div>

    <v-tabs
      class="rounded-t-lg"
      v-model="tabsIndex"
      background-color="background-light"
    >
      <v-tab v-for="tab in tabs" :key="tab.title">
        {{ tab.title }}
      </v-tab>
    </v-tabs>

    <v-tabs-items
      v-model="tabsIndex"
      style="background: var(--v-background-light-base)"
      class="rounded-b-lg"
    >
      <v-tab-item v-for="tab in tabs" :key="tab.title">
        <v-progress-linear v-if="loading" indeterminate class="pt-2" />
        <component v-if="!loading && item" :is="tab.component" :template="item">
          <v-row>
            <v-col :cols="12" :md="6">
              <json-editor
                :json="item[tab.title?.toLowerCase()]"
                @changeValue="(data) => (item[tab.title?.toLowerCase()] = data)"
              />
            </v-col>
          </v-row>
        </component>
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script>
import config from "@/config.js";
import JsonEditor from "../components/JsonEditor.vue";

export default {
  components: { JsonEditor },
  name: "service-providers-view",
  data: () => ({
    found: false,
    tabsIndex: 0,
    navTitles: config.navTitles ?? {},
    item: null,
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
    uuid() {
      return this.$route.params.uuid;
    },
    tabs() {
      const tabs = [
        {
          title: "Info",
          component: () => import("@/components/ServicesProvider/info.vue"),
        },
        {
          title: "Secrets",
          component: () => import("@/components/ServicesProvider/secrets.vue"),
        },
        {
          title: "Vars",
          component: () => import("@/components/ServicesProvider/vars.vue"),
        },
        {
          title: "Hooks",
          component: () => import("@/components/ServicesProvider/hooks.vue"),
        },
        {
          title: "Suspend rules",
          component: () =>
            import("@/components/ServicesProvider/suspend-rules.vue"),
        },
        {
          title: "Map",
          component:
            this.item?.type === "ovh"
              ? () => import("@/components/ServicesProvider/ovhMap.vue")
              : () => import("@/components/ServicesProvider/map.vue"),
        },
        {
          title: "Showcase",
          component: () => import("@/components/ServicesProvider/showcase.vue"),
        },
        {
          title: "Event log",
          component: () => import("@/components/ServicesProvider/history.vue"),
        },
        {
          title: "Template",
          component: () => import("@/components/ServicesProvider/template.vue"),
        },
      ].filter((el) => el?.title);

      if (this.item?.type === "ione")
        tabs.splice(3, 0, {
          title: "Nebula",
          component: () => import("@/components/ServicesProvider/nebula.vue"),
        });

      return tabs;
    },
    title() {
      return this?.item?.title ?? "not found";
    },
    loading() {
      return this.$store.getters["servicesProviders/isLoading"];
    },
  },
  created() {
    this.$store.dispatch("servicesProviders/fetchById", this.uuid).then(() => {
      const items = this.$store.getters["servicesProviders/all"];
      this.item = items.find((el) => el.uuid == this.uuid);
      this.found = !!this.service;
      document.title = `${this.title} | NoCloud`;
    });
  },
  mounted() {
    document.title = `${this.title} | NoCloud`;
    this.$store.commit("reloadBtn/setCallback", {
      type: "servicesProviders/fetchById",
      params: this.uuid,
    });
  },
};
</script>

<style>
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>
