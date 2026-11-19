<template>
  <div
    class="plugin-page h-100 w-100"
    :class="{ 'pa-4': viewport > 900, 'plugin-page--keyboard': keyboardOpen }"
    :style="frameStyle"
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
const vvTop = ref(0);
const vvHeight = ref(0);
const keyboardOpen = ref(false);

function syncVisualViewport() {
  const vv = window.visualViewport;
  const isChats = String(url.value || "").includes("cc.ui");
  if (!vv || viewport.value > 900 || !isChats) {
    keyboardOpen.value = false;
    return;
  }
  vvTop.value = vv.offsetTop;
  vvHeight.value = vv.height;
  keyboardOpen.value = window.innerHeight - vv.height > 150;
}

const frameStyle = computed(() => {
  if (isMaximized.value) return { width: "100vw", height: "100vh" };
  if (!keyboardOpen.value) return undefined;
  return {
    top: `${vvTop.value}px`,
    height: `${vvHeight.value}px`,
  };
});

onMounted(() => {
  window.addEventListener("resize", onResize);
  window.visualViewport?.addEventListener("resize", syncVisualViewport);
  window.visualViewport?.addEventListener("scroll", syncVisualViewport);
});
onUnmounted(() => {
  window.removeEventListener("resize", onResize);
  window.visualViewport?.removeEventListener("resize", syncVisualViewport);
  window.visualViewport?.removeEventListener("scroll", syncVisualViewport);
});
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

.plugin-page--keyboard {
  position: fixed;
  left: 0;
  right: 0;
  z-index: 20;
  width: 100%;
  padding: 0 !important;
}

.w-100 {
  width: 100%;
}
</style>
