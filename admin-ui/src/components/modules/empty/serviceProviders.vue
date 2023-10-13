<template>
  <div class="sp-empty"></div>
</template>

<script>
export default {
  name: "servicesProviders-create-empty",
  props: {
    secrets: {
      type: Object,
      default: () => ({}),
    },
    vars: {
      type: Object,
      default: () => ({}),
    },
    passed: {
      type: Boolean,
      default: false,
    },
  },
  data: () => ({}),
  methods: {},
};
</script>

<style></style>
