<template>
  <component :is="icon" v-if="icon" style="font-size: 24px" />
</template>

<script setup>
import { shallowRef, watch } from "vue";
import { loadAntIcons } from "@/components/ui/nocloudIcons";

const props = defineProps({ name: { type: String, default: "" } });
const icon = shallowRef(null);

watch(
  () => props.name,
  async (name) => {
    if (!name) {
      icon.value = null;
      return;
    }
    const icons = await loadAntIcons();
    icon.value = icons[name] ?? null;
  },
  { immediate: true },
);
</script>
