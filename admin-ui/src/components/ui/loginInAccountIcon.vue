<template>
  <v-icon @click="loginHandler"> mdi-login </v-icon>
</template>

<script setup>
import { defineProps } from "vue";
import api from "@/api";
import { useStore } from "@/store";

const props = defineProps(["uuid"]);

const store = useStore();

const loginHandler = () => {
  store
    .dispatch("auth/loginToApp", { uuid: props.uuid, type: "whmcs" })
    .then(({ token }) => {
      api.settings.get(["app"]).then((res) => {
        const url = JSON.parse(res["app"]).url;
        const win = window.open(url);
        setTimeout(() => {
          win.postMessage(token, url);
        }, 100);
      });
    });
};
</script>

<style scoped></style>
