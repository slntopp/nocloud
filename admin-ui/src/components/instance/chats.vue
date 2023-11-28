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
import { useStore } from "@/store";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const accountId = computed(() => {
  const namespace = store.getters["namespaces/all"]?.find(
    (n) => n.uuid === template.value?.access.namespace
  );
  const account = store.getters["accounts/all"].find(
    (a) => a.uuid === namespace?.access.namespace
  );
  return account?.uuid;
});
</script>

<script>
export default { name: "instance-chats" };
</script>
