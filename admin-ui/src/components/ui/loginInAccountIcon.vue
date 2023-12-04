<template>
  <v-icon @click="onLogin"> mdi-login </v-icon>
</template>

<script setup>
import useLoginInClient from "@/hooks/useLoginInClient";

const props = defineProps(["uuid", "instanceId", "type"]);

const { loginHandler } = useLoginInClient();

const onLogin = () => {
  loginHandler({
    accountUuid: props.uuid,
    type: props.type,
    instanceId: props.instanceId,
  });
};
</script>

<style scoped></style>
