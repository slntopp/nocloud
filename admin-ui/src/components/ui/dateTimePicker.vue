<template>
  <v-text-field
    :prepend-inner-icon="!editIcon ? 'mdi-calendar' : undefined"
    :value="placeholder || value"
    :dense="dense"
    :clearable="clearable"
    @input="emits('input', $event)"
    :label="label"
    :disabled="disabled"
    :min="min"
    type="datetime-local"
    :append-icon="editIcon ? 'mdi-pencil' : undefined"
  />
</template>

<script setup>
import { toRefs } from "vue";

const props = defineProps({
  value: {},
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
const { value, min, label, dense } = toRefs(props);
const emits = defineEmits(["input"]);
</script>

<style>
.v-picker__title {
  height: 88px !important;
}
</style>
