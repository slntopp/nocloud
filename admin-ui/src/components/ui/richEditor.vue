<template>
  <div class="editor">
    <jodit-editor :config="config" :value="value" @input="onInput" />
  </div>
</template>

<script setup>
import "jodit/build/jodit.min.css";
import { JoditEditor } from "jodit-vue";
import { computed, toRefs } from "vue";
import { useStore } from "@/store";

const props = defineProps({
  value: {},
  disabled: { type: Boolean, default: false },
});
const emit = defineEmits(["input"]);

const store = useStore();

const onInput = (e) => {
  emit("input", e);
};

const { value, disabled } = toRefs(props);

const theme = computed(() => store.getters["app/theme"]);

const config = computed(() => ({
  theme: theme.value,
  disabled: disabled.value,
  minHeight: 400,
}));
</script>
