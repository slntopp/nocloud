<template>
  <v-autocomplete
    item-title="icon"
    item-value="icon"
    :label="label"
    :loading="isIconsLoading"
    :items="icons"
    :model-value="value"
    @update:model-value="emits('update:modelValue', $event)"
  >
    <template v-slot:prepend>
      <nocloud-icon :type="currentIcon?.type" :icon="currentIcon?.icon" />
    </template>
    <template v-slot:item="{ internalItem: item, props: itemProps }">
      <v-list-item v-bind="itemProps" :title="toKebabCase(item.raw.icon)">
        <template v-slot:append>
          <nocloud-icon
            :type="item.raw.type"
            :icon="item.raw.icon"
            style="font-size: 24px"
          />
        </template>
      </v-list-item>
    </template>
  </v-autocomplete>
</template>

<script setup>
import { computed, onMounted, ref, toRefs } from "vue";
import { toKebabCase } from "@/functions";
import NocloudIcon from "@/components/ui/nocloudIcon.vue";
import {
  isAntIconName,
  loadAntIcons,
  nocloudIconNames,
} from "@/components/ui/nocloudIcons";

const props = defineProps({
  modelValue: { type: String, default: "CloudOutlined" },
  label: { type: String, default: "" },
});
const emits = defineEmits(["update:modelValue"]);

const { modelValue: value } = toRefs(props);
const icons = ref([]);
const currentIcon = computed(() =>
  icons.value.find((icon) => icon.icon === value.value)
);
const isIconsLoading = ref(false);

onMounted(fetchIcons);

async function fetchIcons() {
  isIconsLoading.value = true;
  const antIcons = await loadAntIcons();

  icons.value = [
    ...Object.keys(antIcons)
      .filter(isAntIconName)
      .map((icon) => ({ icon, type: "ant" })),
    ...nocloudIconNames.map((icon) => ({ icon, type: "nocloud" })),
  ];

  isIconsLoading.value = false;
}
</script>
