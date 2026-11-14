<script setup>
import { ref, toRefs } from "vue";

const props = defineProps({
  modelValue: {},
  label: {},
  disabled: { type: Boolean, default: false },
  dense: { type: Boolean, default: false },
});
const { disabled, dense, modelValue: value, label } = toRefs(props);

const emit = defineEmits(["update:modelValue"]);

const items = ref([
  { value: true, title: "Enabled" },
  { value: false, title: "Disabled" },
]);
</script>

<template>
  <v-select
    :label="label"
    :model-value="value"
    :items="items"
    :disabled="disabled"
    :density="dense ? 'compact' : undefined"
    @update:model-value="emit('update:modelValue', $event)"
    item-title="title"
    item-value="value"
    clearable
  />
</template>

<style scoped></style>
