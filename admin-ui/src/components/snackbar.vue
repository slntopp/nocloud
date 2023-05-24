<template>
  <v-snackbar v-model="visibility" :timeout="timeout" :color="color">
    {{ message }}
    <template v-if="route && Object.keys(route).length > 0">
      <router-link :to="route"> Look up. </router-link>
    </template>

    <template v-slot:action="{ attrs }">
      <v-btn
        :color="buttonColor"
        text
        v-bind="attrs"
        @click="visibility = false"
      >
        Close
      </v-btn>
    </template>
  </v-snackbar>
</template>

<script>
import { mapGetters } from "vuex";

export default {
  name:'app-snackbar',
  computed: {
    ...mapGetters("snackbar", [
      "timeout",
      "color",
      "message",
      "route",
      "buttonColor",
    ]),
    visibility: {
      get() {
        return this.$store.getters["snackbar/visibility"];
      },
      set(val) {
        this.$store.commit("snackbar/setVisibility", val);
      },
    },
  },
};
</script>