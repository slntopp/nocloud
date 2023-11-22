<template>
  <plugin-iframe
    style="height: 100vh; width: 100%"
    url="/cc.ui/"
    :params="{ filterByAccount: $route.params.accountId }"
  />
</template>

<script setup>
import PluginIframe from "@/components/plugin/iframe.vue";
</script>

<script>
export default { name: 'account-chats' }
</script>
