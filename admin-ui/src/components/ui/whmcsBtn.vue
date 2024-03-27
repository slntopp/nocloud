<template>
  <v-btn icon small>
    <v-icon @click="goToWhmcs"> mdi-alpha-w-box </v-icon>
  </v-btn>
</template>

<script setup>
import { toRefs } from "vue";
import { useStore } from "@/store";

const props = defineProps(["account"]);
const { account } = toRefs(props);

const store = useStore();

const goToWhmcs = () => {
  if (!account.value.data?.whmcs_id) {
    return;
  }
  const url = /https:\/\/(.+?\.?\/)/.exec(
    store.getters["settings/whmcsApi"]
  )[0];
  window.open(
    `${url}admin/clientssummary.php?userid=${account.value.data?.whmcs_id}`,
    "_blank"
  );
};
</script>
