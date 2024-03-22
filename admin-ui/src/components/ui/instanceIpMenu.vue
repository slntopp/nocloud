<template>
  <v-menu bottom nudge-top="20" nudge-left="15" transition="slide-y-transition" :close-on-content-click="false">
    <template v-slot:activator="{ on, attrs }">
      <v-text-field :label="`IPs_${type}`" v-if="props.ui !== 'span'" v-bind="attrs" v-on="on" readonly :value="ip"
        :append-icon="ip ? 'mdi-content-copy' : ''" @click:append="addToClipboard(ip)">
      </v-text-field>
      <span v-else v-bind="attrs" v-on="on">{{ ip }}</span>
    </template>

    <v-list v-if="item.state.meta.networking?.[type]?.length" dense>
      <v-list-item v-for="net of item.state.meta.networking?.[type]" :key="net">
        <v-list-item-title>{{ net }}</v-list-item-title>
        <v-icon @click="addToClipboard(net)"> mdi-content-copy</v-icon>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script setup>
import { computed, defineProps } from "vue";
import { addToClipboard } from "@/functions";
const props = defineProps({
  item: {},
  ui: {},
  type: { type: String, default: "public" },
});

const ip = computed(
  () =>
    props.item.state.meta.networking?.[props.type]?.find(
      (ip) =>
        /^(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}$/gm.exec(
          ip
        ) || /\/32$/.exec(ip)
    ) || props.item.state.meta.networking?.[props.type]?.[0]
);
</script>
