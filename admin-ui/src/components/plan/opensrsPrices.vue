<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-expansion-panels>
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
          >
            <v-btn class="mt-4" color="secondary">Set rules</v-btn>
          </confirm-dialog>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>

    <nocloud-table
      table-name="opensrs-used-domens"
      :show-select="false"
      :items="existedDomens"
      :headers="headers"
    >
      <template v-slot:[`item.price`]="{ item }">
        <v-text-field
          class="mr-2"
          v-model.number="item.price"
          type="number"
          append-icon="mdi-pencil"
        />
      </template>
    </nocloud-table>

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

<script setup>
import api from "@/api.js";
import planOpensrs from "@/components/plan/opensrs/planOpensrs.vue";
import confirmDialog from "@/components/confirmDialog.vue";
import nocloudTable from "@/components/table.vue";
import { computed, ref, toRefs } from "vue";
import { useStore } from "@/store/";
import { getBillingPeriod } from "@/functions";

const props = defineProps({ template: { type: Object, required: true } });
const { template } = toRefs(props);

const store = useStore();

const headers = [
  { text: "Domain", value: "key" },
  { text: "Period", value: "period" },
  { text: "Incoming price", value: "meta.basePrice" },
  { text: "Sale price", value: "price" },
];

const fee = ref(template.value.fee || {});
const isEditLoading = ref(false);
const isCreateLoading = ref(false);
const isDialogVisible = ref(false);
const isValid = ref(true);

const existedDomens = computed(() =>
  Object.keys(template.value.products || {}).map((key) => ({
    ...template.value.products[key],
    key,
    period: getBillingPeriod(template.value.products[key].period),
  }))
);

const tryToSend = async (action) => {
  const products = {};
  existedDomens.value.map((product) => {
    products[product.key] = {
      ...template.value.products[product.key],
      price: product.price,
    };
  });

  const newPlan = {
    ...template.value,
    fee: fee.value,
    products: products,
  };
  const isEdit = action === "edit";
  if (isEdit) {
    isEditLoading.value = true;
  } else {
    isCreateLoading.value = true;
  }

  const request = isEdit
    ? api.plans.update(newPlan.uuid, newPlan)
    : api.plans.create(newPlan);

  try {
    await request;
    store.commit("snackbar/showSnackbarSuccess", {
      message: "Price model edited successfully",
    });
  } catch (err) {
    const message = err.response?.data?.message ?? err.message ?? err;

    store.commit("snackbar/showSnackbarError", { message });
  } finally {
    isEditLoading.value = false;
    isCreateLoading.value = false;
    isDialogVisible.value = false;
  }
};

const changeFee = (value) => {
  fee.value = JSON.parse(JSON.stringify(value));
};
</script>

<script>
export default {
  name: "openrs-prices",
};
</script>

<style></style>
