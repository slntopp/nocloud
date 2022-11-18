<template>
  <div class="sp-goget">
    <v-row align="center" v-for="field in Object.keys(fields)" :key="field">
      <v-col cols="4">
        <v-subheader>
          {{ fields[field].subheader || field }}
        </v-subheader>
      </v-col>

      <v-col cols="4" v-if="fields[field].items">
        <v-select
          item-text="text"
          item-value="key"
          :value="getValue(field)"
          :items="fields[field].items"
          :label="fields[field].label"
          :rules="fields[field].rules"
          :error-messages="errors[field]"
          @change="(data) => changeHandler(field, data)"
        />
      </v-col>
      <v-col :cols="fields[field].items ? 4 : 8">
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
  name: "servicesProviders-create-goget",
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
    hostWarning: false,
    errors: {
      host: [],
      username: [],
      password: [],
    },
    fields: {
      host: {
        label: "example.com",
        subheader: "Host",
        rules: [
          (value) => !!value || "Field is required",
          (value) => {
            try {
              new URL(value);
              return true;
            } catch (err) {
              return "Is not valid domain";
            }
          },
        ],
      },
      username: {
        label: "username",
        subheader: "Username (login)",
        rules: [(value) => !!value || "Field is required"],
      },
      password: {
        label: "password",
        subheader: "Password",
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
  watch: {
    "secrets.host"(newVal) {
      try {
        const url = new URL(newVal);
        if (url.pathname !== "/RPC2") this.hostWarning = url.pathname;
        else this.hostWarning = false;
      } catch (err) {
        this.hostWarning = false;
      }
    },
  },
};
</script>
