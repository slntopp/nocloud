<template>
  <div class="pa-4 h-100 w-100">
    <iframe
      class="h-100 w-100"
      frameborder="0"
      v-if="$route.query.url"
      :src="src"
    ></iframe>
  </div>
</template>

<script>
import { Buffer } from "buffer";

export default {
  name: "plugin-view",
  computed: {
    src() {
      const { title } = this.$store.getters["auth/userdata"];
      const { token } = this.$store.state.auth;
      const url = this.$route.params.url || this.$route.query.url;
      const params = JSON.stringify({ title, token, api: location.host });
      return `${url}?a=${Buffer.from(params).toString("base64")}`;
    },
  },
};
</script>

<style>
.h-100 {
  height: 100%;
}

.w-100 {
  width: 100%;
}
</style>
