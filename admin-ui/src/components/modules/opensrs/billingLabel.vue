<template>
  <billing-label
    :account="account"
    :template="template"
    :tariff-price="tariffPrice"
    :due-date="dueDate"
    renew-disabled
    :addons-price="addonsPrices"
    @update="emit('update', $event)"
  />
</template>

<script setup>
import { computed, onMounted, ref, toRefs } from "vue";
import billingLabel from "@/components/ui/billingLabel.vue";
import { formatSecondsToDate } from "@/functions";

const props = defineProps(["template", "account"]);
const emit = defineEmits(["update"]);

const { template, account, addons } = toRefs(props);

const tariffPrice = ref(
  template.value.billingPlan.products[template.value.product]?.price ?? 0
);
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

const dueDate = computed(() =>
  formatSecondsToDate(template.value.data.next_payment_date)
);
</script>

<style scoped></style>
