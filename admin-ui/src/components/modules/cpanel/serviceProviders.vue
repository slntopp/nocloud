<template>
  <div class="sp-ovh">
    <v-row align="center" v-for="field in Object.keys(fields)" :key="field">
      <v-col cols="4">
        <v-subheader>
          {{ fields[field].subheader || field }}
        </v-subheader>
      </v-col>

      <v-col cols="8">
        <v-text-field
          :value="getValue(field)"
          :label="fields[field].label"
          :rules="fields[field].rules"
          :error-messages="errors[field]"
          @change="(data) => changeHandler(field, data)"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script>
export default {
  name: "servicesProviders-create-cpanel",
  props: {
    secrets: {
      type: Object,
      default: () => ({}),
    },
    passed: {
      type: Boolean,
      default: false,
    },
  },
  data: () => ({
    errors: {
      domain: [],
      token: [],
    },
    fields: {
      domain: {
        label: "domain",
        subheader: "Domain",
        rules: [(value) => !!value || "Field is required"],
      },
      token: {
        label: "token",
        subheader: "Token",
        rules: [(value) => !!value || "Field is required"],
      },
    },
  }),
  methods: {
    changeHandler(input, data) {
      const errors = {};
      const newSecrets = {};

      for (const key of Object.keys(this.secrets)) {
        newSecrets[key] = this.secrets[key];
      }

      newSecrets[input] = data;
      this.$emit(`change:secrets`, newSecrets);
      this.$emit(`passed`, Object.keys(errors).length === 0);

      this.fields[input].rules.forEach((rule) => {
        const result = rule(data);
        if (typeof result == "string") {
          this.errors[input] = [result];
          errors[input] = result;
        } else {
          this.errors[input] = [];
        }
      });
    },
    getValue(fieldName) {
      return this.secrets[fieldName];
    },
  },
};
</script>
