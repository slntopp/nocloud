<template>
  <div class="services pa-4">
    <div class="pb-8 pt-4 buttons__inline">
      <v-btn
        color="background-light"
        class="mr-2"
        :to="{ name: 'Service create' }"
      >
        create
      </v-btn>
      <confirm-dialog
        :disabled="selected.length < 1"
        @confirm="deleteSelectedServices"
      >
        <v-btn
          :disabled="selected.length < 1"
          color="background-light"
          class="mr-2"
        >
          delete
        </v-btn>
      </confirm-dialog>
      <v-btn
        color="background-light"
        class="mr-2"
        @click="upServices"
        :disabled="selected.length < 1"
      >
        UP
      </v-btn>
      <v-btn
        color="background-light"
        class="mr-2"
        @click="downServices"
        :disabled="selected.length < 1"
      >
        DOWN
      </v-btn>
    </div>

    <services-table v-model="selected" :refetch="refetch"></services-table>
  </div>
</template>

<script>
import api from "@/api";
import servicesTable from "@/components/servicesTable.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import snackbar from "@/mixins/snackbar.js";

export default {
  name: "Services-view",
  components: {
    servicesTable,
    confirmDialog,
  },
  mixins: [snackbar],
  data: () => ({
    selected: [],
    refetch:false
  }),
  computed: {
    services() {
      return this.$store.getters["services/all"];
    },
    servicesProviders() {
      return this.$store.getters["servicesProviders/all"];
    },
  },
  created() {
    this.$store.dispatch("servicesProviders/fetch", { anonymously: true });
  },
  methods: {
    deleteSelectedServices() {
      if (this.selected.length > 0) {
        const deletePromices = this.selected.map((el) =>
          api.services.delete(el.uuid)
        );
        Promise.all(deletePromices)
          .then((res) => {
            if (res.every((el) => el.result)) {
              console.log("all ok");
              this.$store.dispatch("services/fetch", { showDeleted: true });

              const ending = deletePromices.length == 1 ? "" : "s";
              this.showSnackbar({
                message: `Service${ending} deleted successfully.`,
              });
            } else {
              this.showSnackbarError({
                message: `Can’t delete Services Provider: Has Services deployed.`,
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
    upServices() {
      const servicesToUp = this.getSelectedServicesWithStatus("INIT");
      if (!servicesToUp) {
        return;
      }
      Promise.all(
        servicesToUp.map((s) => {
          return api.services.up(s.uuid);
        })
      );

      this.refetchServices();
    },
    downServices() {
      const servicesToDown = this.getSelectedServicesWithStatus("UP");
      if (!servicesToDown) {
        return;
      }
      Promise.all(
        servicesToDown.map((s) => {
          return api.services.down(s.uuid);
        })
      );
      this.refetchServices();
    },
    getSelectedServicesWithStatus(status) {
      const services = this.selected.filter(
        (service) => service.status === status
      );
      if (!services.length) {
        return;
      }
      return services;
    },
    refetchServices() {
      this.refetch=!this.refetch
    },
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      event:()=>this.refetchServices(),
    });
  },
};
</script>

<style>
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}

.v-expansion-panel-content .v-expansion-panel-content__wrap {
  padding: 22px;
}

.instance-group-status {
  max-width: 30px;
  align-items: center;
  margin-left: 25px;
}
</style>
