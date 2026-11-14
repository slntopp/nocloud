<template>
  <v-autocomplete
    item-title="icon"
    item-value="icon"
    :label="label"
    :loading="isIconsLoading"
    :items="icons"
    :model-value="value"
    @update:model-value="emits('input:value', $event)"
  >
    <template v-slot:prepend>
      <nocloud-icon :type="currentIcon?.type" :icon="currentIcon?.icon" />
    </template>
    <template v-slot:item="{ item }">
      <div class="d-flex justify-space-between" style="width: 100%">
        <span>{{ toKebabCase(item.icon) }}</span>
        <nocloud-icon
          :type="item.type"
          :icon="item.icon"
          style="font-size: 24px"
        />
      </div>
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
  value: { type: String, default: "CloudOutlined" },
  label: { type: String, default: "" },
});
const emits = defineEmits(["input:value"]);

const { value } = toRefs(props);
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
