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
        <v-col cols="2">
          <v-select
            v-model="selectedPeriod"
            :items="uniqueBillingPeriods"
            label="billing period"
          />
        </v-col>
        <v-col cols="4">
          <v-select
            v-model="product"
            label="tariff"
            item-text="title"
            item-value="key"
            :items="filteredTariffs"
          />
        </v-col>
        <v-col v-if="accountRate">
          <v-text-field
            :suffix="accountCurrency"
            readonly
            :value="isSelectedPlanAvailable ? accountPrice : null"
            label="account price"
          />
        </v-col>
        <v-col>
          <v-text-field
            :suffix="defaultCurrency"
            readonly
            :value="isSelectedPlanAvailable ? fullProduct?.price : null"
            label="price"
          />
        </v-col>
      </v-row>

      <v-row align="center">
        <v-col cols="2">
          <v-text-field
            :value="fullProduct?.resources?.cpu"
            readonly
            label="CPU"
          />
        </v-col>
        <v-col cols="4">
          <v-text-field
            :value="fullProduct?.resources?.ram"
            readonly
            label="RAM"
          />
        </v-col>

        <v-col>
          <v-text-field
            :value="fullProduct?.resources?.drive_type"
            readonly
            label="Drive type"
          />
        </v-col>

        <v-col>
          <v-text-field
            :value="fullProduct?.resources?.drive_size"
            readonly
            label="Drive size"
          />
        </v-col>
      </v-row>

      <v-row justify="end">
        <v-btn class="mx-3" @click="cancel">Cancel</v-btn>
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
import useCurrency from "@/hooks/useCurrency";

const props = defineProps([
  "template",
  "service",
  "value",
  "plans",
  "accountRate",
  "accountCurrency",
]);
const emit = defineEmits(["refresh", "input"]);

const { convertTo, defaultCurrency } = useCurrency();

const { template, plans, service, accountRate, accountCurrency } =
  toRefs(props);

const isChangePMLoading = ref(false);
const plan = ref({});
const product = ref({});
const selectedPeriod = ref("");

onMounted(() => {
  setDefaultPlan();
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

const filteredTariffs = computed(() => {
  const filtred = tariffs.value.filter(
    (t) => billingPeriods.value[t.key] === selectedPeriod.value
  );

  if (template.value?.config?.type === "cloud") {
    return filtred.filter(
      (p) =>
        p.meta.region ===
        template.value.billingPlan.products[originalProduct.value]?.meta.region
    );
  }

  return filtred;
});

const billingPeriods = computed(() => {
  const billingPeriods = {};

  tariffs.value.forEach((t) => {
    billingPeriods[t.key] = getBillingPeriod(t?.period);
  });

  return billingPeriods;
});

const uniqueBillingPeriods = computed(() => {
  const unique = new Map();
  Object.keys(billingPeriods.value).forEach((k) => {
    unique.set(billingPeriods.value[k], billingPeriods.value[k]);
  });
  return [...unique.values()];
});

const isSelectedPlanAvailable = computed(() =>
  filteredTariffs.value.find((t) => product.value === t.key)
);

const originalProduct = computed(() => {
  switch (template.value.type) {
    case "ovh": {
      return (
        template.value.config.duration + " " + template.value.config.planCode
      );
    }
    default: {
      return template.value.product;
    }
  }
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
    default: {
      return (
        template.value.billingPlan.products[template.value.product]?.price || 0
      );
    }
  }
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
  accountCurrency.value && fullProduct.value?.price
    ? convertTo(fullProduct.value.price, accountCurrency.value)
    : 0
);

const changePM = () => {
  const tempService = JSON.parse(JSON.stringify(service.value));

  const igIndex = tempService.instancesGroups.findIndex((ig) =>
    ig.instances.find((i) => i.uuid === template.value.uuid)
  );
  const instanceIndex = tempService.instancesGroups[
    igIndex
  ].instances.findIndex((i) => i.uuid === template.value.uuid);

  tempService.instancesGroups[igIndex].instances[instanceIndex].billingPlan =
    plan.value;
  if (template.value.type.includes("ovh")) {
    const duration = product.value.split(" ")[0];
    tempService.instancesGroups[igIndex].instances[
      instanceIndex
    ].config.duration = duration;
    tempService.instancesGroups[igIndex].instances[
      instanceIndex
    ].config.planCode = product.value.split(" ")[1];
    tempService.instancesGroups[igIndex].instances[
      instanceIndex
    ].config.pricingMode = duration === "P1M" ? "default" : "upfront12";
  }
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

const cancel = () => {
  setDefaultPlan();
  emit("input", false);
};

const setDefaultPlan = () => {
  setPlan();
  setProduct();
  selectedPeriod.value = billingPeriods.value[originalProduct.value];
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
  } else {
    product.value = template.value.product;
  }
};

watch(plans, setPlan);
</script>

<style scoped></style>
