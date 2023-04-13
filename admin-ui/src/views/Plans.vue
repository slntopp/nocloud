<template>
  <div class="pa-4">
    <v-btn class="mr-2" color="background-light" :to="{ name: 'Plans create' }">
      Create
    </v-btn>

    <confirm-dialog v-if="linked.length < 1" @confirm="deleteSelectedPlan">
      <v-btn
        class="mr-2"
        color="background-light"
        :disabled="selected.length < 1"
        :loading="isDeleteLoading"
      >
        Delete
      </v-btn>
    </confirm-dialog>

    <confirm-dialog
      v-else
      title="You can't delete a price model while there are instances using it!"
      subtitle="To delete price model, select the price model that these instances will use."
      :width="625"
      :disabled="linked.some(({ plan }) => plan === selected[0].uuid)"
      @confirm="deleteSelectedPlan"
    >
      <v-btn
        class="mr-2"
        color="background-light"
        :disabled="selected.length < 1"
        :loading="isDeleteLoading"
      >
        Delete
      </v-btn>
      <template #actions>
        <nocloud-table
          :show-select="false"
          :items="linked"
          :headers="linkedHeaders"
        >
          <template v-slot:[`item.title`]="{ item }">
            <router-link
              :to="{ name: 'Service', params: { serviceId: item.service } }"
            >
              {{ item.title }}
            </router-link>
          </template>
          <template v-slot:[`item.plan`]="{ item }">
            <v-select
              dense
              placeholder="none"
              item-text="title"
              item-value="uuid"
              style="max-width: 200px"
              v-model="item.plan"
              :items="availablePlans"
            />
          </template>
        </nocloud-table>
      </template>
    </confirm-dialog>

    <v-select
      label="Filter by SP"
      item-text="title"
      item-value="uuid"
      class="d-inline-block"
      v-model="serviceProvider"
      :items="servicesProviders"
    />

    <nocloud-table
      table-name="plans"
      single-select
      class="mt-4"
      :items="filtredPlans"
      :headers="headers"
      :value="selected"
      :loading="isLoading"
      :footer-error="fetchError"
      @input="(v) => (selected = v)"
      :filters-values="selectedFilters"
      :filters-items="filterItems"
      @input:filter="selectedFilters[$event.key] = $event.value"
    >
      <template v-slot:[`item.title`]="{ item }">
        <router-link :to="{ name: 'Plan', params: { planId: item.uuid } }">
          {{ item.title }}
        </router-link>
      </template>
      <template v-slot:[`item.kind`]="{ value }">
        {{ value.toLowerCase() }}
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
import nocloudTable from "@/components/table.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import { filterArrayByTitleAndUuid } from "@/functions";

export default {
  name: "plans-view",
  components: { nocloudTable, confirmDialog },
  mixins: [snackbar, search],
  data: () => ({
    headers: [
      { text: "Title ", value: "title" },
      { text: "UUID ", value: "uuid" },
      { text: "Kind ", value: "kind", customFilter: true },
      { text: "Type ", value: "type", customFilter: true },
      { text: "Public ", value: "public", customFilter: true },
    ],
    linkedHeaders: [
      { text: "Instance", value: "title" },
      { text: "Price model", value: "plan" },
    ],

    linked: [],
    isDeleteLoading: false,
    selected: [],
    copyed: -1,
    fetchError: "",
    serviceProvider: null,
    selectedFilters: { type: [], kind: [], public: [] },
  }),
  methods: {
    changePlan() {
      this.linked = [];
      this.services.forEach((service) => {
        service.instancesGroups.forEach(({ instances, sp }) => {
          instances.forEach(({ uuid, title, billingPlan }) => {
            if (billingPlan.uuid === this.selected[0]?.uuid) {
              this.linked.push({
                uuid,
                title,
                sp,
                service: service.uuid,
                plan: billingPlan.uuid,
              });
            }
          });
        });
      });
    },
    deleteSelectedPlan() {
      this.linked.forEach((el) => {
        const service = this.services.find(({ uuid }) => uuid === el.service);
        const group = service.instancesGroups.find(({ sp }) => sp === el.sp);
        const inst = group.instances.find(({ uuid }) => uuid === el.uuid);

        inst.billingPlan = this.plans.find(({ uuid }) => uuid === el.plan);
      });

      const promises = [];
      if (this.linked.length > 0) {
        const services = new Set(this.linked.map(({ service }) => service));

        services.forEach((el) => {
          const service = this.services.find(({ uuid }) => uuid === el);

          promises.push(api.services._update(service));
        });
      }

      this.isDeleteLoading = true;
      Promise.all(promises)
        .then(() => api.plans.delete(this.selected[0].uuid))
        .then(() => {
          this.$store.dispatch("plans/fetch");
          this.showSnackbar({
            message: "Price model deleted successfully.",
          });
        })
        .catch((err) => {
          if (err.response.status >= 500 || err.response.status < 600) {
            this.showSnackbarError({
              message: `Price model Unavailable: ${
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
          params: {
            sp_uuid: this.serviceProvider,
          },
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
    this.$store.dispatch("services/fetch");
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
    services() {
      return this.$store.getters["services/all"];
    },
    searchParam() {
      return this.$store.getters["appSearch/param"];
    },
    availablePlans() {
      const plan = this.selected[0];

      if (!plan) return [];
      return this.plans.filter(
        ({ uuid, type }) => uuid !== plan.uuid && type === plan.type
      );
    },
    filtredPlans() {
      const plans = this.plans.filter((plan) => {
        return Object.keys(this.selectedFilters).every(
          (key) =>
            this.selectedFilters[key].length === 0 ||
            this.selectedFilters[key].includes(
              plan[key]?.toString()?.toLowerCase()
            )
        );
      });

      if (this.searchParam) {
        return filterArrayByTitleAndUuid(plans, this.searchParam);
      }
      return plans;
    },
    isLoading() {
      return this.$store.getters["plans/isLoading"];
    },
    servicesProviders() {
      const sp = this.$store.getters["servicesProviders/all"];

      return [...sp];
    },
    filterItems() {
      return {
        kind: ["static", "dynamic"],
        type: this.typeItems,
        public: ["true", "false"],
      };
    },
    typeItems() {
      return [...new Set(this.plans.map((p) => p.type.toLowerCase()))];
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
    selected() {
      this.changePlan();
    },
  },
};
</script>
