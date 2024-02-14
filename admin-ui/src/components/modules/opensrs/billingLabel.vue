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
import { useStore } from "@/store";
import billingLabel from "@/components/ui/billingLabel.vue";
import {formatSecondsToDate} from "@/functions";

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

const tariffPrice = ref(template.value.billingPlan.resources[0]?.price ?? 0);

const dueDate = computed(() => formatSecondsToDate(template.value.data.expiry.expiredate));
</script>

<style scoped></style>
