<template>
  <v-chip
    @click="$emit('click')"
    style="cursor: pointer"
    v-if="balance !== undefined"
    :color="colorChip"
    :small="small"
  >
    {{ title }}{{ abs ? Math.abs(balance) : balance }}
    {{ (!hideCurrency && (currency?.title || defaultCurrency?.title)) || "" }}
  </v-chip>
</template>

<script>
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
      if (this.logedInUser) {
        const { balance = 0 } = this.$store.getters["auth/userdata"];

        return parseFloat(balance).toFixed(2);
      }

      return parseFloat(this.value || 0).toFixed(2);
    },
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
    colorChip() {
      if (this.balance > 0) {
        return this["positive-color"] || "success";
      } else if (this.balance < 0) {
        return this["negative-color"] || "error";
      } else {
        return "gray";
      }
    },
  },
};
</script>
