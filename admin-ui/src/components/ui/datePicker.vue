<template>
  <v-menu
    :close-on-content-click="false"
    transition="scale-transition"
    min-width="auto"
  >
    <template v-slot:activator="{ props }">
      <v-text-field
       
        v-bind="props"
        :prepend-inner-icon="!editIcon ? 'mdi-calendar' : undefined"
        :model-value="placeholder || value"
        readonly
        :density="dense ? 'compact' : undefined"
        :clearable="clearable"
        @update:model-value="updateInput"
        :label="label"
        :disabled="disabled"
        :append-icon="editIcon ? 'mdi-pencil' : undefined"
      />
    </template>
    <v-date-picker
      scrollable
      :disabled="disabled"
      :min="min"
      :range="range"
      :value="value"
      @input="updateInput"
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
  range: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  clearable: { type: Boolean, default: true },
  editIcon: { type: Boolean, default: false },
  placeholder: {},
});
const { value, min, label, dense } = toRefs(props);

const emits = defineEmits(["input"]);

const updateInput = (value) => {
  if (props.range && Array.isArray(value) && value.length === 2) {
    const [a, b] = value;
    if (a && b) {
      const ta = new Date(a).getTime();
      const tb = new Date(b).getTime();
      if (!Number.isNaN(ta) && !Number.isNaN(tb) && ta > tb) {
        emits("input", [b, a]);
        return;
      }
    }
  }

  emits("input", value);
};
</script>

<style scoped></style>
