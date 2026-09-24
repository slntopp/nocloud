<template>
  <plugin-iframe
    v-if="url"
    class="campaigns__frame"
    :url="url"
    :params="params"
  />
</template>

<script setup>
import { computed } from "vue";
import { useRoute } from "vue-router/composables";
import { useStore } from "@/store";
import PluginIframe from "@/components/plugin/iframe.vue";

// The campaigns plugin's journal, filtered to this customer: everything that
// has gone out to them, the tickets its chains opened and the letters nocloud
// asked WHMCS to send, in one list. Built the way the helpdesk tab is — the
// plugin in an iframe, told who the page is about.
//
// The address is not hardcoded the way the chats one is: plugins are listed in
// nocloud's own settings and this one's url ends in /campaigns.ui/. Not
// installed, no tab — see AccountPage.
const route = useRoute();
const store = useStore();

const url = computed(
  () =>
    (store.getters["plugins/all"] || []).find((p) =>
      (p.url || "").includes("campaigns.ui")
    )?.url
);

// redirect is how the plugin is told where to open, the same way the chats one
// is; embed tells it that it is inside somebody else's page and should leave
// its own header, language picker and customer filter out.
const params = computed(() => ({
  embed: true,
  redirect: `dashboard/sends?account=${route.params.accountId}`,
  fullUrl: location.href,
}));
</script>

<script>
// Named the way the helpdesk tab names itself, for the same linter.
export default { name: "account-campaigns" };
</script>

<style scoped>
/* The tab, not the window: the page keeps its own header and tab bar above,
   and only the journal inside scrolls. */
.campaigns__frame {
  width: 100%;
  height: calc(100vh - 260px);
  min-height: 480px;
}
</style>
