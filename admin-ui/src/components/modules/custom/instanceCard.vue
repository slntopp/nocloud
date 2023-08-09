<template>
  <div>
    <v-row class="mt-3">
      <h3 v-if="dense">Config:</h3>
      <v-card-title v-else class="px-0">Config:</v-card-title>
    </v-row>
    <v-row align="center">
      <v-col v-for="(item, key) in template.config" :key="key">
        <v-text-field
          readonly
          style="display: inline-block; width: 200px"
          :label="dictionary[key] ?? key.replace('_', ' ')"
          :value="item"
        />
      </v-col>
    </v-row>

    <v-row class="mt-3">
      <h3 v-if="dense">Resources:</h3>
      <v-card-title v-else class="px-0">Resources:</v-card-title>
    </v-row>
    <v-row align="center">
      <v-col v-for="(item, key) in template.resources" :key="key">
        <v-text-field
          readonly
          style="display: inline-block; width: 200px"
          :label="dictionary[key] ?? key.replace('_', ' ')"
          :value="item"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script>
export default {
  name: "instance-card",
  props: {
    template: { type: Object, required: true },
    dense: { type: Boolean },
  },
  data: () => ({
    dictionary: {},
  }),
};
</script>
