<template>
  <!-- Same STATIC prepaid model and data keys (next_payment_date, last_monitoring) as keyweb. -->
  <keyweb-billing-info
    v-bind="$attrs"
    @update="$emit('update', $event)"
    @refresh="$emit('refresh', $event)"
  />
</template>

<script>
import KeywebBillingInfo from "@/components/modules/keyweb/billingInfo.vue";

export default {
  name: "mondoo-billing-info",
  components: { KeywebBillingInfo },
  inheritAttrs: false,
  // declared so these listeners are not also forwarded through $attrs (Vue 3)
  emits: ["refresh", "update"],
};
</script>
