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
import { computed, toRefs } from "vue";
import { formatSecondsToDate, getInstancePrice } from "@/functions";
import billingLabel from "@/components/ui/billingLabel.vue";

const props = defineProps(["template", "account", "addons"]);
const emit = defineEmits(["update"]);

const { template, account, addons } = toRefs(props);

const addonsPrices = computed(() => {
  const prices = {};
  template.value.billingPlan.resources.forEach((curr) => {
    if (
      curr.key === `drive_${template.value.resources.drive_type.toLowerCase()}`
    ) {
      const key = "drive";

      return (prices[key] =
        (curr.price * template.value.resources.drive_size) / 1024);
    } else if (template.value.resources[curr.key]) {
      const key = curr.key.replace("_", " ");

      return (prices[key] = curr.price * template.value.resources[curr.key]);
    }
  });

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

const tariffPrice = computed(() => getInstancePrice(template.value));

const isRenewDisabled = computed(
  () => template.value.billingPlan.kind === "DYNAMIC"
);
</script>

<style scoped></style>
