<template>
  <v-icon @click="loginHandler"> mdi-login </v-icon>
</template>

<script setup>
import api from "@/api";
import { useStore } from "@/store";

const props = defineProps(["uuid", "instanceId", "type"]);
const store = useStore();

const loginHandler = () => {
  store
    .dispatch("auth/loginToApp", { uuid: props.uuid, type: "whmcs" })
    .then(({ token }) => {
      api.settings.get(["app"]).then((res) => {
        const win = window.open(JSON.parse(res.app).url);

        window.addEventListener('message', () => {
          win.postMessage({ token, uuid: props.instanceId, type: props.type }, "*");
        });
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
