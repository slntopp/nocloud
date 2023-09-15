<template>
  <v-dialog
    persistent
    :value="value"
    @input="emit('input', value)"
    max-width="80%"
  >
    <v-card class="pa-5">
      <v-card-title class="text-center">Change price model</v-card-title>
      <v-row align="center">
        <v-col cols="12">
          <v-autocomplete
            label="price model"
            item-text="title"
            item-value="uuid"
            return-object
            v-model="plan"
            :items="availablePlans"
          />
        </v-col>
      </v-row>
      <v-row align="center">
        <v-col cols="4">
          <v-select
            v-model="product"
            label="tariff"
            item-text="title"
            item-value="key"
            :items="tariffs"
          />
        </v-col>
        <v-col cols="2">
          <v-text-field
            readonly
            v-model="productBillingPeriod"
            label="billing period"
          />
        </v-col>
        <v-col v-if="accountRate">
          <v-text-field
            :suffix="accountCurrency"
            readonly
            :value="accountPrice"
            label="account price"
          />
        </v-col>
        <v-col>
          <v-text-field
            :suffix="defaultCurrency"
            readonly
            :value="fullProduct?.price"
            label="price"
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
import { toRefs, ref, computed, onMounted, watch } from "vue";
import api from "@/api";
import { getBillingPeriod } from "@/functions";
import useRate from "@/hooks/useRate";
import { useStore } from "@/store";

const props = defineProps([
  "template",
  "service",
  "value",
  "plans",
  "accountRate",
  "accountCurrency",
]);
const emit = defineEmits(["refresh", "input"]);

const { convertFrom } = useRate();

const { template, plans, service, accountRate, accountCurrency } =
  toRefs(props);

const store = useStore();

const isChangePMLoading = ref(false);
const plan = ref({});
const product = ref({});

onMounted(() => {
  setPlan();
  setProduct();
});

const tariffs = computed(() => {
  const tariffs = [];
  Object.keys(plan.value?.products || {}).forEach((key) => {
    if (plan.value.products[key]?.price > instanceTariffPrice.value)
      tariffs.push({ ...plan.value.products[key], key });
  });

  if (plan.value?.uuid === template.value.billingPlan.uuid) {
    tariffs.push({
      ...plan.value.products[originalProduct.value],
      key: originalProduct.value,
    });
  }
  return tariffs;
});

const originalProduct = computed(() => {
  switch (template.value.type) {
    case "ovh": {
      return (
        template.value.config.duration + " " + template.value.config.planCode
      );
    }
    case "ione":
    case "virtual": {
      return template.value.product;
    }
  }

  return null;
});

const availablePlans = computed(() => {
  const availablePlans = [];

  const copyPlans = JSON.parse(JSON.stringify(plans.value)).filter(
    (p) => p.type === template.value.type
  );

  copyPlans.forEach((p) => {
    const keys = Object.keys(p.products).filter(
      (key) => p.products[key]?.price > instanceTariffPrice.value
    );
    if (keys.length > 0) {
      availablePlans.push(p);
    }
  });

  availablePlans.push(template.value.billingPlan);

  return availablePlans;
});

const instanceTariffPrice = computed(() => {
  switch (template.value.type) {
    case "ovh": {
      return template.value.billingPlan.products[
        template.value.config.duration + " " + template.value.config.planCode
      ]?.price;
    }
    case "ione":
    case "virtual": {
      return template.value.billingPlan.products[template.value.product]?.price;
    }
  }

  return 0;
});

const isChangeBtnDisabled = computed(() => {
  return (
    !plan.value ||
    (!product.value && tariffs.value.length > 0) ||
    !product.value
  );
});

const fullProduct = computed(() => plan.value?.products?.[product.value]);

const accountPrice = computed(() =>
  accountRate.value && fullProduct.value?.price
    ? convertFrom(fullProduct.value.price, accountRate.value)
    : 0
);

const defaultCurrency = computed(() => store.getters["currencies/default"]);

const productBillingPeriod = computed(() => {
  return getBillingPeriod(fullProduct.value?.period);
});

const changePM = () => {
  const planCode = product.value.slice(4).toLowerCase().replace(" ", "-");
  const tempService = JSON.parse(JSON.stringify(service.value));

  const igIndex = tempService.instancesGroups.findIndex((ig) =>
    ig.instances.find((i) => i.uuid === template.value.uuid)
  );
  const instanceIndex = tempService.instancesGroups[
    igIndex
  ].instances.findIndex((i) => i.uuid === template.value.uuid);

  tempService.instancesGroups[igIndex].instances[instanceIndex].billingPlan =
    plan.value;
  tempService.instancesGroups[igIndex].instances[
    instanceIndex
  ].config.planCode = planCode;
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

const setPlan = () => {
  plan.value =
    plans.value.find(({ uuid }) => uuid === template.value.billingPlan.uuid) ||
    template.value.billingPlan;
};

const setProduct = () => {
  if (template.value.type === "ovh") {
    product.value =
      template.value.config.duration + " " + template.value.config.planCode;
  } else if (
    template.value.type === "ione" ||
    template.value.type === "virtual"
  ) {
    product.value = template.value.product;
  }
};

watch(plans, setPlan);
</script>

<style scoped></style>
