<template>
  <v-btn
    small
    id="button"
    class="mx-2"
    color="background"
    @click="$emit('add')"
    :disabled="disabled"
  >
    <v-icon dark> mdi-plus-circle-outline </v-icon>
    add instance
  </v-btn>
</template>

<script>
export default {
  name: "addInstance",
  emits: ["add"],
  props: ["disabled"],
};
</script>

<style scoped></style>
