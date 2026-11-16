<template>
  <div class="editor" v-if="isVisible">
    <jodit-editor :config="config" :model-value="value ?? ''" @update:model-value="onInput" />
  </div>
</template>

<script setup>
import "jodit/build/jodit.min.css";
import { JoditEditor } from "jodit-vue";
import { computed, toRefs, watch, ref } from "vue";
import { useStore } from "@/store";

const props = defineProps({
  modelValue: {},
  disabled: { type: Boolean, default: false },
});
const emit = defineEmits(["update:modelValue"]);

const store = useStore();

const isVisible = ref(true);

const onInput = (e) => {
  emit("update:modelValue", e);
};

const { modelValue: value, disabled } = toRefs(props);

const theme = computed(() => store.getters["app/theme"]);

const config = computed(() => ({
  theme: theme.value,
  disabled: disabled.value,
  minHeight: 400,
}));

watch(
  config,
  () => {
    isVisible.value = false;

    setTimeout(() => (isVisible.value = true), 0);
  },
  { deep: true }
);
</script>
