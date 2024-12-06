<template>
  <v-chip
    @click="$emit('click')"
    style="cursor: pointer"
    v-if="balance !== undefined"
    :color="colorChip"
    :small="small"
  >
    {{ title }}{{ abs ? Math.abs(balance) : balance }}
    {{ (!hideCurrency && (currency?.code || defaultCurrency?.code)) || "" }}
  </v-chip>
</template>

<script>
import { formatPrice } from "../functions";

export default {
  name: "balance-display",
  props: {
    title: {},
    value: {},
    "positive-color": {},
    "negative-color": {},
    currency: {},
    hideCurrency: {},
    logedInUser: { type: Boolean, default: false },
    small: { type: Boolean, default: false },
    abs: { type: Boolean, default: false },
  },
  computed: {
    balance() {
      let value = this.value || 0;
      if (this.logedInUser) {
        const { balance = 0 } = this.$store.getters["auth/userdata"];
        value = balance;
      }

      if (value < 0.01 && value > -1) {
        return 0;
      }

      return formatPrice(value, this.currency || this.defaultCurrency);
    },
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
    colorChip() {
      if (+this.balance > 0) {
        return this["positive-color"] || "success";
      } else if (+this.balance < 0) {
        return this["negative-color"] || "error";
      } else {
        return "gray";
      }
    },
  },
};
</script>
