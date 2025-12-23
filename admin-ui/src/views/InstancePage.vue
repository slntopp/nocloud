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
      <v-tab v-for="tab of tabs" :key="tab.title"
        >{{ tab.title }}
        <v-chip class="ml-1" small v-if="tab.title === 'notes'">{{
          instance?.adminNotes?.length || 0
        }}</v-chip></v-tab
      >
    </v-tabs>
    <v-tabs-items
      class="rounded-b-lg"
      style="background: var(--v-background-light-base)"
      v-model="tabsIndex"
    >
      <v-tab-item v-for="tab of tabs" :key="tab.title">
        <v-progress-linear
          indeterminate
          class="pt-2"
          v-if="isLoading || isAddonsLoading"
        />
        <component
          v-else
          :is="tab.component"
          :template="instance"
          :addons="addons"
          :account="account"
        />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script>
import config from "@/config.js";
import snackbar from "@/mixins/snackbar";
import api from "@/api";

let socket;

export default {
  name: "instance-view",
  data: () => ({
    tabsIndex: 0,
    navTitles: config.navTitles ?? {},
    isLoading: false,
    account: null,
    addons: [],
    isAddonsLoading: false,
  }),
  mixins: [snackbar],
  methods: {
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
    initSocket() {
      if (!socket && this.instance?.service) {
        const url = `${/^(.*?)\/admin/
          .exec(window.location.href)[1]
          .replace("https", "wss")}/services/${this.instance.service}/stream`;
        socket = new WebSocket(url, [
          "Bearer",
          this.$store.getters["auth/token"],
        ]);
        socket.onmessage = (msg) => {
          const response = JSON.parse(msg.data).result;
          if (!response) {
            this.showSnackbarError({
              message: `Empty response, message:${
                JSON.parse(msg.data).error.message
              }`,
            });
            return;
          }

          try {
            if (response.state) {
              response.state = { state: response.state.state };
            }
            this.$store.commit("services/updateInstance", {
              value: response,
              uuid: this.instance.service,
            });
          } catch {
            socket.close(1000, "job is done");
          }
        };
      }
    },
    async fetchAddons() {
      const addons = [];
      this.isAddonsLoading = true;
      try {
        await Promise.allSettled(
          (this.instance.addons || []).map(async (uuid) => {
            const addon = await this.$store.getters["addons/addonsClient"].get({
              uuid,
            });
            addons.push(addon.toJson());
          })
        );
      } finally {
        this.isAddonsLoading = false;
        this.addons = addons;
      }
    },
    async fetchInstanceData() {
      try {
        this.isLoading = true;
        await Promise.all([
          this.$store.dispatch("instances/get", this.$route.params?.instanceId),
        ]);

        this.$store.dispatch("servicesProviders/fetch", { anonymously: false });
        this.$store.dispatch("services/fetch", {
          filters: { uuid: [this.instance.service] },
        });
        this.$store.dispatch("namespaces/fetch", {
          filters: { uuid: [this.instance.namespace] },
        });
        this.account = await api.accounts.get(this.instance.account);
      } catch (err) {
        console.log(err);
        this.$store.commit("snackbar/showSnackbarError", {
          message: err.message,
        });
      } finally {
        this.isLoading = false;
      }

      this.initSocket();
    },
  },
  computed: {
    instance() {
      const instanceResponse = this.$store.getters["instances/one"];

      if (!instanceResponse) {
        return;
      }

      if (
        !instanceResponse.instance?.billingPlan?.products?.[
          instanceResponse.instance.product
        ]?.price &&
        instanceResponse.instance?.billingPlan?.products?.[
          instanceResponse.instance.product
        ] !== undefined
      ) {
        instanceResponse.instance.billingPlan.products[
          instanceResponse.instance.product
        ].price = 0;
      }

      return {
        ...instanceResponse.instance,
        ...instanceResponse,
        instance: undefined,
      };
    },
    tabs() {
      return [
        {
          title: "Info",
          component: () => import("@/components/instance/info.vue"),
        },
        {
          title: "promocodes",
          component: () => import("@/components/instance/promocodes.vue"),
        },
        this.instance?.state &&
          this.instance.billingPlan.type !== "ovh dedicated" && {
            title: "Snapshots",
            component: () => import("@/components/instance/snapshots.vue"),
          },
        {
          title: "notes",
          component: () => import("@/components/instance/notes.vue"),
        },
        {
          title: "Event log",
          component: () => import("@/components/instance/history.vue"),
        },
        {
          title: "Invoices",
          component: () => import("@/components/instance/invoices.vue"),
        },
        {
          title: "Reports",
          component: () => import("@/components/instance/reports.vue"),
        },
        {
          title: "Helpdesk",
          component: () => import("@/components/instance/chats.vue"),
        },
        {
          title: "Template",
          component: () => import("@/components/instance/template.vue"),
        },
      ].filter((el) => !!el);
    },
    instanceTitle() {
      if (this.isLoading) {
        return "...";
      }

      if (!this.instance) {
        return "Not found";
      }
      return this.instance.title;
    },
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "services/fetch",
      event: this.fetchInstanceData,
    });

    this.fetchInstanceData();
  },
  destroyed() {
    socket?.close(1000, "job is done");
  },
  watch: {
    instance(newVal) {
      if (newVal) {
        this.initSocket();
        this.$store.dispatch("plans/fetchItem", this.instance.billingPlan.uuid);
        this.fetchAddons();
      }
    },
    instanceTitle(newVal) {
      document.title = `${newVal} | NoCloud`;
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
