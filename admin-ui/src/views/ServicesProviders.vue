<template>
  <div class="servicesProviders pa-4 flex-wrap">
    <div class="buttons__inline pb-8 pt-4">
      <v-btn
        color="background-light"
        class="mr-2"
        :to="{ name: 'ServicesProviders create' }"
      >
        create
      </v-btn>

      <confirm-dialog
        @confirm="deleteSelectedServicesProviders"
        :disabled="selected.length < 1"
      >
        <v-btn
          color="background-light"
          class="mr-2"
          :disabled="selected.length < 1"
        >
          delete
        </v-btn>
      </confirm-dialog>

      <v-btn
        color="background-light"
        :to="{ name: 'ServicesProviders showcases' }"
      >
        showcases
      </v-btn>
    </div>

    <services-providers :searchParams="searchParams" v-model="selected">
    </services-providers>

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
import servicesProviders from "@/components/servicesproviders_table.vue";

import snackbar from "@/mixins/snackbar.js";
import search from "@/mixins/search.js";
import ConfirmDialog from "../components/confirmDialog.vue";

export default {
  name: "servicesProviders-create",
  components: {
    servicesProviders,
    ConfirmDialog,
  },
  mixins: [snackbar, search],
  data() {
    return {
      selected: [],
      allTypes: [],
    };
  },
  computed: {
    searchParams() {
      return this.$store.getters["appSearch/advancedParams"];
    },
  },
  methods: {
    deleteSelectedServicesProviders() {
      if (this.selected.length > 0) {
        const deletePromices = this.selected.map((el) =>
          api.delete(`/sp/${el.uuid}`)
        );
        Promise.all(deletePromices)
          .then((res) => {
            if (res.every((el) => el.result)) {
              console.log("all ok");
              this.$store.dispatch("servicesProviders/fetch");

              const ending = deletePromices.length == 1 ? "" : "s";
              this.showSnackbar({
                message: `Service${ending} provider${ending} deleted successfully.`,
              });
            } else {
              this.showSnackbar({
                message: `Can’t delete Services Provider: Has Services deployed.`,
                route: {
                  name: "Services",
                  query: {
                    filter: "uuid",
                    ["items[]"]: res.find((el) => !el.result).services,
                  },
                },
              });
            }
          })
          .catch((err) => {
            if (err.response.status >= 500 || err.response.status < 600) {
              const opts = {
                message: `Service Unavailable: ${
                  err?.response?.data?.message ?? "Unknown"
                }.`,
                timeout: 0,
              };
              this.showSnackbarError(opts);
            } else {
              const opts = {
                message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
              };
              this.showSnackbarError(opts);
            }
          });
      }
    },
  },
  created() {
    this.$store.dispatch("servicesProviders/fetch").catch((err) => {
      if (err.response.status == 501 || err.response.status == 502) {
        const opts = {
          message: `Service Unavailable: ${err.response.data.message}.`,
          timeout: 0,
        };
        this.showSnackbarError(opts);
      }
    });
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "servicesProviders/fetch",
    });
  },
};
</script>

<style></style>
