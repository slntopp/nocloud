<template>
  <div class="settings pa-10">
    <v-row justify="center">
      <v-card-title>Map settings</v-card-title>
    </v-row>
    <v-card-title>Map widget script:</v-card-title>
    <v-text-field
      filled
      color="background-light"
      :value="script"
      readonly
      append-icon="mdi-content-copy"
      @click:append="copyScript"
    />
  </div>
</template>

<script>
export default {
  data: () => ({
    script: null,
  }),
  mounted() {
    this.script = "<script src=./my-lib.umd.js><script>";
  },
  methods: {
    copyScript() {
      navigator.clipboard.writeText(this.script);
    },
  },
};
</script>
