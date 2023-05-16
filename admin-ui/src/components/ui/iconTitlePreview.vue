<template>
  <div class="d-flex justify-space-between" style="width: 100%">
    <span>{{ instance?.title ?? title }}</span>
    <v-icon class="ml-2" v-if="isMdi">{{ `mdi-${icon}` }}</v-icon>
    <ant-icon
      class="ml-2"
      v-else-if="icon || instance?.icon"
      :name="toPascalCase(instance?.icon ?? icon)"
    />
  </div>
</template>

<script>
import AntIcon from "@/components/ui/antIcon.vue";
import { toPascalCase } from "@/functions.js";

export default {
  name: "iconTitlePreview",
  components: { AntIcon },
  props: {
    isMdi: { type: Boolean, default: true },
    instance: { type: Object, default: null },
    title: { type: String, default: 'none' },
    icon: { type: String, default: '' }
  },
  methods: { toPascalCase }
};
</script>

<style scoped></style>
