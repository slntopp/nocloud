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
        :append-inner-icon="editIcon ? 'mdi-pencil' : undefined"
      />
    </template>
    <v-date-picker
      :disabled="disabled"
      :min="pickerMin"
      :multiple="range ? 'range' : false"
      :model-value="pickerValue"
      @update:model-value="onPickerInput"
    ></v-date-picker>
  </v-menu>
</template>

<script setup>
import { computed, toRefs } from "vue";

const props = defineProps({
  modelValue: {},
  min: {},
  label: {},
  dense: { type: Boolean, default: false },
  range: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  clearable: { type: Boolean, default: true },
  editIcon: { type: Boolean, default: false },
  placeholder: {},
});
const { modelValue: value, min, label, dense } = toRefs(props);

const emits = defineEmits(["update:modelValue"]);

// VDatePicker v3 works with Date objects; the outside contract stays 'YYYY-MM-DD' strings
const toDate = (str) => {
  const match = /^(\d{4})-(\d{2})-(\d{2})/.exec(str ?? "");
  if (!match) return undefined;
  return new Date(+match[1], +match[2] - 1, +match[3]);
};
const toDateString = (date) => {
  const pad = (n) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(
    date.getDate()
  )}`;
};

const pickerMin = computed(() => toDate(min.value) ?? min.value);
const pickerValue = computed(() => {
  if (props.range) {
    return Array.isArray(value.value)
      ? value.value.map(toDate).filter(Boolean)
      : [];
  }
  return toDate(value.value) ?? null;
});

const onPickerInput = (val) => {
  if (Array.isArray(val)) {
    // multiple="range" emits every date of the range: keep only the bounds
    const bounds = val.length > 1 ? [val[0], val[val.length - 1]] : val;
    updateInput(bounds.map(toDateString));
    return;
  }
  updateInput(val ? toDateString(val) : val);
};

const updateInput = (value) => {
  if (props.range && Array.isArray(value) && value.length === 2) {
    const [a, b] = value;
    if (a && b) {
      const ta = new Date(a).getTime();
      const tb = new Date(b).getTime();
      if (!Number.isNaN(ta) && !Number.isNaN(tb) && ta > tb) {
        emits("update:modelValue", [b, a]);
        return;
      }
    }
  }

  emits("update:modelValue", value);
};
</script>

<style scoped></style>
