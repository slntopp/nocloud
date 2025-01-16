<template>
  <div class="d-flex justify-space-between" style="width: 100%">
    <span>{{ title }}</span>
    <nocloud-icon :color="color" :type="type" :icon="icon" />
  </div>
</template>

<script>
import NocloudIcon from "@/components/ui/nocloudIcon.vue";

export default {
  name: "iconTitlePreview",
  components: { NocloudIcon },
  props: {
    type: {},
    title: { type: String, default: "none" },
    icon: { type: String, default: "" },
    color: { type: String, default: "" },
  },
};
</script>

<style scoped></style>
