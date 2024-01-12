<script setup>
const openPlanWiki = () => {
  window.open(
    "https://github.com/slntopp/nocloud/wiki/Billing-Plans",
    "_blank"
  );
};
</script>

<template>
  <v-icon class="mx-3" large color="light" @click="openPlanWiki">
    mdi-information-outline
  </v-icon>
</template>

<style scoped></style>
