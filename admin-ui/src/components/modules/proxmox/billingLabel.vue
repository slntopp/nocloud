<template>
  <!-- Shared with IONE (TZ §18.1). -->
  <ione-billing-label v-bind="$attrs" @update="$emit('update', $event)" />
</template>

<script>
import IoneBillingLabel from "@/components/modules/ione/billingLabel.vue";

export default {
  name: "proxmox-billing-label",
  components: { IoneBillingLabel },
  inheritAttrs: false,
};
</script>
