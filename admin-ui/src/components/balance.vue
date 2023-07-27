<template>
  <v-chip
    @click="$emit('click')"
    style="cursor: pointer"
    v-if="balance !== undefined"
    :color="colorChip"
  >
    {{ title }}{{ balance }} {{!hideCurrency && (currency || defaultCurrency) || ''}}
  </v-chip>
</template>

<script>
export default {
  name: "balance-display",
  props: ["title", "value", "positive-color", "negative-color", "currency",'hideCurrency'],
  mounted() {
    if (!this.balance) {
      this.$store
        .dispatch("accounts/fetch")
        .catch((err) => console.error(err.toJSON()));
    }
    if (this.defaultCurrency === "") {
      this.$store
        .dispatch("currencies/fetch")
        .catch((err) => console.error(err.toJSON()));
    }
  },
  computed: {
    balance() {
      if (this.value) return parseFloat(this.value).toFixed(2);

      const { balance = 0 } = this.$store.getters["auth/userdata"];

      return parseFloat(balance).toFixed(2);
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
