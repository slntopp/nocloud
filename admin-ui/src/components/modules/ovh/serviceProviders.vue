<template>
  <div class="sp-ovh">
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
  name: "servicesProviders-create-ovh",
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
      appKey: [],
      appSecret: [],
      consumerKey: [],
      endpoint: [],
    },
    fields: {
      app_key: {
        label: "app key",
        subheader: "App key",
        rules: [(value) => !!value || "Field is required"],
      },
      app_secret: {
        label: "app secret",
        subheader: "App secret",
        rules: [(value) => !!value || "Field is required"],
      },
      consumer_key: {
        label: "consumer key",
        subheader: "Consumer key",
        rules: [(value) => !!value || "Field is required"],
      },
      endpoint: {
        label: "endpoint",
        subheader: "Endpoint",
        items: [
          { key: "ovh-eu", text: "Europe" },
          { key: "ovh-us", text: "USA" },
          { key: "ovh-ca", text: "Canada" },
        ],
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
