<template>
  <v-icon class="ml-2" v-if="isMdi">{{ `mdi-${iconName}` }}</v-icon>
  <ant-icon v-else-if="isAnt" class="ml-2" :name="iconName" />
  <img
    v-else-if="iconName && !loadedError"
    width="22px"
    height="22px"
    :src="iconName"
  />
  <ant-icon v-else class="ml-2" :name="icon" />
</template>

<script setup>
import AntIcon from "@/components/ui/antIcon.vue";
import { computed, toRefs, watch, ref, onMounted } from "vue";

const props = defineProps({
  type: { type: String, default: "mdi" },
  title: { type: String, default: "none" },
  icon: { type: String, default: "" },
});
const { type, icon } = toRefs(props);

const loadedIcon = ref();
const loadedError = ref(false);

onMounted(() => {
  if (isNocloud.value) {
    tryLoad();
  }
});

const isMdi = computed(() => type.value === "mdi");
const isAnt = computed(() => type.value === "ant");
const isNocloud = computed(() => type.value === "nocloud");

const iconName = computed(() => {
  if (isNocloud.value) {
    return loadedIcon.value?.default;
  }

  return icon.value;
});

const tryLoad = async () => {
  loadedError.value = false;
  try {
    const loaded = await import(`nocloud-ui/assets/icons/${icon.value}.svg`);

    loadedIcon.value = loaded;
  } catch {
    loadedError.value = true;
  }
};

watch(isNocloud, async (val) => {
  if (val) {
    tryLoad();
  }
});
</script>

<style scoped></style>
