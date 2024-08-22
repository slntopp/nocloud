<template>
  <div>
    <v-row>
      <v-col>
        <instance-ip-menu edit :item="template" />
      </v-col>
      <v-col>
        <instance-ip-menu edit type="private" :item="template" />
      </v-col>

      <v-col v-for="key in Object.keys(template.resources || {})" :key="key">
        <v-text-field
          append-icon="mdi-pencil"
          @input="emit('update', { key: `resources.${key}`, value: $event })"
          :label="key"
          :value="template.resources[key]"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import { toRefs, defineProps } from "vue";
import InstanceIpMenu from "@/components/ui/instanceIpMenu.vue";

const props = defineProps(["template", "sp"]);
const { template } = toRefs(props);

const emit = defineEmits(["update"]);
</script>

<style scoped></style>
