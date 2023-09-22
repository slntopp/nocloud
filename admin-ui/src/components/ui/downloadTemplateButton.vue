<template>
  <v-btn class="mr-2" @click="downloadFile" :disabled="disabled">
    Download {{ isJson ? "JSON" : "YAML" }}
  </v-btn>
</template>

<script setup>
import { downloadJSONFile, downloadYAMLFile } from "@/functions";
import { computed, toRefs } from "vue";

const props = defineProps(["name", "template", "type", "disabled"]);
const { name, template, type, disabled } = toRefs(props);

const isJson = computed(() => {
  return type.value === "JSON";
});

const downloadFile = () => {
  const params = [template.value, name.value];
  if (isJson.value) {
    downloadJSONFile(...params);
  } else {
    downloadYAMLFile(...params);
  }
};
</script>
