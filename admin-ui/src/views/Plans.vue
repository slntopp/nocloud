<template>
  <div class="pa-4">
    <v-btn class="mr-2" :to="{ name: 'Plans create' }"> Create </v-btn>

    <confirm-dialog v-if="linked.length !== 1" @confirm="deleteSelectedPlan">
      <v-btn
        class="mr-2"
        :disabled="isDeleteDisabled"
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
      :success-disabled="linked.some(({ plan }) => plan === selected[0].uuid)"
      @confirm="deleteSelectedPlan"
    >
      <v-btn
        class="mr-2"
        :disabled="selected.length < 1"
        :loading="isDeleteLoading"
      >
        Delete
      </v-btn>
      <template #actions>
        <nocloud-table
          table-name="linked-plans"
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

    <v-autocomplete
      :filter="defaultFilterObject"
      label="Filter by SP"
      item-text="title"
      item-value="uuid"
      class="d-inline-block"
      v-model="serviceProvider"
      :items="servicesProviders"
    />

    <nocloud-table
      table-name="plans"
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
      <template v-slot:[`item.kind`]="{ value }">
        {{ value.toLowerCase() }}
      </template>
      <template v-slot:[`item.instanceCount`]="{ item }">
        <v-progress-circular
          v-if="isInstanceCountLoading"
          size="20"
          indeterminate
        />
        <template v-else>
          {{ instanceCountMap[item.uuid] }}
        </template>
      </template>

      <template v-slot:footer.prepend>
        <div class="d-flex align-center mt-2">
          <v-select
            style="width: 100px"
            :items="fileTypes"
            label="File type"
            v-model="selectedFileType"
            class="d-inline-block mx-2"
          />
          <download-template-button
            class="mx-2"
            :disabled="!selected.length"
            name="selected_copy"
            :type="selectedFileType"
            :template="selected"
          />
        </div>
      </template>
    </nocloud-table>
  </div>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import search from "@/mixins/search.js";
import nocloudTable from "@/components/table.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import { defaultFilterObject, filterArrayByTitleAndUuid } from "@/functions";
import { mapGetters } from "vuex";
import DownloadTemplateButton from "@/components/ui/downloadTemplateButton.vue";

export default {
  name: "plans-view",
  components: { DownloadTemplateButton, nocloudTable, confirmDialog },
  mixins: [snackbar, search],
  data: () => ({
    headers: [
      { text: "Title ", value: "title" },
      { text: "UUID ", value: "uuid" },
      { text: "Kind ", value: "kind" },
      { text: "Type ", value: "type" },
      { text: "Public ", value: "public" },
      { text: "Linked instances count ", value: "instanceCount" },
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

    fileTypes: ["JSON", "YAML"],
    selectedFileType: "JSON",
  }),
  methods: {
    defaultFilterObject,
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
      const deletePromises = this.selected.map((s) => api.plans.delete(s.uuid));

      Promise.all(promises)
        .then(() => Promise.all(deletePromises))
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
          withCount: true,
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
        params: {
          sp_uuid: this.serviceProvider,
          anonymously: false,
        },
        withCount: true,
      },
    });

    this.$store.commit("appSearch/setSearchName", "all-plans");
  },
  computed: {
    ...mapGetters("plans", {
      plans: "all",
      isLoading: "isLoading",
      isInstanceCountLoading: "isInstanceCountLoading",
      instanceCountMap: "instanceCountMap",
    }),
    ...mapGetters("appSearch", {
      searchParams: "customParams",
      searchParam: "customSearchParam",
    }),
    services() {
      return this.$store.getters["services/all"];
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
        return Object.keys(this.searchParams).every(
          (key) =>
            this.searchParams[key].length === 0 ||
            this.searchParams[key]
              .map((f) => f.value)
              .includes(plan[key]?.toString()?.toLowerCase())
        );
      });

      if (this.searchParam) {
        return filterArrayByTitleAndUuid(plans, this.searchParam);
      }
      return plans;
    },
    servicesProviders() {
      const sp = this.$store.getters["servicesProviders/all"];

      return [...sp];
    },
    typeItems() {
      return [...new Set(this.plans.map((p) => p.type.toLowerCase()))];
    },
    isDeleteDisabled() {
      if (
        !Object.keys(this.instanceCountMap).length ||
        this.selected.length === 0
      ) {
        return true;
      }
      const withInstances = this.selected.filter(
        (s) => this.instanceCountMap[s.uuid] > 0
      );
      const withOutInstances = this.selected.filter(
        (s) => this.instanceCountMap[s.uuid] === 0
      );

      return (
        withInstances.length > 1 ||
        (withOutInstances.length > 0 && withInstances.length > 0)
      );
    },
  },
  watch: {
    plans() {
      this.fetchError = "";
    },
    typeItems() {
      this.$store.commit("appSearch/setVariants", {
        kind: { items: ["static", "dynamic"], title: "Kind", isArray: true },
        type: { items: this.typeItems, isArray: true, title: "Type" },
        public: { items: ["true", "false"], title: "Public", isArray: true },
      });
    },
    serviceProvider() {
      this.getPlans();
      this.$store.commit("reloadBtn/setCallback", {
        type: "plans/fetch",
        params: {
          params: {
            sp_uuid: this.serviceProvider,
            anonymously: false,
          },
          withCount: true,
        },
      });
    },
    selected() {
      this.changePlan();
    },
  },
};
</script>
