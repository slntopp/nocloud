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
    </div>

    <services-providers v-model="selected"> </services-providers>
  </div>
</template>

<script>
import api from "@/api.js";
import servicesProviders from "@/components/servicesproviders_table.vue";

import snackbar from "@/mixins/snackbar.js";
import ConfirmDialog from "../components/confirmDialog.vue";

export default {
  name: "servicesProviders-create",
  components: {
    servicesProviders,
    ConfirmDialog,
  },
  mixins: [snackbar],
  data() {
    return {
      selected: [],
      allTypes: [],
    };
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
              this.$store.dispatch("servicesProviders/fetch", {
                anonymously: false,
                showDeleted: true,
              });

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
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "servicesProviders/fetch",
      params: { anonymously: false, showDeleted: true },
    });
  },
};
</script>

<style></style>