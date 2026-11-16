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
      bg-color="background-light"
    >
      <v-tab v-for="tab in tabs" :key="tab.title">
        {{ tab.title }}
      </v-tab>
    </v-tabs>

    <v-window
      v-model="tabsIndex"
      style="background: rgb(var(--v-theme-background-light))"
      class="rounded-b-lg"
    >
      <v-window-item v-for="tab in tabs" :key="tab.title">
        <v-progress-linear v-if="loading" indeterminate class="pt-2" />
        <component v-if="!loading && sp" :is="tab.component" :template="sp">
          <v-row>
            <v-col :cols="12" :md="6">
              <json-editor
                :json="sp[tab.title?.toLowerCase()]"
                @changeValue="(data) => (sp[tab.title?.toLowerCase()] = data)"
              />
            </v-col>
          </v-row>
        </component>
      </v-window-item>
    </v-window>
  </div>
</template>

<script>
import { defineAsyncComponent } from "vue";
import config from "@/config.js";
import JsonEditor from "../components/JsonEditor.vue";

export default {
  components: { JsonEditor },
  name: "service-providers-view",
  data: () => ({
    found: false,
    tabsIndex: 0,
    navTitles: config.navTitles ?? {},
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
          component: defineAsyncComponent(() => import("@/components/ServicesProvider/info.vue")),
        },
        {
          title: "Secrets",
          component: defineAsyncComponent(() => import("@/components/ServicesProvider/secrets.vue")),
        },
        {
          title: "Vars",
          component: defineAsyncComponent(() => import("@/components/ServicesProvider/vars.vue")),
        },
        {
          title: "Hooks",
          component: defineAsyncComponent(() => import("@/components/ServicesProvider/hooks.vue")),
        },
        {
          title: "Suspend rules",
          component: defineAsyncComponent(() =>
            import("@/components/ServicesProvider/suspend-rules.vue")),
        },
        {
          title: "Map",
          component:
            this.sp?.type === "ovh"
              ? defineAsyncComponent(() => import("@/components/ServicesProvider/ovhMap.vue"))
              : defineAsyncComponent(() => import("@/components/ServicesProvider/map.vue")),
        },
        {
          title: "Showcase",
          component: defineAsyncComponent(() => import("@/components/ServicesProvider/showcase.vue")),
        },
        {
          title: "Event log",
          component: defineAsyncComponent(() => import("@/components/ServicesProvider/history.vue")),
        },
        {
          title: "Template",
          component: defineAsyncComponent(() => import("@/components/ServicesProvider/template.vue")),
        },
      ].filter((el) => el?.title);

      if (this.sp?.type === "ione")
        tabs.splice(3, 0, {
          title: "Nebula",
          component: defineAsyncComponent(() => import("@/components/ServicesProvider/nebula.vue")),
        });

      if (this.sp?.type === "bitrix24")
        tabs.splice(3, 0, {
          title: "Licences",
          component: defineAsyncComponent(() => import("@/components/ServicesProvider/licences.vue")),
        });

      // ONE-style virtual networks: address ranges, leases, hold/release
      if (this.sp?.type === "proxmox")
        tabs.splice(3, 0, {
          title: "Networks",
          component: defineAsyncComponent(() => import("@/components/modules/proxmox/networks.vue")),
        });

      return tabs;
    },
    title() {
      return this?.sp?.title ?? "not found";
    },
    loading() {
      return this.$store.getters["servicesProviders/isLoading"];
    },
    sp() {
      const items = this.$store.getters["servicesProviders/all"];
      return items.find((el) => el.uuid == this.uuid);
    },
  },
  created() {
    this.$store.dispatch("servicesProviders/fetchById", this.uuid).then(() => {
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
  color: rgb(var(--v-theme-primary));
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>
