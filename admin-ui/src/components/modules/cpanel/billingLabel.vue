<template>
  <billing-label
    :account="account"
    :template="template"
    :addons-price="0"
    :tariff-price="tariffPrice"
    :due-date="dueDate"
    @update="emit('update', $event)"
  />
</template>

<script setup>
import { computed, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";
import billingLabel from "@/components/ui/billingLabel.vue";

const props = defineProps(["template", "account"]);
const emit = defineEmits(["update"]);

const { template, account } = toRefs(props);

const tariffPrice = computed(
  () =>
    template.value.billingPlan.products[template.value.resources.plan]?.price ??
    0
);

const dueDate = computed(() => {
  return formatSecondsToDate(+template.value?.data?.next_payment_date);
});
</script>

<style scoped></style>
