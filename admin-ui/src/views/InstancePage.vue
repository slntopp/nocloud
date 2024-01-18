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
import snackbar from "@/mixins/snackbar";

let socket;

export default {
  name: "instance-view",
  data: () => ({
    tabsIndex: 0,
    navTitles: config.navTitles ?? {},
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
  },
  computed: {
    instance() {
      const id = this.$route.params?.instanceId;

      return this.$store.getters["services/getInstances"].find(
        ({ uuid }) => uuid === id
      );
    },
    instanceLoading() {
      return this.$store.getters["services/isLoading"];
    },
    tabs() {
      return [
        {
          title: "Info",
          component: () => import("@/components/instance/info.vue"),
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
      if (this.instanceLoading) {
        return "...";
      }

      if (!this.instance) {
        return "Not found";
      }
      return this.instance.title;
    },
  },
  created() {
    this.$store.dispatch("services/fetch", { showDeleted: true });
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "services/fetch",
      params: {
        showDeleted: true,
      },
    });
    this.$store.dispatch("namespaces/fetch");
    this.$store.dispatch("accounts/fetch");
    this.$store.dispatch("servicesProviders/fetch", { anonymously: false });
    this.$store.dispatch("plans/fetch");

    this.initSocket();
  },
  destroyed() {
    socket?.close(1000, "job is done");
  },
  watch: {
    instance(newVal) {
      if (newVal) {
        this.initSocket();
        this.$store.dispatch("plans/fetchItem", this.instance.billingPlan.uuid);
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
