<script>
export default {
  name: "PluginIframe",
};
</script>

<script setup>
import { Buffer } from "buffer";
import { useStore } from "@/store";
import { computed, toRefs } from "vue";
import { useRoute } from "vue-router/composables";

const props = defineProps(["url", "params"]);
const { url, params } = toRefs(props);

const store = useStore();
const route = useRoute();

const theme = computed(() => store.getters["app/theme"]);
const src = computed(() => {
  const { title } = store.getters["auth/userdata"];
  const { token } = store.state.auth;
  const fullParams = JSON.stringify({
    title,
    token,
    api: location.host,
    theme: theme.value,
    params: params.value,
    fullscrean: route.query["fullscrean"] === "true",
  });

  return `${url.value}?a=${Buffer.from(
    unescape(encodeURIComponent(fullParams))
  ).toString("base64")}`;
});
</script>

<template>
  <iframe frameborder="0" :src="src"></iframe>
</template>

<style scoped></style>
