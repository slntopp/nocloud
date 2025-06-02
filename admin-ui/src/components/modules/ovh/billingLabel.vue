<template>
  <billing-label
    :due-date="dueDate"
    :tariff-price="tariffPrice"
    :template="template"
    :addons-price="addonsPrices"
    :account="account"
    @update="emit('update', $event)"
    :renew-disabled="isRenewDisabled"
  />
</template>

<script setup>
import { computed, onMounted, ref, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";
import billingLabel from "@/components/ui/billingLabel.vue";

const props = defineProps(["template", "account", "addons"]);
const emit = defineEmits(["update"]);

const { template, account, addons } = toRefs(props);

const dueDate = computed(() => {
  return formatSecondsToDate(+props.template?.data?.next_payment_date);
});

const tariffPrice = ref(0);
const addonsPrices = ref({});

onMounted(() => {
  const prices = {};

  addons.value.forEach(
    (a) =>
      (prices[a.title] =
        a.periods[
          template.value.billingPlan.products[template.value.product]?.period
        ])
  );

  addonsPrices.value = prices;

  tariffPrice.value =
    (template.value.estimate || 0) -
    Object.keys(prices).reduce((acc, key) => acc + prices[key], 0);
});

const isRenewDisabled = computed(
  () => template.value.billingPlan.type === "ovh cloud"
);
</script>

<style scoped></style>
