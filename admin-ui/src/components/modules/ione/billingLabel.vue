<template>
  <billing-label
    :account="account"
    :template="template"
    :addons-price="addonsPrice"
    :tariff-price="tariffPrice"
    :due-date="dueDate"
    :renew-disabled="isRenewDisabled"
    @update="emit('update', $event)"
  />
</template>

<script setup>
import { computed, ref, toRefs } from "vue";
import { formatSecondsToDate, getInstancePrice } from "@/functions";
import billingLabel from "@/components/ui/billingLabel.vue";

const props = defineProps(["template", "account"]);
const emit = defineEmits(["update"]);

const { template, account } = toRefs(props);

const addonsPrice = ref({});

const dueDate = computed(() => {
  return formatSecondsToDate(+template.value?.data?.next_payment_date);
});

const tariffPrice = computed(() => getInstancePrice(template.value));

const isRenewDisabled = computed(
  () => template.value.billingPlan.kind === "DYNAMIC"
);
</script>

<style scoped></style>
