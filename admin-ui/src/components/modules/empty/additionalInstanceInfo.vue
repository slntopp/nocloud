<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          label="IP"
          :value="template.resources.ip"
          @input="$emit('update', { key: 'resources.ip', value: $event })"
        />
      </v-col>

      <v-col v-for="key in Object.keys(template.resources)" :key="key">
        <v-text-field :label="key" :value="template.resources[key]" readonly />
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import { toRefs, defineProps } from "vue";

const props = defineProps(["template", "sp"]);

const { template } = toRefs(props);
</script>

<style scoped></style>
