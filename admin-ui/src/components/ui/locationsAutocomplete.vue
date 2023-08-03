<template>
  <v-autocomplete
    item-value="id"
    item-text="title"
    :multiple="!solo"
    :label="label"
    :return-object="!solo"
    :items="locations"
    :value="value"
    @input="emits('input', $event)"
  >
    <template v-slot:item="{ item }">
      <span
        >{{ item.title }} {{ item.sp ? `(${item.sp})` : "" }}
        {{ item.type ? `Service type: ${item.type}` : "" }}</span
      >
    </template>
  </v-autocomplete>
</template>

<script setup>
import { toRefs } from "vue";

const props = defineProps(["locations", "value", "solo", "label"]);
const emits = defineEmits("input");

const { value, locations, label } = toRefs(props);
</script>

<style scoped></style>
