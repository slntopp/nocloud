<template>
  <div class="pa-4">
    <v-btn class="mr-2" color="background-light" :to="{ name: 'Instance create' }">
      Create
    </v-btn>

    <confirm-dialog @confirm="deleteSelectedInstances">
      <v-btn
        class="mr-2"
        color="background-light"
        :disabled="selected.length < 1"
        :loading="isDeleteLoading"
      >
        Delete
      </v-btn>
    </confirm-dialog>

    <nocloud-table
      class="mt-4"
      v-model="selected"
      :items="instances"
      :headers="headers"
      :loading="isLoading"
      :footer-error="fetchError"
    >
      <template v-slot:[`item.title`]="{ item }">
        <router-link :to="{ name: 'Instance', params: { instanceId: item.uuid } }">
          {{ item.title }}
        </router-link>
      </template>

      <template v-slot:[`item.state`]="{ item }">
        <v-chip small :color="chipColor(item)">
          {{ getState(item) }}
        </v-chip>
      </template>

      <template v-slot:[`item.service`]="{ value }">
        <router-link :to="{ name: 'Service', params: { serviceId: value } }">
          {{ getService(value) }}
        </router-link>
      </template>

      <template v-slot:[`item.billingPlan`]="{ item }">
        <router-link :to="{ name: 'Plan', params: { planId: item.billingPlan.uuid } }">
          {{ item.billingPlan.title }}
        </router-link>
      </template>

      <template v-slot:[`item.state.meta`]="{ item }">
        <template v-if="!item.state?.meta.networking?.public">-</template>
        <v-menu bottom
          open-on-hover
          v-else
          nudge-top="20"
          nudge-left="15"
          transition="slide-y-transition"
        >
          <template v-slot:activator="{ on, attrs }">
            <span v-bind="attrs" v-on="on">
              {{ item.state.meta.networking.public[0] }}
            </span>
          </template>

          <v-list dense>
            <v-list-item v-for="net of item.state.meta.networking.public" :key="net">
              <v-list-item-title>{{ net }}</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </template>
    </nocloud-table>

    <v-snackbar
      v-model="snackbar.visibility"
      :timeout="snackbar.timeout"
      :color="snackbar.color"
    >
      {{ snackbar.message }}
      <template v-if="snackbar.route && Object.keys(snackbar.route).length > 0">
        <router-link :to="snackbar.route"> Look up. </router-link>
      </template>

      <template v-slot:action="{ attrs }">
        <v-btn
          :color="snackbar.buttonColor"
          text
          v-bind="attrs"
          @click="snackbar.visibility = false"
        >
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </div>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import nocloudTable from "@/components/table.vue";
import confirmDialog from "@/components/confirmDialog.vue";

export default {
  name: "instances-view",
  components: { nocloudTable, confirmDialog },
  mixins: [snackbar],
  data: () => ({
    headers: [
      { text: "Title", value: "title" },
      { text: "Type", value: "type" },
      { text: "Status", value: "state" },
      { text: "UUID", value: "uuid" },
      { text: "Service", value: "service" },
      { text: "Price model", value: "billingPlan" },
      { text: "IP", value: "state.meta" },
    ],

    isDeleteLoading: false,
    selected: [],
    fetchError: "",
  }),
  methods: {
    fetchServices() {
      this.$store.dispatch("services/fetch")
        .then(() => { this.fetchError = "" })
        .catch((err) => {
          console.log(err);
          this.fetchError = "Can't reach the server";
          if (err.response) {
            this.fetchError += `: [ERROR]: ${err.response.data.message}`;
          } else {
            this.fetchError += `: [ERROR]: ${err.toJSON().message}`;
          }
        });
    },
    deleteSelectedInstances() {
      if (this.selected.length > 0) {
        const deletePromises = this.selected.map((el) =>
          api.delete(`/instances/${el.uuid}`)
        );

        Promise.all(deletePromises)
          .then((res) => {
            if (res.result) {
              const ending = deletePromises.length === 1 ? "" : "s";

              this.$store.dispatch("services/fetch");
              this.showSnackbarSuccess({
                message: `Instance${ending} deleted successfully.`,
              });
            } else {
              this.showSnackbarError({
                message: `Error: ${res.response?.data?.message ?? res.message ?? "Unknown"}.`,
              });
            }
          })
          .catch((err) => {
            if (err.response.status >= 500 || err.response.status < 600) {
              const opts = {
                message: `Service Unavailable: ${
                  err.response?.data?.message ?? err.message ?? "Unknown"
                }.`,
                timeout: 0,
              };
              this.showSnackbarError(opts);
            } else {
              const opts = {
                message: `Error: ${err.response?.data?.message ?? "Unknown"}.`,
              };
              this.showSnackbarError(opts);
            }
          });
      }
    },
    chipColor(item) {
      if (!item.state) return "error";
      const state = (item.config?.os) ? item.state.state : item.state.meta?.lcm_state_str;

      switch (state) {
        case "RUNNING":
          return "success"
        case "LCM_INIT":
        case "STOPPED":
          return "warning"
        case "SUSPENDED":
        case "UNKNOWN":
          return "error"
        default:
          return "blue-grey darken-2"
      }
    },
    getState(item) {
      if (!item.state) return "UNKNOWN";
      const state = (item.config?.os) ? item.state.state : item.state.meta?.lcm_state_str;

      switch (item.state.meta.state) {
        case 1:
          return "PENDING";
        case 5:
          return "SUSPENDED";
        case "BUILD":
          return "BUILD";
      }
      switch (state) {
        case "LCM_INIT":
          return "POWEROFF";
        default:
          return state.replaceAll('_', ' ');
      }
    },
    getService(service) {
      const services = this.$store.getters['services/all'];

      return services.find(({ uuid }) => service === uuid)?.title ?? '';
    },
  },
  created() { this.fetchServices() },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "services/fetch",
    });
  },
  computed: {
    instances() {
      return this.$store.getters['services/getInstances'];
    },
    isLoading() {
      return this.$store.getters['services/isLoading'];
    }
  },
  watch: {
    instances() {
      this.fetchError = "";
    }
  }
}
</script>
