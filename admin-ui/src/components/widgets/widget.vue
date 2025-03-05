<template>
  <v-card
    elevation="0"
    max-width="520"
    color="background-light"
    class="flex-grow-1 rounded-lg"
    :min-width="width >= 450 ? 350 : 200"
    :height="410"
    :style="{ width: width < 1375 ? '100%' : 'auto' }"
    :loading="loading"
  >
    <template slot="progress">
      <v-progress-linear
        color="primary"
        height="5"
        indeterminate
        absolute
      ></v-progress-linear>
    </template>

    <div class="d-flex justify-space-between align-center px-3">
      <v-card-title>{{ title }}</v-card-title>
      <v-btn v-if="more" @click="$router.push(more)" small text> More </v-btn>
    </div>

    <v-divider class="mx-4"></v-divider>

    <v-card-text class="pt-0">
      <slot></slot>
    </v-card-text>
  </v-card>
</template>

<script>
export default {
  name: "widget-origin",
  props: {
    title: String,
    more: { type: Object },
    loading: {
      type: Boolean,
      default: false,
    },
  },
  computed: {
    width() {
      return document.documentElement.clientWidth;
    },
  },
};
</script>

<style></style>
