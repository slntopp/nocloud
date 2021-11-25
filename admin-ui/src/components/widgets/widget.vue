<template>
  <v-card
    elevation="0"
    color="background-light"
    :loading="loading"
    class="flex-shrink-0 flex-grow-1 rounded-lg"
    max-width="520"
    min-width="450"
  >
    <template slot="progress">
      <v-progress-linear
        color="primary"
        height="5"
        indeterminate
				absolute
      ></v-progress-linear>
    </template>
    <v-card-title>{{title}}</v-card-title>

    <v-divider class="mx-4"></v-divider>

    <v-card-text>
      <slot></slot>
    </v-card-text>

  </v-card>
</template>

<script>
export default {
  name: "widget",
  props: {
    title: String,
    loading: {
      type: Boolean,
      default: false
    }
  }
}
</script>

<style>

</style>