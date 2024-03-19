<template>
  <div class="pa-4 h-100 w-100">
    <plugin-iframe
      :redirect="redirect"
      class="h-100 w-100"
      v-if="$route.query.url"
      :url="url"
      :params="params"
    />
  </div>
</template>

<script>
import PluginIframe from "@/components/plugin/iframe.vue";

export default {
  name: "plugin-view",
  components: { PluginIframe },
  computed: {
    url() {
      return this.$route.params.url || this.$route.query.url;
    },
    params() {
      return this.$route.params.params;
    },
  },
};
</script>

<style>
.h-100 {
  height: 100%;
}

.w-100 {
  width: 100%;
}
</style>
