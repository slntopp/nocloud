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
import { computed, ref, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";
import { useStore } from "@/store";
import billingLabel from "@/components/ui/billingLabel.vue";

const props = defineProps(["template"]);
const emit = defineEmits(["update"]);

const store = useStore();

const { template } = toRefs(props);

const account = computed(() => {
  const namespace = store.getters["namespaces/all"]?.find(
      (n) => n.uuid === template.value?.access.namespace
  );
  const account = store.getters["accounts/all"].find(
      (a) => a.uuid === namespace?.access.namespace
  );
  return account;
});

const tariffPrice = ref(
  template.value.billingPlan.products[template.value.product]?.price ?? 0
);

const dueDate = computed(() => {
  return formatSecondsToDate(+template.value?.data?.next_payment_date);
});
</script>

<style scoped></style>
