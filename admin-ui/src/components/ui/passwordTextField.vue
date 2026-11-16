<template>
  <v-text-field
    :readonly="readonly"
    :label="label"
    :type="isVisible ? 'text' : 'password'"
    :model-value="value"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <template v-slot:append-inner>
      <v-icon @click="isVisible = !isVisible">{{
        isVisible ? "mdi-eye" : "mdi-eye-off"
      }}</v-icon>
      <v-icon v-if="copy" class="ml-1" @click="addToClipboard(value)"
        >mdi-content-copy</v-icon
      >
    </template>
  </v-text-field>
</template>

<script setup>
import { ref, toRefs } from "vue";
import { addToClipboard } from "@/functions";

const props = defineProps({
  modelValue: {
    type: String,
    default: "",
  },
  label: { type: String, default: "Password" },
  copy: { type: Boolean, default: false },
  readonly: { type: Boolean, default: true },
});
const emit = defineEmits(["update:modelValue"]);
const { modelValue: value, label } = toRefs(props);

const isVisible = ref(false);
</script>
