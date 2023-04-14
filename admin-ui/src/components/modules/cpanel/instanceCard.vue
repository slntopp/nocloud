<template>
  <v-container>
    <v-row>
      <v-col cols="6">
        <v-text-field
          readonly
          :value="template.config.domain"
          label="domain"
        ></v-text-field>
      </v-col>
      <v-col cols="6">
        <v-text-field
          readonly
          :value="template.config.mail"
          label="mail"
        ></v-text-field>
      </v-col>
    </v-row>
    <v-row>
      <v-col cols="6">
        <v-text-field
          readonly
          :value="template.resources.plan"
          label="plan"
        ></v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.config.password"
          label="password"
        ></v-text-field>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
export default {
  name: "instance-card",
  props: {
    template: { type: Object, required: true },
    provider: { type: String, required: true },
    dense: { type: Boolean },
  },
};
</script>
