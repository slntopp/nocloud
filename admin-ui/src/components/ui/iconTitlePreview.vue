<template>
  <div class="d-flex justify-space-between" style="width: 100%">
    <span>{{ title }}</span>
    <v-icon class="ma-auto ml-2" v-if="isMdi">{{ `mdi-${icon}` }}</v-icon>
    <ant-icon class="ma-auto ml-2" v-else :name="icon" />
  </div>
</template>

<script>
import AntIcon from "@/components/ui/antIcon.vue";

export default {
  name: "iconTitlePreview",
  components: { AntIcon },
  props: { isMdi: { type: Boolean, default: true }, title: {}, icon: {} },
};
</script>

<style scoped></style>
