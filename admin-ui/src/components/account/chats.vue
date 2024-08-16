<template>
  <plugin-iframe
    style="height: 100vh; width: 100%"
    url="/cc.ui/"
    :params="params"
  />
</template>

<script setup>
import { computed } from "vue";
import { useRoute } from "vue-router/composables";
import PluginIframe from "@/components/plugin/iframe.vue";

const route = useRoute()
const params = computed(() =>({
  filterByAccount: route.params.accountId,
  fullUrl: location.href
}))

window.addEventListener("message", ({ data, origin, source }) => {
  if (origin.includes("localhost") || !data) return;
  if (data === "ready") return;
  if (data.type === "get-user") {
    setTimeout(() => {
      source.postMessage({ type: "user-uuid", value: route.params.accountId }, "*");
    }, 300);
  }
})
</script>

<script>
export default { name: 'account-chats' }
</script>
