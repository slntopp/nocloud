<template>
  <billing-label
    :due-date="dueDate"
    :tariff-price="tariffPrice"
    :template="template"
    :addons-price="addonsPrice"
    :account="account"
    @update="emit('update', $event)"
  />
</template>

<script setup>
import { computed, ref, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";
import billingLabel from "@/components/ui/billingLabel.vue";
import { useStore } from "@/store";

const props = defineProps(["template"]);
const emit = defineEmits(["update"]);

const { template } = toRefs(props);

const store = useStore();

const account = computed(() => {
  const namespace = store.getters["namespaces/all"]?.find(
      (n) => n.uuid === template.value?.access.namespace
  );
  const account = store.getters["accounts/all"].find(
      (a) => a.uuid === namespace?.access.namespace
  );
  return account;
});

const dueDate = computed(() => {
  return formatSecondsToDate(+props.template?.data?.next_payment_date);
});

const tariffPrice = computed(() => {
  const { duration, planCode } = props.template.config;
  const key = `${duration} ${planCode}`;

  return props.template.billingPlan.products[key]?.price ?? 0;
});
const addonsPrice = ref(
  props.template.config.addons?.reduce((res, addon) => {
    const { price } =
      props.template.billingPlan.resources?.find(
        ({ key }) => key === `${props.template.config.duration} ${addon}`
      ) ?? {};
    let key = "";

    if (addon.includes("ram")) return res;
    if (addon.includes("raid")) return res;
    if (addon.includes("vrack")) key = "Vrack";
    if (addon.includes("bandwidth")) key = "Traffic";
    if (addon.includes("additional")) key = "Additional drive";
    if (addon.includes("snapshot")) key = "Snapshot";
    if (addon.includes("backup")) key = "Backup";
    if (addon.includes("windows")) key = "Windows";

    return { ...res, [key]: +price || 0 };
  }, {})
);
</script>

<style scoped></style>
