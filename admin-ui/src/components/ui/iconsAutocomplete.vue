<template>
  <v-autocomplete
    item-text="icon"
    item-value="icon"
    :label="label"
    :loading="isIconsLoading"
    :items="icons"
    :value="value"
    @input="emits('input:value', $event)"
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
  const iconsRes = await import("@ant-design/icons");

  icons.value = Object.entries(iconsRes).map(([name, icon]) => {
    let displayName = name;

    if (name.endsWith("Outline")) {
      displayName = name.replace("Outline", "$&d");
    } else if (name.endsWith("Fill")) {
      displayName = name.replace("Fill", "$&ed");
    }

    return { ...icon, icon: displayName, type: "ant" };
  });

  const nocloudUiIcons = require.context(
    "nocloud-ui/assets/icons",
    true,
    /\.svg/
  );

  nocloudUiIcons.keys().forEach((key) => {
    if (key.startsWith("nocloud-ui/assets/icons")) {
      icons.value.push({
        type: "nocloud",
        icon: key.split("/").slice(-1)[0].replace(".svg", ""),
      });
    }
  });

  isIconsLoading.value = false;
}
</script>
