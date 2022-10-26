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

      Object.keys(this.fields).forEach((fieldName) => {
        this.fields[fieldName].rules.forEach((rule) => {
          const result = rule(this.secrets[fieldName]);

          if (typeof result === "string") {
            this.errors[fieldName] = [result];
            errors[fieldName] = result;
          } else {
            this.errors[fieldName] = [];
          }
        });
      });

      if (this.secrets.app_key) {
        newSecrets.app_key = this.secrets.app_key;
      }
      if (this.secrets.app_secret) {
        newSecrets.app_secret = this.secrets.app_secret;
      }
      if (this.secrets.consumer_key) {
        newSecrets.consumer_key = this.secrets.consumer_key;
      }
      if (this.secrets.endpoint) {
        newSecrets.endpoint = this.secrets.endpoint;
      }

      newSecrets[input] = data;

      console.log(this.secrets);
      console.log(newSecrets);

      this.$emit(`change:secrets`, newSecrets);
      this.$emit(`passed`, Object.keys(errors).length === 0);
      console.log(errors);
    },
    getValue(fieldName) {
      switch (fieldName) {
        case "app_key":
          return this.secrets.app_key;
        case "app_secret":
          return this.secrets.app_secret;
        case "consumer_key":
          return this.secrets.consumer_key;
        case "endpoint":
          return this.secrets.endpoint;
        default:
          return "";
      }
    },
  },
};
</script>
