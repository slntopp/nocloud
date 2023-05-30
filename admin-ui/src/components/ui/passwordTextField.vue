<template>
  <v-text-field
    readonly
    :label="label"
    :type="isVisible ? 'text' : 'password'"
    :value="value"
    :append-icon="isVisible ? 'mdi-eye' : 'mdi-eye-off'"
    @click:append="isVisible = !isVisible"
  />
</template>

<script setup>
import { defineProps, ref, toRefs } from "vue";

const props = defineProps({
  value: {
    type: String,
    default: "",
  },
  label: { type: String, default: "Pass" },
});
const { value,label } = toRefs(props);

const isVisible = ref(false);
</script>
