<template>
  <v-chip v-if="balance !== undefined" :color="colorChip">
    {{ title }}{{ balance }} NCU
  </v-chip>
</template>

<script>
export default {
  name: 'balance-display',
  props: ['title', 'value', 'positive-color', 'negative-color'],
  mounted() {
    if (!this.balance) {
      this.$store.dispatch('accounts/fetch')
        .catch(err => console.error(err.toJSON()));
    }
  },
  computed: {
    balance() {
      if (this.value) return parseFloat(this.value).toFixed(2);

      const { balance = 0 } = this.$store.getters['auth/userdata'];

      return parseFloat(balance).toFixed(2);
    },
    colorChip() {
      if (this.balance > 0) {
        return this['positive-color'] || 'success';
      } else if (this.balance < 0) {
        return this['negative-color'] || 'error';
      } else {
        return 'gray';
      }
    }
  }
}
</script>
