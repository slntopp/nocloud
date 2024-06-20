<template>
  <billing-label
    :account="account"
    :template="template"
    :addons-price="addonsPrices"
    :tariff-price="tariffPrice"
    :due-date="dueDate"
    @update="emit('update', $event)"
  />
</template>

<script setup>
import { computed, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";
import billingLabel from "@/components/ui/billingLabel.vue";

const props = defineProps(["template", "account", "addons"]);
const emit = defineEmits(["update"]);

const { template, account, addons } = toRefs(props);

const tariffPrice = computed(
  () =>
    template.value.billingPlan.products[template.value.resources.plan]?.price ??
    0
);

const addonsPrices = computed(() => {
  const prices = {};

  addons.value.forEach(
    (a) =>
      (prices[a.uuid] =
        a.periods[
          template.value.billingPlan.products[template.value.product]?.period
        ])
  );
  return prices;
});

const dueDate = computed(() => {
  return formatSecondsToDate(+template.value?.data?.next_payment_date);
});
</script>

<style scoped></style>
