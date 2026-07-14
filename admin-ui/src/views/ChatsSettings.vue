<template>
  <div class="pa-4 h-100" style="height: 100%;width: 100%;">
    <plugin-iframe
      style="height: 100%;width: 100%;"
      url="/cc.ui/"
      :params="{ redirect: 'settings' }"
    />
  </div>
</template>

<script setup>
import PluginIframe from "@/components/plugin/iframe.vue";
</script>

<script>
export default { name: "ChatsSettingsView" };
</script>
