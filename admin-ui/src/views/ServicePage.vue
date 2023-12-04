<template>
  <div class="service pa-4 h-100">
    <div class="page__title mb-5">
      <router-link :to="{ name: 'Services' }">{{
        navTitle("Services")
      }}</router-link>
      /
      {{ serviceTitle }}
      <v-chip x-small :color="chipColor"> </v-chip>
    </div>

    <v-tabs
      class="rounded-t-lg"
      v-model="tabs"
      background-color="background-light"
    >
      <v-tab>Info</v-tab>
      <v-tab>Reports</v-tab>
      <v-tab>Event log</v-tab>
      <!-- <v-tab>Control</v-tab> -->
      <v-tab>Template</v-tab>
    </v-tabs>

    <v-tabs-items
      v-model="tabs"
      style="background: var(--v-background-light-base)"
      class="rounded-b-lg"
    >
      <v-tab-item>
        <v-progress-linear v-if="servicesLoading" indeterminate class="pt-2" />
        <service-info
          @refresh="fetchService"
          v-if="service"
          :service="service"
          :chipColor="chipColor"
        />
      </v-tab-item>

      <v-tab-item>
        <v-progress-linear v-if="servicesLoading" indeterminate class="pt-2" />
        <service-reports v-if="service" :service="service" />
      </v-tab-item>

      <!-- <v-tab-item>
        <v-progress-linear v-if="servicesLoading" indeterminate class="pt-2" />
        <service-control
          v-if="service"
          :service="service"
          :chip-color="chipColor"
        />
      </v-tab-item> -->

      <v-tab-item>
        <v-progress-linear v-if="servicesLoading" indeterminate class="pt-2" />
        <service-history v-if="service" :template="service" />
      </v-tab-item>
      <v-tab-item>
        <v-progress-linear v-if="servicesLoading" indeterminate class="pt-2" />
        <service-template v-if="service" :template="service" />
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script>
import config from "@/config.js";
import serviceTemplate from "@/components/service/template.vue";
import serviceInfo from "@/components/service/info.vue";
import snackbar from "@/mixins/snackbar.js";
import ServiceHistory from "@/components/service/history.vue";
import ServiceReports from "@/components/service/reports.vue";

let socket;

export default {
  name: "service-view",
  components: { ServiceReports, ServiceHistory, serviceTemplate, serviceInfo },
  mixins: [snackbar],
  data: () => ({
    found: false,
    tabs: 0,
    navTitles: config.navTitles ?? {},
  }),
  methods: {
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
    fetchService() {
      this.$store.dispatch("servicesProviders/fetch");
      this.$store.dispatch("services/fetchById", this.serviceId).then(() => {
        this.found = !!this.service;
        document.title = `${this.serviceTitle} | NoCloud`;
      });
      this.$store.dispatch("services/fetch",{showDeleted:true});
    },
  },
  computed: {
    service() {
      const items = this.$store.getters["services/all"];
      const item = items.find((el) => el.uuid == this.serviceId);

      if (item) return item;

      return null;
    },
    serviceId() {
      return this.$route.params.serviceId;
    },
    chipColor() {
      const dict = {
        init: "orange darken-2",
        up: "green darken-2",
        del: "gray darken-2",
      };
      return dict?.[this?.service?.status] ?? "blue-grey darken-2";
    },
    serviceTitle() {
      return this?.service?.title ?? "not found";
    },
    servicesLoading() {
      return this.$store.getters["services/loading"];
    },
  },
  created() {
    this.fetchService();
  },
  mounted() {
    document.title = `${this.serviceTitle} | NoCloud`;
    this.$store.commit("reloadBtn/setCallback", {
      type: "services/fetchById",
      params: this.serviceId,
    });

    this.$store.dispatch("accounts/fetch");

    const url = `${/^(.*?)\/admin/
      .exec(window.location.href)[1]
      .replace("https", "wss")}/services/${this.serviceId}/stream`;
    socket = new WebSocket(url, ["Bearer", this.$store.getters["auth/token"]]);
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
        this.$store.commit("services/updateInstance", {
          value: response,
          uuid: this.serviceId,
        });
      } catch {
        socket.close(1000, "job is done");
      }
    };
  },
  destroyed() {
    socket.close(1000, "job is done");
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
