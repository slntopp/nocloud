<template>
  <div class="pa-4">
    <v-menu :value="true" :close-on-content-click="false">
      <template v-slot:activator="{ on, attrs }">
        <v-icon class="group-icon" v-bind="attrs" v-on="on">mdi-filter</v-icon>
      </template>

      <v-list dense>
        <v-list-item dense v-for="item of filters[column]" :key="item">
          <v-checkbox
            dense
            v-model="selectedFilters[column]"
            :value="item"
            :label="item"
            @change="selectedFilters = Object.assign({}, selectedFilters)"
          />
        </v-list-item>
      </v-list>
    </v-menu>

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
      :custom-sort="sortInstances"
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

      <template v-slot:[`item.billingPlan.title`]="{ item, value }">
        <router-link :to="{ name: 'Plan', params: { planId: item.billingPlan.uuid } }">
          {{ value }}
        </router-link>
      </template>

      <template v-slot:[`item.resources.period`]="{ value }">
        {{ value }} {{ (value > 1) ? 'months' : 'month' }}
        <!--  {{ (item.type === 'goget') ? 'years' : 'months' }} -->
      </template>

      <template v-slot:[`item.resources.cpu`]="{ value }">
        {{ value }} {{ (value > 1) ? 'cores' : 'core' }}
      </template>

      <template v-slot:[`item.resources.ram`]="{ value }">
        {{ value / 1024 }} GB
      </template>

      <template v-slot:[`item.resources.drive_size`]="{ value }">
        {{ value / 1024 }} GB
      </template>

      <template v-slot:[`item.config.template_id`]="{ item, value }">
        {{ getOSName(value, item.sp) }}
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
    types: ["all"],
    column: "",
    filters: {},
    selectedFilters: {},
    newInstance: { isValid: false, service: '', sp: '' },
    rules: {
      req: [(v) => !!v || 'This field is required!']
    },

    isDeleteLoading: false,
    selected: [],
    fetchError: "",
  }),
  methods: {
    changeIcon() {
      setTimeout(() => {
        const headers = document.querySelectorAll('.groupable');

        headers.forEach(({ firstElementChild, children }) => {
          if (!children[1]?.className.includes('group-icon')) {
            const element = document.querySelector('.group-icon');
            const icon = element.cloneNode(true);

            firstElementChild.after(icon);
            icon.style = 'display: inline-flex';

            icon.addEventListener('click', (e) => {
              const menu = document.querySelector('.v-menu__content');
              const { x, y } = icon.getBoundingClientRect();

              if (menu.className.includes('menuable__content__active')) return;

              this.column = firstElementChild.innerText;
              element.dispatchEvent(new Event('click'));
              e.stopPropagation();

              setTimeout(() => {
                const width = document.documentElement.offsetWidth;
                const menuWidth = menu.offsetWidth;
                let marginLeft = 20;

                if (width < menuWidth + x) marginLeft = width - (menuWidth + x) - 35;
                const marginTop = (marginLeft < 20) ? 20 : 0

                menu.style.left = `${x + marginLeft + window.scrollX}px`;
                menu.style.top = `${y + marginTop + window.scrollY}px`;
              }, 100);
            });
          }
        });
      }, 100);
    },
    changeFilters() {
      const instances = this.$store.getters["services/getInstances"];

      this.filters = {};
      this.selectedFilters = {};

      instances.forEach((inst) => {
        if (!this.types.includes(inst.type)) this.types.push(inst.type);

        for (let i = 0; i < this.headers.length; i++) {
          const el = this.headers[i];

          if (!el.class) continue;
          let filter = inst;

          el.value.split('.').forEach((key) => {
            filter = filter[key];
          });

          if (this.type !== "all" && inst.type !== this.type) continue;
          switch (el.text) {
            case "Service":
              filter = this.getService(inst);
              break;
            case "Status":
              filter = this.getState(inst).toLowerCase();
              break;
            case "OS":
              filter = this.getOSName(inst.config.template_id, inst.sp);
              break;
            case "RAM":
            case "Disk":
              filter = filter / 1024;
          }

          if (filter === undefined) continue;
          if (!this.filters[el.text]) {
            this.selectedFilters[el.text] = [];
            this.filters[el.text] = [];
          }
          if (!this.filters[el.text].includes(`${filter}`)) {
            this.selectedFilters[el.text].push(`${filter}`);
            this.filters[el.text].push(`${filter}`);
          }
        }
      });
    },
    fetchServices() {
      this.$store.dispatch("services/fetch")
        .then(() => {
          this.fetchError = "";
          this.changeFilters();
          this.changeIcon();
        })
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
    sortInstances(items, sortBy, sortDesc) {
      return items.sort((a, b) => {
        for (let i = 0; i < sortBy.length; i++) {
          if (sortDesc[i]) [a, b] = [b, a];

          let valueA = a;
          let valueB = b;

          sortBy[i].split('.').forEach((key) => {
            valueA = valueA[key];
            valueB = valueB[key];
          });

          if (sortBy[i] === "state") {
            return this.getState(a) < this.getState(b);
          }

          if (sortBy[i] === "service") {
            return this.getService(a) < this.getService(b);
          }

          if (typeof valueA === "string") {
            return valueA.toLowerCase() < valueB.toLowerCase();
          }
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
          return state?.replaceAll('_', ' ') ?? '';
      }
    },
    getService({ service }) {
      return this.services.find(({ uuid }) => service === uuid)?.title ?? '';
    },
    getOSName(id, sp) {
      if (!id) return;
      return this.sp.find(({ uuid }) => uuid === sp).publicData.templates[id].name;
    },
  },
  created() {
    this.fetchServices();
    this.$store.dispatch("servicesProviders/fetch", false);
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "services/fetch",
    });

    const icon = document.querySelector('.group-icon');

    icon.dispatchEvent(new Event('click'));
  },
  computed: {
    services() {
      return this.$store.getters["services/all"];
    },
    instances() {
      const instances = this.$store.getters["services/getInstances"]
        .filter((inst) => {
          const result = [];

          Object.entries(this.selectedFilters).forEach(([key, filters]) => {
            const { value, text } = this.headers.find(({ text }) => text === key) || {};
            let filter = inst;

            if (!value) return [];
            value.split('.').forEach((key) => {
              filter = filter[key];
            });

            switch (text) {
              case "Service":
                filter = this.getService(inst);
                break;
              case "Status":
                filter = this.getState(inst).toLowerCase();
                break;
              case "OS":
                filter = this.getOSName(inst.config.template_id, inst.sp);
                break;
              case "RAM":
              case "Disk":
                filter = filter / 1024;
            }

            if (filters.includes(`${filter}`)) result.push(true);
            else result.push(false);
          });

          return result.every((el) => el);
        });

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
    sp() {
      return this.$store.getters['servicesProviders/all'];
    },
    headers() {
      const headers = [
        { text: "Title", value: "title" },
        { text: "Type", value: "type" },
        { text: "Status", value: "state", class: "groupable" },
        { text: "UUID", value: "uuid" },
        { text: "Service", value: "service", class: "groupable" },
        { text: "Price model", value: "billingPlan.title", class: "groupable" },
      ];

      switch (this.type) {
        case 'ione':
          headers.push(
            { text: "IP", value: "state.meta.networking" },
            { text: "CPU", value: "resources.cpu", class: "groupable" },
            { text: "RAM", value: "resources.ram", class: "groupable" },
            { text: "Disk", value: "resources.drive_size", class: "groupable" },
            { text: "OS", value: "config.template_id", class: "groupable" },
          );
          break;
        case 'ovh':
          headers.push(
            { text: "Tariff", value: "config.planCode", class: "groupable" },
            { text: "IP", value: "state.meta.networking" },
            { text: "Creation", value: "data.creation", class: "groupable" },
            { text: "Expiration", value: "data.expiration", class: "groupable" },
          );
          break;
        case 'goget':
        case 'opensrs':
          headers.push(
            { text: "Domain", value: "resources.domain", class: "groupable" },
            { text: "Period", value: "resources.period", class: "groupable" },
          );

          if (this.type === 'opensrs') break;
          else headers.push(
            { text: "DCV", value: "resources.dcv", class: "groupable" },
            { text: "Email", value: "resources.email", class: "groupable" }
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
    },
    headers() {
      const headers = document.querySelectorAll('.groupable');

      headers.forEach(({ children }) => {
        children[1].remove();
      });

      this.changeIcon();
      this.changeFilters();
    }
  }
}
</script>

<style>
.pa-4 .v-icon.group-icon {
  display: none;
  margin: 0 0 2px 4px;
  font-size: 18px;
  opacity: 0.5;
  cursor: pointer;
}
</style>
