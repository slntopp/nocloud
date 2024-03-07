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

const props = defineProps(["template", "account"]);
const emit = defineEmits(["update"]);

const { template, account } = toRefs(props);

const dueDate = computed(() => {
  return formatSecondsToDate(+props.template?.data?.next_payment_date);
});

const instancePrice = computed(() => {
  const key = props.template.product;
  const tariff = props.template.billingPlan.products[key];
  const getAddonKey = (addon) =>
    [props.template.config?.configurations[addon], key].join("$");

  const addons = Object.keys(props.template.config?.configurations || {}).map(
    (key) =>
      props.template.billingPlan?.resources?.find(
        (r) => r.key === getAddonKey(key)
      )
  );

  return (
    (+tariff.price || 0) + (addons.reduce((acc, a) => acc + a.price, 0) || 0)
  );
});
</script>

<style scoped></style>
