<template>
  <v-menu
    bottom
    open-on-hover
    nudge-top="20"
    nudge-left="15"
    transition="slide-y-transition"
  >
    <template v-slot:activator="{ on, attrs }">
        <v-text-field v-if="props.ui!=='span'" v-bind="attrs" v-on="on" readonly :value="ip" />
        <span v-else v-bind="attrs" v-on="on">{{ip}}</span>
    </template>

    <v-list dense>
      <v-list-item v-for="net of item.state.meta.networking.public" :key="net">
        <v-list-item-title>{{ net }}</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script setup>
import { computed, defineProps } from "vue";
const props = defineProps(["item",'ui']);

const ip = computed(
  () =>
    props.item.state.meta.networking.public.find(
      (ip) =>
        /^(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}$/gm.exec(
          ip
        ) || /\/32$/.exec(ip)
    ) || props.item.state.meta.networking.public[0]
);
</script>
