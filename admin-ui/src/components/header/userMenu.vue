<template>
  <v-menu transition="slide-y-transition">
    <template v-slot:activator="{ props }">
      <v-btn
        class="mx-2"
        icon
        variant="elevated"
        size="large"
        color="background-light"
        v-bind="props"
      >
        <v-icon > mdi-account </v-icon>
      </v-btn>
    </template>
    <v-list dence min-width="250px">
      <v-list-item>
        
          <v-list-item-title class="text-h6">
            {{ userdata.title }}
          </v-list-item-title>
          <v-list-item-subtitle>#{{ userdata.uuid }}</v-list-item-subtitle>
        
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
