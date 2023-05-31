<template>
  <div class="sp-virtual">
  </div>
</template>

<script>


export default {
  name: "servicesProviders-create-virtual",
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
  data: () => ({
  }),
  methods: {
  },
};
</script>

<style></style>
