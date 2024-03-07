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

const props = defineProps(["template", "account"]);
const emit = defineEmits(["update"]);

const { template } = toRefs(props);

const dueDate = computed(() => {
  return formatSecondsToDate(+props.template?.data?.next_payment_date);
});

const instancePrice = computed(() => {
  const key = props.template.product;

  return props.template.billingPlan.products[key]?.price ?? 0;
});

const addonsPrice = computed(() => {
  const addons = {};

  props.template.config?.addons?.forEach((a) => {
    const r = props.template.billingPlan?.resources?.find((r) => r.key === a);
    addons[r?.title || a] = r?.price || 0;
  });

  return addons;
});
</script>

<style scoped></style>
