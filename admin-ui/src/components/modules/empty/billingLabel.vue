<template>
  <billing-label
    :due-date="dueDate"
    :tariff-price="instancePrice"
    :template="template"
    :addons-price="addonsPrice"
    :account="account"
    @update="emit('update', $event)"
  />
</template>

<script setup>
import { computed, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";
import billingLabel from "@/components/ui/billingLabel.vue";

const props = defineProps(["template", "account", "addons"]);
const emit = defineEmits(["update"]);

const { template, addons } = toRefs(props);

const dueDate = computed(() => {
  return formatSecondsToDate(+props.template?.data?.next_payment_date);
});

const instancePrice = computed(() => {
  const key = props.template.product;

  return props.template.billingPlan.products[key]?.price ?? 0;
});

const addonsPrice = computed(() => {
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
</script>

<style scoped></style>
