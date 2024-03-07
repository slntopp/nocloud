<template>
  <billing-label
    :due-date="dueDate"
    :tariff-price="tariffPrice"
    :template="template"
    :addons-price="addonsPrice"
    :account="account"
    @update="emit('update', $event)"
    :renew-disabled="isRenewDisabled"
  />
</template>

<script setup>
import { computed, ref, toRefs } from "vue";
import { formatSecondsToDate } from "@/functions";
import billingLabel from "@/components/ui/billingLabel.vue";

const props = defineProps(["template", "account"]);
const emit = defineEmits(["update"]);

const { template, account } = toRefs(props);

const dueDate = computed(() => {
  return formatSecondsToDate(+props.template?.data?.next_payment_date);
});

const tariffPrice = computed(() => {
  const { duration, planCode } = props.template.config;
  const key = `${duration} ${planCode}`;

  return props.template.billingPlan.products[key]?.price ?? 0;
});

const getAddonKey = (key) => {
  let keys = [];
  if (template.value.config.type === "dedicated") {
    keys = [
      template.value.config.duration,
      template.value.config.planCode,
      key,
    ];
  } else {
    keys = [template.value.config.duration, key];
  }
  return keys.join(" ");
};

const addonsPrice = ref(
  props.template.config.addons?.reduce((res, addon) => {
    const addonKey = getAddonKey(addon);
    const { price } =
      props.template.billingPlan.resources?.find(
        ({ key }) => key === addonKey
      ) ?? {};

    return { ...res, [addonKey]: +price || 0 };
  }, {})
);

const isRenewDisabled = computed(
  () => template.value.billingPlan.type === "ovh cloud"
);
</script>

<style scoped></style>
