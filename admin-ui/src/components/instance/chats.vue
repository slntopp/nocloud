<template>
  <plugin-iframe
    v-if="accountId"
    style="height: 100vh; width: 100%"
    url="/cc.ui/"
    :params="{ filterByAccount: accountId }"
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
</script>

<script>
export default { name: "instance-chats" };
</script>
