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
import { useStore } from "@/store";

const props = defineProps(["template"]);
const emit = defineEmits(["update"]);

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

const instancePrice = computed(() => {
  const key = props.template.product;
  const tarrif = props.template.billingPlan.products[key]?.price ?? 0;

  return (
    tarrif +
    props.template.config?.addons?.reduce(
      (acc, a) =>
        acc +
        props.template.billingPlan?.resources?.find((r) => r.key === a)?.price,
      0
    )
  );
});
</script>

<style scoped></style>
