<template>
  <v-tooltip bottom>
    <template v-slot:activator="{ on, attrs }">
      <div class="d-inline-block" v-bind="attrs" v-on="on">
        <slot></slot>
      </div>
    </template>
    <v-list-item-title> {{ (hint || '').toUpperCase() }}</v-list-item-title>
  </v-tooltip>
</template>

<script setup>
import { toRefs } from "vue";

const props = defineProps(["hint"]);
const { hint } = toRefs(props);
</script>

<style scoped></style>
