<template>
  <billing-label
    :account="account"
    :template="template"
    :addons-price="addonsPrices"
    :tariff-price="tariffPrice"
    :due-date="dueDate"
    :renew-disabled="isRenewDisabled"
    @update="emit('update', $event)"
  />
</template>

<script setup>
import { computed, onMounted, ref, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";
import billingLabel from "@/components/ui/billingLabel.vue";

const props = defineProps(["template", "account", "addons"]);
const emit = defineEmits(["update"]);

const { template, account, addons } = toRefs(props);

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

const dueDate = computed(() => {
  return formatSecondsToDate(+template.value?.data?.next_payment_date);
});

const isRenewDisabled = computed(
  () => template.value.billingPlan.kind === "DYNAMIC"
);
</script>

<style scoped></style>
