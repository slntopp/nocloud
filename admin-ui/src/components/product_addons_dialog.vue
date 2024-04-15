<template>
  <v-dialog width="90vw">
    <template v-slot:activator="{ on, attrs }">
      <v-btn icon v-bind="attrs" v-on="on">
        <v-icon> mdi-menu-open </v-icon>
      </v-btn>
    </template>

    <v-card color="background-light">
      <plans-addons-table
        @change:addons="emit('change:addons', $event)"
        :addons="addons"
      />
    </v-card>
  </v-dialog>
</template>

<script setup>
import { toRefs } from "vue";
import PlansAddonsTable from "@/components/planAddonsTable.vue";

const props = defineProps({
  addons: { type: Array, required: true },
});
const { addons } = toRefs(props);

const emit = defineEmits(["change:addons"]);
</script>
