<template>
  <v-btn class="mr-2" @click="downloadFile" :disabled="disabled" :small="small">
    {{ title || `Download ${fullType}` }}
  </v-btn>
</template>

<script setup>
import { downloadJSONFile, downloadYAMLFile } from "@/functions";
import { computed, toRefs } from "vue";

const props = defineProps({
  name: {},
  template: {},
  type: {},
  disabled: {},
  title: {},
  small: { default: false, type: Boolean },
});
const { name, template, type, disabled, small, title } = toRefs(props);

const emit = defineEmits("click:xlsx");

const isJson = computed(() => {
  return type.value === "JSON";
});
const fullType = computed(() => {
  return type.value?.toUpperCase();
});

const downloadFile = () => {
  const params = [template.value, name.value];

  if (type.value === "xlsx".toUpperCase()) {
    return emit("click:xlsx");
  }

  if (isJson.value) {
    downloadJSONFile(...params);
  } else {
    downloadYAMLFile(...params);
  }
};
</script>
