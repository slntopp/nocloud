<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-expansion-panels v-if="!isPlansLoading">
      <v-expansion-panel>
        <v-expansion-panel-header color="indigo darken-4">
          Margin rules:
        </v-expansion-panel-header>
        <v-expansion-panel-content color="indigo darken-4">
          <plan-opensrs
            :fee="fee"
            :isEdit="true"
            @changeFee="(data) => (fee = data)"
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
      ref="table"
      :is="tableComponent"
      :fee="fee"
      :template="template"
      :isPlansLoading="isPlansLoading"
      :getPeriod="getPeriod"
      @changeFee="(value) => fee = value"
      @changeLoading="isPlansLoading = !isPlansLoading"
    />

    <v-btn class="mt-4" @click="isDialogVisible = true">Save</v-btn>
    <v-dialog :max-width="600" v-model="isDialogVisible">
      <v-card color="background-light">
        <v-card-title>Do you really want to change your current price model?</v-card-title>
        <v-card-subtitle>You can also create a new price model based on the current one.</v-card-subtitle>
        <v-card-actions>
          <v-btn class="mr-2" :loading="isLoading" @click="tryToSend('create')">
            Create
          </v-btn>
          <v-btn :loading="isLoading" @click="tryToSend('edit')">
            Edit
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

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
  </v-card>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import planOpensrs from '@/components/plan/opensrs/planOpensrs.vue';
import confirmDialog from '@/components/confirmDialog.vue';

export default {
  name: 'plan-prices',
  components: { planOpensrs, confirmDialog },
  mixins: [snackbar],
  props: { template: { type: Object, required: true } },
  data: () => ({
    fee: {},
    isDialogVisible: false,
    isPlansLoading: false,
    isLoading: false,
    isValid: true
  }),
  methods: {
    tryToSend(action) {
      if (!this.testConfig()) return;
      const newPlan = { ...this.template, fee: this.fee, resources: [], products: {} };

      this.$refs.table.changePlan(newPlan);

      if (action === 'create') delete newPlan.uuid;
      const request = (action === 'edit')
        ? api.plans.update(newPlan.uuid, newPlan)
        : api.plans.create(newPlan);

      this.isLoading = true;
      request.then(() => {
        this.showSnackbarSuccess({
          message: (action === 'edit')
            ? "Price model edited successfully"
            : "Price model created successfully",
        });
      })
      .catch((err) => {
        const message = err.response?.data?.message ?? err.message ?? err;

        this.showSnackbarError({ message });
        console.error(err);
      })
      .finally(() => {
        this.isLoading = false;
        this.isDialogVisible = false;
      });
    },
    testConfig() {
      let message = '';
      if (!this.isValid) {
        message = 'Margin rules is not valid';
      }
      message = this.$refs.table.testConfig() ?? '';

      if (message) {
        this.showSnackbarError({ message });
        return false;
      }

      return true;
    },
    getPeriod(duration) {
      switch (duration) {
        case 'P1M':
          return 3600 * 24 * 30;
        case 'P1Y':
          return 3600 * 24 * 30 * 12;
      }
    },
    setFee() {
      this.$refs.table.setFee();
    }
  },
  computed: {
    tableComponent() {
      switch (this.template.type) {
        case 'ovh vps':
          return () => import('@/components/plan/vpsTable.vue');

        default:
          return () => import('@/components/plan/dedicatedTable.vue');
      }
    }
  }
}
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
