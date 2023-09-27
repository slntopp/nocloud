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
        :value="value"
        readonly
        :dense="dense"
        :label="label"
      />
    </template>
    <v-date-picker
      scrollable
      :min="min"
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
  dense: { type: Boolean, default: false },
});
const { value, min, label, dense } = toRefs(props);

const emits = defineEmits(["input"]);
</script>

<style scoped></style>
