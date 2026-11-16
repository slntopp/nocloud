<template>
  <v-btn @click="trigerInput">
    <v-file-input
        id="fileInput"
      hidden="hidden"
     hide-details
      prepend-icon=""
      hide-input
      :accept="'.'+accept"
      @update:model-value="$emit('file',$event)"
    />
    <slot></slot>
  </v-btn>
</template>

<script>
export default {
  name: "uploadFileButton",
  props: ["accept"],
  emits: ["file"],
  methods: {
    trigerInput() {
      document.getElementById('fileInput').click()
    },
  },
};
</script>

<style scoped></style>
