<script setup>
import { ref, toRefs } from "vue";

const props = defineProps({
  value: {},
  label: {},
  disabled: { type: Boolean, default: false },
  dense: { type: Boolean, default: false },
});
const { disabled, dense, value, label } = toRefs(props);

const emit = defineEmits(["input"]);

const items = ref([
  { value: true, title: "Enabled" },
  { value: false, title: "Disabled" },
]);
</script>

<template>
  <v-select
    :label="label"
    :value="value"
    :items="items"
    :disabled="disabled"
    :dense="dense"
    @input="emit('input', $event)"
    item-text="title"
    item-value="value"
    clearable
  />
</template>

<style scoped></style>
