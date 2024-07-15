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
          <accounts-autocomplete
            dense
            label="account"
            style="width: 300px"
            v-model="newInstance.account"
            :rules="rules.req"
          />
          <v-autocomplete
            dense
            :filter="defaultFilterObject"
            label="service"
            item-text="title"
            item-value="uuid"
            style="width: 300px"
            v-if="selectedAccountServices?.length > 1"
            v-model="newInstance.service"
            :items="selectedAccountServices"
            :rules="rules.req"
          />
          <v-autocomplete
            dense
            label="type"
            style="width: 300px"
            v-model="newInstance.type"
            :items="allTypes"
            :rules="rules.req"
          />
          <v-text-field
            dense
            label="type name"
            v-if="newInstance.type === 'custom'"
            style="width: 300px"
            v-model="newInstance.customTitle"
            :rules="rules.req"
          />

          <v-btn
            :to="{
              name: 'Instance create',
              query: {
                serviceId,
                accountId: newInstance.account,
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

    <confirm-dialog
      :disabled="selected.length < 1"
      @confirm="deleteSelectedInstances"
    >
      <v-btn
        class="mr-2"
        color="background-light"
        :disabled="selected.length < 1"
        :loading="isDeleteLoading"
      >
        Terminate
      </v-btn>
    </confirm-dialog>

    <instances-table
      v-model="selected"
      :column="column"
      :items="instances"
      :selected="selectedFilters"
      @getHeaders="(value) => (headers = value)"
      @changeColumn="(value) => (column = value)"
    />
  </div>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import confirmDialog from "@/components/confirmDialog.vue";
import instancesTable from "@/components/instancesTable.vue";
import { defaultFilterObject } from "@/functions";
import AccountsAutocomplete from "@/components/ui/accountsAutocomplete.vue";

export default {
  name: "instances-view",
  components: { AccountsAutocomplete, confirmDialog, instancesTable },
  mixins: [snackbar],
  data: () => ({
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
    defaultFilterObject,
    deleteSelectedInstances() {
      if (this.selected.length > 0) {
        const deletePromises = this.selected.map((el) =>
          api.delete(`/instances/${el.uuid}`)
        );

        Promise.all(deletePromises)
          .then((res) => {
            if (res.every(({ result }) => result)) {
              const ending = deletePromises.length === 1 ? "" : "s";

              this.$store.dispatch("services/fetch", { showDeleted: true });
              this.selected = [];
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
    this.$store.dispatch("namespaces/fetch", false);
    this.$store.dispatch("plans/fetch", { showDeleted: false });
    this.$store.dispatch("servicesProviders/fetch", { anonymously: false });

    const types = require.context(
      "@/components/modules/",
      true,
      /instanceCreate\.vue$/
    );

    types.keys().forEach((key) => {
      const matched = key.match(
        /\.\/([A-Za-z0-9-_,\s]*)\/instanceCreate\.vue/i
      );
      if (matched && matched.length > 1) {
        this.allTypes.push(matched[1]);
      }
    });
  },
  mounted() {
    const icon = document.querySelector(".group-icon");
    icon.dispatchEvent(new Event("click"));
  },
  computed: {
    services() {
      return this.$store.getters["services/all"];
    },
    namespaces() {
      return this.$store.getters["namespaces/all"];
    },
    instances() {
      return this.$store.getters["services/getInstances"];
    },
    sp() {
      return this.$store.getters["servicesProviders/all"].filter(
        (sp) => !!sp.type
      );
    },
    service() {
      return this.services.find(
        ({ uuid }) => uuid === this.newInstance.service
      );
    },
    selectedNamespace() {
      return this.namespaces.find(
        (n) => n.access.namespace === this.newInstance.account
      );
    },
    selectedAccountServices() {
      if (!this.newInstance.account) {
        return [];
      }
      return this.services.filter(
        (s) => s.access.namespace === this.selectedNamespace.uuid
      );
    },
    serviceId() {
      if (this.newInstance.service) {
        return this.newInstance.service;
      }
      return this.selectedAccountServices[0]?.uuid;
    },
    typedServiceProviders() {
      return this.sp.filter((sp) => sp.type === this.newInstance.type);
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
