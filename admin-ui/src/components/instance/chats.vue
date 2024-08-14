<template>
  <plugin-iframe
    v-if="accountId"
    style="height: 100vh; width: 100%"
    url="/cc.ui/"
    :params="params"
  />
</template>

<script setup>
import PluginIframe from "@/components/plugin/iframe.vue";
import { computed, toRefs } from "vue";

const props = defineProps(["template", "account"]);
const { account } = toRefs(props);

const accountId = computed(() => {
  return account.value?.uuid;
});

const params = computed(() =>({
  filterByAccount: accountId.value,
  fullUrl: location.href
}));

window.addEventListener("message", ({ data, origin, source }) => {
  if (origin.includes("localhost") || !data) return;
  if (data === "ready") return;
  if (data.type === "get-user") {
    setTimeout(() => {
      source.postMessage(
        { type: "user-uuid", value: accountId.value },
        "*"
      );
    }, 300);
  }
});
</script>

<script>
export default { name: "instance-chats" };
</script>
