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
      if (this.value) return this.value;

      const { title } = this.$store.getters['auth/userdata'];

      return this.$store.getters['accounts/all']
        .find((account) => account.title === title)
        ?.balance;
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
