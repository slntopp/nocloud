<template>
  <billing-label
    :due-date="dueDate"
    :tariff-price="instancePrice"
    :template="template"
    :addons-price="addonsPrices"
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

const { template, account } = toRefs(props);

const dueDate = computed(() => {
  return formatSecondsToDate(+props.template?.data?.next_payment_date);
});

const instancePrice = computed(() => props.template.estimate);

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
});
</script>

<style scoped></style>
