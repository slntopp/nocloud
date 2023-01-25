<template>
  <div class="pa-4">
    <v-menu offset-y :close-on-content-click="false">
      <template v-slot:activator="{ on, attrs }">
        <v-btn class="mr-2" color="background-light" v-bind="attrs" v-on="on">
          Create
        </v-btn>
      </template>

      <v-card class="pa-4">
        <v-form ref="form" v-model="newInstance.isValid">
          <v-select
            dense
            item-text="title"
            item-value="uuid"
            label="service"
            style="width: 300px"
            v-model="newInstance.service"
            :items="services"
            :rules="rules.req"
          />

          <v-select
            dense
            item-text="title"
            item-value="sp"
            label="group"
            style="width: 300px"
            v-model="newInstance.sp"
            :items="service?.instancesGroups"
            :rules="rules.req"
          />

          <v-btn
            :to="{ name: 'Service edit', params: {
              serviceId: newInstance.service, sp: newInstance.sp
            }}"
            :disabled="!newInstance.isValid"
          >
            OK
          </v-btn>
        </v-form>
      </v-card>
    </v-menu>

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

    <v-select
      label="Filter by type"
      class="d-inline-block"
      v-model="type"
      :items="types"
    />

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

      <template v-slot:[`item.service`]="{ item, value }">
        <router-link :to="{ name: 'Service', params: { serviceId: value } }">
          {{ getService(item) }}
        </router-link>
      </template>

      <template v-slot:[`item.billingPlan`]="{ item }">
        <router-link :to="{ name: 'Plan', params: { planId: item.billingPlan.uuid } }">
          {{ item.billingPlan.title }}
        </router-link>
      </template>

      <template v-slot:[`item.resources.period`]="{ value }">
        {{ value }} {{ (value > 1) ? 'months' : 'month' }}
        <!--  {{ (item.type === 'goget') ? 'years' : 'months' }} -->
      </template>

      <template v-slot:[`item.state.meta.networking`]="{ item }">
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
import { filterArrayIncludes } from "@/functions.js";

export default {
  name: "instances-view",
  components: { nocloudTable, confirmDialog },
  mixins: [snackbar],
  data: () => ({
    type: "all",
    types: ['all', 'ione', 'ovh', 'goget'],
    newInstance: { isValid: false, service: '', sp: '' },
    rules: {
      req: [(v) => !!v || 'This field is required!']
    },

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
    getService({ service }) {
      return this.services.find(({ uuid }) => service === uuid)?.title ?? '';
    },
  },
  created() { this.fetchServices() },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "services/fetch",
    });
  },
  computed: {
    services() {
      return this.$store.getters["services/all"];
    },
    instances() {
      const instances = this.$store.getters["services/getInstances"];
      const filtered = filterArrayIncludes(instances, {
        keys: ["uuid", "service", "title", "billingPlan", "state"],
        value: this.searchParam,
        params: {
          billingPlan: "title",
          service: this.getService,
          state: this.getState
        }
      });

      if (this.type === 'all') return filtered;
      return filtered.filter(({ type }) => type === this.type);
    },
    service() {
      return this.services.find(({ uuid }) => uuid === this.newInstance.service);
    },
    headers() {
      const headers = [
        { text: "Title", value: "title" },
        { text: "Type", value: "type" },
        { text: "Status", value: "state" },
        { text: "UUID", value: "uuid" },
        { text: "Service", value: "service" },
        { text: "Price model", value: "billingPlan" },
      ];

      switch (this.type) {
        case 'ione':
          headers.push({ text: "IP", value: "state.meta.networking" });
          break;
        case 'ovh':
          headers.push(
            { text: "IP", value: "state.meta.networking" },
            { text: "Creation", value: "data.creation" },
            { text: "Expiration", value: "data.expiration" },
          );
          break;
        case 'goget':
        case 'opensrs':
          headers.push(
            { text: "Domain", value: "resources.domain" },
            { text: "Period", value: "resources.period" },
          );
      }

      return headers;
    },
    isLoading() {
      return this.$store.getters["services/isLoading"];
    },
    searchParam() {
      return this.$store.getters["appSearch/param"];
    }
  },
  watch: {
    instances() {
      this.fetchError = "";
    }
  }
}
</script>
