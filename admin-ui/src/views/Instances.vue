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
            item-value="title"
            label="account"
            style="width: 300px"
            v-model="newInstance.account"
            :items="accounts"
            :rules="rules.req"
          />

          <v-select
            dense
            label="type"
            style="width: 300px"
            v-model="newInstance.type"
            :items="allTypes"
            :rules="rules.req"
          />
          <v-select
            label="service"
            item-text="title"
            item-value="uuid"
            v-if="selectedAccountServices.length > 1"
            v-model="newInstance.service"
            :items="selectedAccountServices"
            :rules="rules.req"
          />
          <v-text-field
            label="type name"
            v-if="newInstance.type === 'custom'"
            v-model="newInstance.customTitle"
            :rules="rules.req"
          />

          <v-btn
            :to="{
              name: 'Service edit',
              params: {
                serviceId,
                type:
                  newInstance.type === 'custom'
                    ? newInstance.customTitle
                    : newInstance.type,
              },
            }"
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
      class="d-inline-block mr-2"
      v-model="type"
      :items="types"
    />

    <instances-table
      v-model="selected"
      :type="type"
      :column="column"
      :selected="selectedFilters"
      :get-state="getState"
      :change-filters="changeFilters"
      @getHeaders="(value) => (headers = value)"
      @changeColumn="(value) => (column = value)"
    />

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
import confirmDialog from "@/components/confirmDialog.vue";
import instancesTable from "../components/instances_table.vue";

export default {
  name: "instances-view",
  components: { confirmDialog, instancesTable },
  mixins: [snackbar],
  data: () => ({
    type: "all",
    types: ["all"],
    allTypes: [],
    column: "",
    filters: {},
    selectedFilters: {},
    newInstance: {
      isValid: false,
      service: [],
      type: "",
      customTitle: "",
      account: "",
    },
    rules: {
      req: [(v) => !!v || "This field is required!"],
    },
    isDeleteLoading: false,
    selected: [],
    headers: [],
  }),
  methods: {
    deleteSelectedInstances() {
      if (this.selected.length > 0) {
        const deletePromises = this.selected.map((el) =>
          api.delete(`/instances/${el.uuid}`)
        );

        Promise.all(deletePromises)
          .then((res) => {
            if (res.every(({ result }) => result)) {
              const ending = deletePromises.length === 1 ? "" : "s";

              this.$store.dispatch("services/fetch");
              this.showSnackbarSuccess({
                message: `Instance${ending} deleted successfully.`,
              });
            } else {
              this.showSnackbarError({
                message: `Error: ${
                  res.response?.data?.message ?? res.message ?? "Unknown"
                }.`,
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

          el.value.split(".").forEach((key) => {
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
    getState(item) {
      if (!item.state) return "UNKNOWN";
      const state =
        item.billingPlan.type === "ione"
          ? item.state.meta?.lcm_state_str
          : item.state.state;

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
          return state?.replaceAll("_", " ") ?? "";
      }
    },
    getService({ service }) {
      return this.services.find(({ uuid }) => service === uuid)?.title ?? "";
    },
    getOSName(id, sp) {
      if (!id) return;
      return this.sp.find(({ uuid }) => uuid === sp).publicData.templates[id]
        .name;
    },
  },
  created() {
    this.$store.dispatch("accounts/fetch", false);
    this.$store.dispatch("namespaces/fetch", false);
    this.$store.dispatch("servicesProviders/fetch", false);

    const types = require.context(
      "@/components/modules/",
      true,
      /serviceCreate\.vue$/
    );

    types.keys().forEach((key) => {
      const matched = key.match(/\.\/([A-Za-z0-9-_,\s]*)\/serviceCreate\.vue/i);
      if (matched && matched.length > 1) {
        this.allTypes.push(matched[1]);
      }
    });
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "services/fetch",
    });

    const icon = document.querySelector(".group-icon");
    icon.dispatchEvent(new Event("click"));
  },
  computed: {
    services() {
      return this.$store.getters["services/all"];
    },
    accounts() {
      return this.$store.getters["accounts/all"];
    },
    sp() {
      return this.$store.getters["servicesProviders/all"];
    },
    service() {
      return this.services.find(
        ({ uuid }) => uuid === this.newInstance.service
      );
    },
    selectedAccountServices() {
      if (!this.newInstance.account) {
        return [];
      }

      return this.services.filter((s) => s.title === this.newInstance.account);
    },
    serviceId() {
      if (this.newInstance.service) {
        return this.newInstance.service;
      }
      return this.selectedAccountServices[0]?.uuid;
    },
  },
  watch: {
    "newInstance.account"() {
      this.newInstance.service = null;
    },
  },
};
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
