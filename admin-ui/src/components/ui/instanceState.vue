<template>
  <v-chip small :color="chipColor">
    {{ state }}
  </v-chip>
</template>

<script setup>
import { computed, toRefs } from "vue";

const props = defineProps(["state"]);
const { state } = toRefs(props);

const chipColor = computed(() => {
  if (!state.value) return "error";

  switch (state.value) {
    case "RUNNING":
      return "success";
    case "LCM_INIT":
    case "STOPPED":
    case "SUSPENDED":
      return "warning";
    case "UNKNOWN":
      return "error";
    case "OPERATION":
      return "yellow darken-2";
    case "PENDING":
      return "blue";
    default:
      return "blue-grey darken-2";
  }
});
</script>

<style scoped></style>
