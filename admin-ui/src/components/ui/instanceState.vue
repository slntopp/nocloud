<template>
  <v-chip v-if="!onlyColor" :small="small" :color="chipColor">
    {{ state }} {{ isDetached ? "(Hided)" : "" }}
    {{ isFreezed ? "(Freezed)" : "" }}
  </v-chip>
  <v-btn
    v-else
    fab
    :color="chipColor"
    x-small
    style="width: 15px; height: 15px"
  >
  </v-btn>
</template>

<script setup>
import { computed, toRefs } from "vue";
import { getState } from "@/functions";

const props = defineProps({
  template: {},
  small: {},
  onlyColor: { type: Boolean, default: false },
});
const { template, onlyColor } = toRefs(props);

const state = computed(() => {
  return getState(template.value);
});

const isDetached = computed(() => {
  return template.value?.status.toLowerCase() === "detached";
});

const isFreezed = computed(() => {
  return template.value?.data?.freeze;
});

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
    case "ERROR":
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
