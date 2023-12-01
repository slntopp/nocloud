<template>
  <v-expansion-panels>
    <v-expansion-panel>
      <v-expansion-panel-header color="background-light">
        <span style="color: var(--v-primary-base)" class="text-h6">{{ title }}</span>
        <template v-slot:actions>
          <v-icon color="primary" x-large> $expand </v-icon>
        </template>
      </v-expansion-panel-header>
      <v-expansion-panel-content color="background-light">
        <slot />
      </v-expansion-panel-content>
    </v-expansion-panel>
  </v-expansion-panels>
</template>

<script>
export default {
  name: "instancesPricesPanels",
  props: ['title']
};
</script>
