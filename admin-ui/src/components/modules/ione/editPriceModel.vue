<template>
  <v-dialog
    persistent
    :value="value"
    @input="emit('input', value)"
    max-width="60%"
  >
    <v-card class="pa-5">
      <v-card-title class="text-center">Change price model</v-card-title>
      <v-row align="center">
        <v-col cols="9">
          <v-select
            label="price model"
            item-text="title"
            item-value="uuid"
            return-object
            v-model="plan"
            :items="plans"
          />
        </v-col>
      </v-row>
      <v-row align="center">
        <v-col cols="9">
          <v-select
            label="product"
            item-text="title"
            item-value="uuid"
            v-model="product"
            v-if="products.length > 0"
            :items="products"
          />
        </v-col>
      </v-row>

      <v-row justify="end">
        <v-btn class="mx-3" @click="emit('input', false)">Close</v-btn>
        <v-btn
          class="mx-3"
          :loading="isChangePMLoading"
          :disabled="isChangeBtnDisabled"
          @click="changePM"
          >Change price model</v-btn
        >
      </v-row>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { onMounted, toRefs, ref, computed } from "vue";
import api from "@/api";

const props = defineProps(["template", "service", "value", "plans"]);
const emit = defineEmits(["refresh", "input"]);

const { template, plans, service } = toRefs(props);

const isChangePMLoading = ref(false);
const plan = ref({});
const product = ref({});

const changePM = () => {
  if (products.value.length === 0) {
    product.value = null;
  }

  const tempService = JSON.parse(JSON.stringify(service.value));

  const igIndex = tempService.instancesGroups.findIndex((ig) =>
    ig.instances.find((i) => i.uuid === template.value.uuid)
  );
  const instanceIndex = tempService.instancesGroups[
    igIndex
  ].instances.findIndex((i) => i.uuid === template.value.uuid);

  tempService.instancesGroups[igIndex].instances[instanceIndex].billingPlan =
    plan.value;
  tempService.instancesGroups[igIndex].instances[instanceIndex].product =
    product.value;

  if (product.value) {
    Object.keys(plan.value.products[product.value].resources).forEach((key) => {
      tempService.instancesGroups[igIndex].instances[instanceIndex].resources[
        key
      ] = plan.value.products[product.value].resources[key];
    });
  }

  isChangePMLoading.value = true;
  api.services
    ._update(tempService)
    .then(() => {
      emit("refresh");
    })
    .finally(() => {
      isChangePMLoading.value = false;
      emit("input", false);
    });
};

onMounted(() => {
  plan.value = template.value.billingPlan;
  product.value = template.value.product;
});

const products = computed(() => {
  const products = [];
  Object.keys(plan.value?.products || {}).forEach((key) => {
    products.push(key);
  });

  return products;
});

const isChangeBtnDisabled = computed(() => {
  if (
    !plan.value ||
    (!product.value && products.value.length > 0) ||
    (plan.value.uuid === template.value.billingPlan.uuid &&
      product.value.name === template.value.product)
  ) {
    return true;
  }
  return false;
});
</script>

<style scoped></style>
