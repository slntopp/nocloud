<template>
  <v-menu
    :close-on-content-click="false"
    transition="scale-transition"
    min-width="auto"
  >
    <template v-slot:activator="{ on, attrs }">
      <v-text-field
        v-bind="attrs"
        v-on="on"
        prepend-inner-icon="mdi-calendar"
        :value="placeholder || value"
        readonly
        :dense="dense"
        clearable
        @input="emits('input', $event)"
        :label="label"
        :disabled="disabled"
      />
    </template>
    <v-date-picker
      scrollable
      :disabled="disabled"
      :min="min"
      :range="range"
      :value="value"
      @input="emits('input', $event)"
    ></v-date-picker>
  </v-menu>
</template>

<script setup>
import { toRefs } from "vue";

const props = defineProps({
  value: {},
  min: {},
  label: {},
  clearable: { type: Boolean, default: false },
  dense: { type: Boolean, default: false },
  range: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  placeholder: {},
});
const { value, min, label, dense } = toRefs(props);

const emits = defineEmits(["input"]);
</script>

<style scoped></style>
