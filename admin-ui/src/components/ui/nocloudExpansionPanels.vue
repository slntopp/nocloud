<template>
  <v-expansion-panels>
    <v-expansion-panel>
      <v-expansion-panel-title color="background-light">
        <span style="color: rgb(var(--v-theme-primary))" class="text-h6">{{ title }}</span>
        <template v-slot:actions>
          <v-icon color="primary" size="x-large"> $expand </v-icon>
        </template>
      </v-expansion-panel-title>
      <v-expansion-panel-text color="background-light">
        <slot />
      </v-expansion-panel-text>
    </v-expansion-panel>
  </v-expansion-panels>
</template>

<script>
export default {
  name: "nocloudExpansionPanels",
  props: ['title']
};
</script>
