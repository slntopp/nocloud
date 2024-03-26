<template>
  <div class="pa-4 h-100 w-100">
    <plugin-iframe
      v-if="$route.query.url"
      ref="plugin"
      class="h-100 w-100"
      :redirect="redirect"
      :url="url"
      :params="params"
    />
  </div>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router/composables";
import { useStore } from "@/store";
import PluginIframe from "@/components/plugin/iframe.vue";

const route = useRoute()
const store = useStore()

const plugin = ref()
const url = computed(() =>
  route.params.url || route.query.url
)
const params = computed(() =>
  route.params.params
)

watch(() => store.getters["app/chatClicks"], () => {
  plugin.value.$el.contentWindow.postMessage({ type: "start-page" }, "*");
})
</script>

<script>
export default { name: "plugin-view" }
</script>

<style>
.h-100 {
  height: 100%;
}

.w-100 {
  width: 100%;
}
</style>
