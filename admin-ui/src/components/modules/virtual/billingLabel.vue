<template>
  <billing-label
    :due-date="dueDate"
    :tariff-price="tariffPrice"
    :template="template"
    :addons-price="{}"
    :account="account"
  />
</template>

<script setup>
import { computed, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";
import billingLabel from "@/components/ui/billingLabel.vue";
import { useStore } from "@/store";

const props = defineProps(["template"]);

const { template } = toRefs(props);

const store = useStore();

const account = computed(() => {
  const namespace = store.getters["namespaces/all"]?.find(
    (n) => n.uuid === template.value.access.namespace
  );
  const account = store.getters["accounts/all"].find(
    (a) => a.uuid === namespace.access.namespace
  );
  return account;
});

const dueDate = computed(() => {
  return formatSecondsToDate(+props.template?.data?.next_payment_date);
});

const tariffPrice = computed(() => {
  const key = props.template.product;

  return props.template.billingPlan.products[key]?.price ?? 0;
});
</script>

<style scoped></style>
