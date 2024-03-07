<template>
  <billing-label
    :account="account"
    :template="template"
    :tariff-price="tariffPrice"
    :due-date="dueDate"
    renew-disabled
    :addons-price="{}"
    @update="emit('update', $event)"
  />
</template>

<script setup>
import { computed, ref, toRefs } from "vue";
import billingLabel from "@/components/ui/billingLabel.vue";
import { formatSecondsToDate } from "@/functions";

const props = defineProps(["template", "account"]);
const emit = defineEmits(["update"]);

const { template, account } = toRefs(props);

const tariffPrice = ref(template.value.billingPlan.resources[0]?.price ?? 0);

const dueDate = computed(() =>
  formatSecondsToDate(template.value.data.expiry.expiredate)
);
</script>

<style scoped></style>
