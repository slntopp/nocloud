<template>
  <v-subheader>
    <slot></slot>
    <v-tooltip
      open-on-click
      :open-on-hover="false"
      :open-on-focus="false"
      :offset-overflow="true"
      v-if="infoText && isInfoExists"
      bottom
    >
      <template v-slot:activator="{ on, attrs }">
        <v-icon class="ml-5" dark v-bind="attrs" v-on="on">
          mdi-information-outline
        </v-icon>
      </template>
      <p class="text subtitle-2 hint">{{ $t(infoText) }}</p>
    </v-tooltip>
  </v-subheader>
</template>

<script>
export default {
  name: "subheader-with-info",
  props: {
    infoText: { type: String },
  },
  computed: {
    isInfoExists() {
      return this.$t(this.infoText) !== this.infoText;
    },
  },
};
</script>

<style scoped>
.hint {
  max-width: 250px;
  inline-size: 250px;
  overflow-wrap: break-word;
  user-select: text;
  pointer-events: all;
}
</style>
