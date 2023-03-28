<template>
  <v-btn
    small
    id="button"
    class="mx-2"
    color="background"
    :disabled="disabled"
    @click="$emit('click')"
  >
    <v-icon dark> mdi-plus-circle-outline </v-icon>
    add instance
  </v-btn>
</template>

<script>
export default {
  name: "add-instance-btn",
  props: ["disabled"],
};
</script>

<style scoped></style>
