<template>
  <div class="editor">
    <jodit-editor :config="config" :value="value" @input="onInput" />
  </div>
</template>

<script setup>
import "jodit/build/jodit.min.css";
import { JoditEditor } from "jodit-vue";
import { computed, toRefs } from "vue";

const props = defineProps({
  value: {},
  disabled: { type: Boolean, default: false },
});
const emit = defineEmits(["input"]);

const onInput = (e) => {
  emit("input", e);
};

const { value, disabled } = toRefs(props);

const config = computed(() => ({
  theme: "dark",
  disabled: disabled.value,
  minHeight: 400,
}));
</script>
