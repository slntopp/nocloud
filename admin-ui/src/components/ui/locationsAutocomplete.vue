<template>
  <v-autocomplete
    item-value="id"
    item-title="title"
    :multiple="!solo"
    :label="label"
    :return-object="!solo"
    :items="locations"
    :model-value="value"
    @update:model-value="emits('update:modelValue', $event)"
  >
    <template v-slot:item="{ internalItem: item, props: itemProps }">
      <v-list-item v-bind="itemProps" :title="undefined">
        <span
          >{{ item.title }} {{ item.raw.sp ? `(${item.raw.sp})` : "" }}
          {{ item.raw.type ? `Service type: ${item.raw.type}` : "" }}</span
        >
      </v-list-item>
    </template>
  </v-autocomplete>
</template>

<script setup>
import { toRefs } from "vue";

const props = defineProps(["locations", "modelValue", "solo", "label"]);
const emits = defineEmits(["update:modelValue"]);

const { modelValue: value, locations, label } = toRefs(props);
</script>

<style scoped></style>
