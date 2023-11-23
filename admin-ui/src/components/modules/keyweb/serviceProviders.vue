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
  name: "servicesProviders-create-openai",
  props: {
    secrets: {
      type: Object,
      default: () => ({}),
    },
    vars: {
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
      mail: {
        label: "Mail",
        subheader: "Mail",
        rules: [
          (value) =>
            !!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.exec(value) || "Wrong email",
        ],
        type: "secrets",
      },
      host: {
        label: "Host",
        subheader: "Host",
        rules: [
          (value) => !value || value.startsWith("http") || "Wrong host",
          (value) => !!value || "Field is required",
        ],
        type: "secrets",
      },
      api_key: {
        label: "Api key",
        subheader: "Api key",
        type: "secrets",
        rules: [(value) => !!value || "Field is required"],
      },
    },
  }),
  methods: {
    changeHandler(input, data) {
      const errors = {};
      const newSecrets = {};
      const newVars = {};

      for (const key of Object.keys(this.secrets)) {
        newSecrets[key] = this.secrets[key];
      }
      for (const key of Object.keys(this.vars)) {
        newVars[key] = this.vars[key];
      }

      if (this.fields[input].type === "secrets") {
        newSecrets[input] = data;
      } else {
        newVars[input] = { value: { default: data } };
      }

      this.$emit(`change:secrets`, newSecrets);
      this.$emit(`change:vars`, newVars);
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
      if (this.fields[fieldName].type === "secrets") {
        return this.secrets[fieldName];
      }
      return this.vars[fieldName]?.value?.default;
    },
  },
};
</script>
