<template>
  <v-menu offset-y transition="slide-y-transition">
    <template v-slot:activator="{ on, attrs }">
      <v-btn class="mx-2" fab color="background-light" v-bind="attrs" v-on="on">
        <v-icon dark> mdi-account </v-icon>
      </v-btn>
    </template>
    <v-list dence min-width="250px">
      <v-list-item>
        <v-list-item-content>
          <v-list-item-title class="text-h6">
            {{ userdata.title }}
          </v-list-item-title>
          <v-list-item-subtitle>#{{ userdata.uuid }}</v-list-item-subtitle>
        </v-list-item-content>
      </v-list-item>
      <v-list-item>
        <balance title="Balance: " loged-in-user />
      </v-list-item>
      <v-divider></v-divider>
      <v-list-item @click="logoutHandler">
        <v-list-item-title>Logout</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script setup>
import balance from "@/components/balance.vue";
import { useStore } from "@/store";
import { computed } from "vue";

const store = useStore();

const userdata = computed(() => {
  return store.getters["auth/userdata"];
});

function logoutHandler() {
  store.dispatch("auth/logout");
}
</script>
