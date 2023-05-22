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
        const url = JSON.parse(res["app"]).url;
        const win = window.open(
          props.instanceId ? `${url}/#/cloud/${props.instanceId}` : url
        );
        console.log(props.instanceId, `${url}/#/cloud/${props.instanceId}`);

        setTimeout(() => {
          win.postMessage(token, url);
        }, 100);
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
