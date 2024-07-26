<template>
  <billing-label
    :due-date="dueDate"
    :tariff-price="instancePrice"
    :template="template"
    :addons-price="{}"
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

const { template, account, addons } = toRefs(props);

const dueDate = computed(() => {
  return formatSecondsToDate(+props.template?.data?.next_payment_date);
});

const instancePrice = computed(() => {
  const key = props.template.product;
  const tariff = props.template.billingPlan.products[key];

  return (
    (+tariff.price || 0) +
    (addons.value.reduce(
      (acc, a) => acc + a.periods?.[tariff.period] || 0,
      0
    ) || 0)
  );
});
</script>

<style scoped></style>
