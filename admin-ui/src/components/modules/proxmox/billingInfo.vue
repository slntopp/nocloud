<template>
  <!-- Billing screens are shared with IONE (TZ §18.1): same STATIC/prepaid model and data keys. -->
  <ione-billing-info
    v-bind="$attrs"
    @update="$emit('update', $event)"
    @refresh="$emit('refresh', $event)"
  />
</template>

<script>
import IoneBillingInfo from "@/components/modules/ione/billingInfo.vue";

export default {
  name: "proxmox-billing-info",
  components: { IoneBillingInfo },
  inheritAttrs: false,
};
</script>
