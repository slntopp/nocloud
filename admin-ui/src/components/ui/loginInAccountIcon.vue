<template>
  <v-icon @click="loginHandler"> mdi-login </v-icon>
</template>

<script setup>
import api from "@/api";
import { useStore } from "@/store";

const props = defineProps(["uuid", "instanceId"]);
const store = useStore();

const loginHandler = () => {
  store
    .dispatch("auth/loginToApp", { uuid: props.uuid, type: "whmcs" })
    .then(({ token }) => {
      api.settings.get(["app"]).then((res) => {
        const win = window.open(JSON.parse(res.app).url);

        setTimeout(() => {
          win.postMessage({ token, uuid: props.instanceId }, "*");
        }, 300);
      });
    })
    .catch((e) => {
      store.commit("snackbar/showSnackbarError", {
        message: e.response?.data?.message || "Error during login",
      });
    });
};
</script>

<style scoped></style>
