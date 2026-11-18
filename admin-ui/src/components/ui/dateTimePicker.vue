<template>
  <v-text-field
    :model-value="placeholder || value"
    :density="dense ? 'compact' : undefined"
    :clearable="clearable"
    @update:model-value="emits('update:modelValue', $event)"
    :label="label"
    :disabled="disabled"
    :min="min"
    type="datetime-local"
    :append-inner-icon="editIcon ? 'mdi-pencil' : undefined"
  />
</template>

<script setup>
import { toRefs } from "vue";

const props = defineProps({
  modelValue: {},
  min: {},
  label: {},
  time: { type: Boolean, default: false },
  dense: { type: Boolean, default: false },
  range: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  clearable: { type: Boolean, default: false },
  editIcon: { type: Boolean, default: false },
  placeholder: {},
});
const { modelValue: value, min, label, dense } = toRefs(props);
const emits = defineEmits(["update:modelValue"]);
</script>

<style>
.v-picker__title {
  height: 88px !important;
}
</style>
