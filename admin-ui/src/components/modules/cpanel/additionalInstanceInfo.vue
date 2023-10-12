<template>
  <v-row>
    <v-col cols="2">
      <v-text-field readonly :value="template.config.domain" label="Domain" />
    </v-col>
    <v-col cols="2">
      <v-text-field readonly :value="template.config.email" label="Email" />
    </v-col>
  </v-row>
</template>

<script setup>
import { toRefs, defineProps } from "vue";

const props = defineProps(["template", "sp"]);

const { template } = toRefs(props);
</script>

<style scoped></style>
