<template>
  <v-autocomplete
    item-text="displayName"
    item-value="displayName"
    :label="label"
    :loading="isIconsLoading"
    :items="icons"
    :value="value"
    @input="emits('input:value', $event)"
  >
    <template v-slot:prepend>
      <component :is="currentIcon" class="ml-3" />
    </template>
    <template v-slot:item="{ item }">
      <div class="d-flex justify-space-between" style="width: 100%">
        <span>{{ toKebabCase(item.displayName) }}</span>
        <component :is="item" style="font-size: 24px" />
      </div>
    </template>
  </v-autocomplete>
</template>

<script setup>
import { computed, onMounted, ref, toRefs } from "vue";
import { toKebabCase } from "@/functions";

const props = defineProps({
  value: { type: String, default: 'CloudOutlined' },
  label: { type: String, default: '' }
});
const emits = defineEmits(["input:value"]);

const { value } = toRefs(props);
const icons = ref([]);
const currentIcon = computed(() =>
  icons.value.find((icon) => icon.displayName === value.value)
);
const isIconsLoading = ref(false);

onMounted(fetchIcons);

async function fetchIcons() {
  isIconsLoading.value = true;
  const iconsRes = await import("@ant-design/icons-vue");

  icons.value = Object.values(iconsRes).filter((icon) =>
    typeof icon !== 'function' && icon.displayName
  );
  isIconsLoading.value = false;
}
</script>
