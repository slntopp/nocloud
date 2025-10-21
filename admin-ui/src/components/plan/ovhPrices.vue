<template>
  <v-card
    :loading="isSpLoading"
    elevation="0"
    color="background-light"
    class="pa-4"
  >
    <v-expansion-panels v-if="!isPlansLoading || !isSpLoading">
      <v-expansion-panel>
        <v-expansion-panel-header color="background">
          Margin rules:
        </v-expansion-panel-header>
        <v-expansion-panel-content color="background">
          <plan-opensrs
            :fee="fee"
            :isEdit="true"
            @changeFee="changeFee"
            @onValid="(data) => (isValid = data)"
          />
          <confirm-dialog
            text="This will apply the rules markup parameters to all prices"
            @confirm="setFee"
          >
            <v-btn class="mt-4" color="secondary">Set rules</v-btn>
          </confirm-dialog>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>

    <component
      v-if="!isSpLoading"
      ref="table"
      :is="tableComponent"
      :fee="fee"
      :sp="sp"
      :template="template"
      :isPlansLoading="isPlansLoading"
      :getPeriod="getPeriod"
      @changeFee="changeFee"
      @changeLoading="isPlansLoading = !isPlansLoading"
    />

    <v-btn class="mt-4" @click="isDialogVisible = true">Save</v-btn>
    <v-dialog :max-width="600" v-model="isDialogVisible">
      <v-card color="background-light">
        <v-card-title
          >Do you really want to change your current price model?</v-card-title
        >
        <v-card-subtitle
          >You can also create a new price model based on the current
          one.</v-card-subtitle
        >
        <v-card-actions>
          <v-btn
            class="mr-2"
            :loading="isCreateLoading"
            :disabled="isEditLoading"
            @click="tryToSend('create')"
          >
            Create
          </v-btn>
          <v-btn
            :loading="isEditLoading"
            :disabled="isCreateLoading"
            @click="tryToSend('edit')"
          >
            Edit
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import planOpensrs from "@/components/plan/opensrs/planOpensrs.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import { replaceNullWithUndefined } from "../../functions";

export default {
  name: "plan-prices",
  components: { planOpensrs, confirmDialog },
  mixins: [snackbar],
  props: { template: { type: Object, required: true } },
  data: () => ({
    fee: {},
    isDialogVisible: false,
    isPlansLoading: false,
    isCreateLoading: false,
    isEditLoading: false,
    isValid: true,
    isSpLoading: false,
  }),
  async mounted() {
    this.isSpLoading = true;
    try {
      await this.$store.dispatch("servicesProviders/fetch", {
        anonymously: true,
      });
    } finally {
      this.isSpLoading = false;
    }
  },
  methods: {
    async tryToSend(action) {
      if (!this.testConfig()) return;
      const newPlan = {
        ...this.template,
        fee: this.fee,
        resources: [],
        products: {},
      };

      const isEdit = action === "edit";
      if (isEdit) {
        this.isEditLoading = true;
      } else {
        this.isCreateLoading = true;
      }
      try {
        const result = await this.$refs.table.changePlan(newPlan);

        if (result === "error") return;
        if (!isEdit) delete newPlan.uuid;

        const request = isEdit
          ? api.plans.update(newPlan.uuid, replaceNullWithUndefined(newPlan))
          : api.plans.create(replaceNullWithUndefined(newPlan));

        await request;

        this.showSnackbarSuccess({
          message: isEdit
            ? "Price model edited successfully"
            : "Price model created successfully",
        });
      } catch (err) {
        const message = err.response?.data?.message ?? err.message ?? err;

        this.showSnackbarError({ message });
        console.error(err);
      } finally {
        this.isCreateLoading = false;
        this.isEditLoading = false;
        this.isDialogVisible = false;
      }
    },
    testConfig() {
      const { testConfig } = this.$refs.table;
      let message = "";

      if (!this.isValid) {
        message = "Margin rules is not valid";
      }
      if (testConfig) message = testConfig() ?? "";

      if (message) {
        this.showSnackbarError({ message });
        return false;
      }

      return true;
    },
    getPeriod(duration) {
      switch (duration) {
        case "P1M":
          return 3600 * 24 * 30;
        case "P1Y":
          return 3600 * 24 * 365;
      }
    },
    setFee() {
      this.$refs.table.setFee();
    },
    changeFee(value) {
      this.fee = JSON.parse(JSON.stringify(value));
    },
  },
  computed: {
    tableComponent() {
      return () =>
        import(
          `@/components/plan/${this.template.type.split(" ")[1]}Table.vue`
        );
    },
    sp() {
      return this.$store.getters["servicesProviders/all"].find(
        (sp) =>
          sp.type === "ovh" && sp.meta?.plans?.includes(this.template.uuid)
      );
    },
  },
};
</script>

<style>
.v-card .v-icon.group-icon {
  display: none;
  margin: 0 0 2px 4px;
  font-size: 18px;
  opacity: 0.5;
  cursor: pointer;
}

.v-data-table__expanded__content {
  background: var(--v-background-base);
}
</style>
