<template>
  <v-text-field
    :readonly="readonly"
    :label="label"
    :type="isVisible ? 'text' : 'password'"
    :value="value"
    @input="$emit('input', $event)"
  >
    <template v-slot:append>
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
import { defineProps, ref, toRefs } from "vue";
import { addToClipboard } from "@/functions";

const props = defineProps({
  value: {
    type: String,
    default: "",
  },
  label: { type: String, default: "Password" },
  copy: { type: Boolean, default: false },
  readonly: { type: Boolean, default: true },
});
const { value, label } = toRefs(props);

const isVisible = ref(false);
</script>
