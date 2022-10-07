<template>
  <div class="pa-4">
    <v-btn class="mr-2" color="background-light" :to="{ name: 'Plans create' }">
      Create
    </v-btn>
    <confirm-dialog
      :disabled="this.selected.length < 1"
      @confirm="deleteSelectedPlans"
    >
      <v-btn
        class="mr-2"
        color="background-light"
        :disabled="this.selected.length < 1"
        :loading="isDeleteLoading"
      >
        Delete
      </v-btn>
    </confirm-dialog>

    <v-select
      label="Service Provider"
      item-text="title"
      item-value="uuid"
      class="d-inline-block"
      v-model="serviceProvider"
      :items="servicesProviders"
    />

    <nocloud-table
      class="mt-4"
      :items="filtredPlans"
      :headers="headers"
      :value="selected"
      :loading="isLoading"
      :footer-error="fetchError"
      @input="(v) => (selected = v)"
    >
      <template v-slot:[`item.title`]="{ item }">
        <router-link :to="{ name: 'Plan', params: { planId: item.uuid } }">
          {{ item.title }}
        </router-link>
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
import search from "@/mixins/search.js";
import noCloudTable from "@/components/table.vue";
import ConfirmDialog from "../components/confirmDialog.vue";
import { filterArrayByTitleAndUuid } from "@/functions";

export default {
  name: "plans-view",
  components: {
    "nocloud-table": noCloudTable,
    ConfirmDialog,
  },
  mixins: [snackbar, search],
  data: () => ({
    headers: [
      { text: "Title ", value: "title" },
      { text: "UUID ", value: "uuid" },
      { text: "Public ", value: "public" },
      { text: "Type ", value: "type" },
    ],
    isDeleteLoading: false,
    selected: [],
    copyed: -1,
    fetchError: "",
    serviceProvider: null,
  }),
  methods: {
    deleteSelectedPlans() {
      this.isDeleteLoading = true;

      const deletePromises = this.selected.map((el) =>
        api.plans.delete(el.uuid)
      );
      Promise.all(deletePromises)
        .then(() => {
          const ending = deletePromises.length === 1 ? "" : "s";

          this.$store.dispatch("plans/fetch");
          this.showSnackbar({
            message: `Plan${ending} deleted successfully.`,
          });
        })
        .catch((err) => {
          if (err.response.status >= 500 || err.response.status < 600) {
            this.showSnackbarError({
              message: `Plan Unavailable: ${
                err?.response?.data?.message ?? "Unknown"
              }.`,
              timeout: 0,
            });
          } else {
            this.showSnackbarError({
              message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
            });
          }
        })
        .finally(() => {
          this.isDeleteLoading = false;
        });
    },
    getPlans() {
      this.$store
        .dispatch("plans/fetch", {
          sp_uuid: this.serviceProvider,
          anonymously: false,
        })
        .then(() => {
          this.fetchError = "";
        })
        .catch((err) => {
          console.error(err);

          this.fetchError = "Can't reach the server";
          if (err.response) {
            this.fetchError += `: [ERROR]: ${err.response.data.message}`;
          } else {
            this.fetchError += `: [ERROR]: ${err.toJSON().message}`;
          }
        });
    },
  },
  created() {
    this.$store.dispatch("servicesProviders/fetch");
    this.getPlans();
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "plans/fetch",
      params: {
        sp_uuid: this.serviceProvider,
        anonymously: false,
      },
    });
  },
  computed: {
    plans() {
      return this.$store.getters["plans/all"];
    },
    searchParam() {
      return this.$store.getters["appSearch/param"];
    },
    filtredPlans() {
      if (this.searchParam) {
        return filterArrayByTitleAndUuid(this.plans, this.searchParam);
      }
      return this.plans;
    },
    isLoading() {
      return this.$store.getters["plans/isLoading"];
    },
    servicesProviders() {
      const sp = this.$store.getters["servicesProviders/all"];

      return [...sp, { title: "none", uuid: null }];
    },
  },
  watch: {
    plans() {
      this.fetchError = "";
    },
    serviceProvider() {
      this.getPlans();
      this.$store.commit("reloadBtn/setCallback", {
        type: "plans/fetch",
        params: {
          sp_uuid: this.serviceProvider,
          anonymously: false,
        },
      });
    },
  },
};
</script>
