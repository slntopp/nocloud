<template>
  <div
    class="plugin-page h-100 w-100"
    :class="{ 'pa-4': viewport > 900 }"
    :style="isMaximized ? { width: '100vw', height: '100vh' } : {}"
  >
    <plugin-iframe
      v-if="$route.query.url"
      ref="plugin"
      class="h-100 w-100"
      :url="url"
      :params="params"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useStore } from "@/store";
import PluginIframe from "@/components/plugin/iframe.vue";

const route = useRoute();
const store = useStore();

const plugin = ref();

const url = computed(() => route.query.url);
const viewport = ref(window.innerWidth);
const onResize = () => (viewport.value = window.innerWidth);
onMounted(() => window.addEventListener("resize", onResize));
onUnmounted(() => window.removeEventListener("resize", onResize));
const isMaximized = computed(() => {
  const isChats = String(url.value || "").includes("cc.ui");
  if (isChats && viewport.value <= 900) return false;
  return route.query.fullscreen === "true";
});
const params = computed(() => {
  const result = {
    fullUrl: location.href,
  };

  if (route.query.redirect) {
    result.redirect = route.query.redirect;
  }

  if (route.query.chat) {
    result.redirect = `dashboard/${route.query.chat}`;
  }

  return result;
});

watch(
  () => store.getters["app/chatClicks"],
  () => {
    plugin.value.$el.contentWindow.postMessage({ type: "start-page" }, "*");
  }
);
</script>

<script>
export default { name: "plugin-view" };
</script>

<style>
.h-100 {
  height: 100%;
  box-sizing: border-box;
  overflow: hidden;
}

.plugin-page {
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.plugin-page > * {
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
}

.w-100 {
  width: 100%;
}
</style>
