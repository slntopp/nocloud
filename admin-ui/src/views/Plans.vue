<template>
  <div class="pa-4">
    <v-btn class="mr-2" :to="{ name: 'Plans create' }"> Create </v-btn>

    <confirm-dialog
      v-if="linked.length !== 1"
      :disabled="isDeleteDisabled"
      @confirm="deleteSelectedPlan"
    >
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
            style="max-width: 80px"
            :items="fileTypes"
            label="File type"
            v-model="selectedFileType"
            class="d-inline-block mx-1"
          />
          <download-template-button
            class="mx-1"
            small
            title="Downlaod"
            :disabled="!selected.length || isPlansUploadLoading"
            name="selected_copy"
            :type="selectedFileType"
            :template="selected"
          />
          <v-file-input
            class="file-input mx-1 mt-4"
            :label="`upload ${selectedFileType} price models...`"
            :accept="`.${selectedFileType}`"
            @change="onJsonInputChange"
          />
          <confirm-dialog
            @confirm="uploadPlans"
            :disabled="!uploadedPlans.length"
            :text="uploadedPlansText"
          >
            <v-btn
              :disabled="!uploadedPlans.length"
              :loading="isPlansUploadLoading"
              small
              >Upload</v-btn
            >
          </confirm-dialog>
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
import {
  compareSearchValue,
  defaultFilterObject,
  filterArrayByTitleAndUuid,
  getDeepObjectValue,
  readJSONFile,
  readYAMLFile,
} from "@/functions";
import { mapGetters } from "vuex";
import DownloadTemplateButton from "@/components/ui/downloadTemplateButton.vue";

export default {
  name: "plans-view",
  components: { DownloadTemplateButton, nocloudTable, confirmDialog },
  mixins: [
    snackbar,
    search({
      name: "billing-plans",
      defaultLayout: {
        title: "Default",
        filter: {
          public: true,
        },
      },
    }),
  ],
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

    fileTypes: ["JSON", "YAML"],
    selectedFileType: "JSON",
    isPlansUploadLoading: false,
    uploadedPlans: [],
  }),
  methods: {
    defaultFilterObject,
    changePlan() {
      this.linked = [];
      this.services.forEach((service) => {
        service.instancesGroups.forEach(({ instances, sp }) => {
          instances.forEach(({ uuid, title, billingPlan, state }) => {
            if (
              billingPlan.uuid === this.selected[0]?.uuid &&
              state?.state !== "DELETED"
            ) {
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
          withCount: true,
          params: {
            anonymously: false,
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
    async onJsonInputChange(file) {
      this.uploadedPlans = [];
      try {
        if (this.isJson) {
          this.uploadedPlans.push(...(await readJSONFile(file)));
        } else {
          this.uploadedPlans.push(...(await readYAMLFile(file)));
        }
      } catch (err) {
        this.uploadedPlans = [];
        this.showSnackbarError({ message: err });
      }
    },
    async uploadPlans() {
      this.isPlansUploadLoading = true;
      try {
        await Promise.all(
          this.uploadedPlans.map((p) =>
            api.plans.create({
              ...p,
              uuid: undefined,
            })
          )
        );
        this.getPlans();
      } catch (err) {
        this.showSnackbarError({ message: err });
      } finally {
        this.uploadedPlans = [];
        this.isPlansUploadLoading = false;
      }
    },
  },
  created() {
    this.$store.dispatch("services/fetch", { showDeleted: true });
    this.$store.dispatch("servicesProviders/fetch");
    this.getPlans();
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "plans/fetch",
      params: {
        params: {
          anonymously: false,
        },
        withCount: true,
      },
    });
  },
  computed: {
    ...mapGetters("plans", {
      plans: "all",
      isLoading: "isLoading",
      isInstanceCountLoading: "isInstanceCountLoading",
      instanceCountMap: "instanceCountMap",
    }),
    ...mapGetters("appSearch", {
      filter: "filter",
      searchParam: "param",
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
      const plans = this.plans.filter((p) =>
        Object.keys(this.filter).every((key) => {
          const data = getDeepObjectValue(p, key);

          return compareSearchValue(
            data,
            this.filter[key],
            this.searchFields.find((f) => f.key === key)
          );
        })
      );

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
    uploadedPlansText() {
      return (
        "Uploaded plans:<br/>" +
        this.uploadedPlans.map((p) => p.title).join("<br/>")
      );
    },
    searchFields() {
      return [
        {
          title: "Title",
          key: "title",
          type: "input",
        },
        {
          items: ["STATIC", "DYNAMIC"],
          title: "Kind",
          key: "kind",
          type: "select",
        },
        { items: this.typeItems, title: "Type", key: "type", type: "select" },
        { title: "Public", key: "public", type: "logic-select" },
      ];
    },
  },
  watch: {
    plans() {
      this.fetchError = "";
    },
    typeItems() {
      this.$store.commit("appSearch/setFields", this.searchFields);
    },
    selected() {
      this.changePlan();
    },
  },
};
</script>

<style scoped>
.file-input {
  max-width: 300px;
  min-width: 200px;
  margin-top: 0;
  padding-top: 0;
}
</style>
