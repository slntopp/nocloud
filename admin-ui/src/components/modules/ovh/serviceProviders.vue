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
    values: {
      appKey: "",
      appSecret: "",
      consumerKey: "",
      endpoint: "",
    },
    fields: {
      appKey: {
        label: "app key",
        subheader: "App key",
        rules: [(value) => !!value || "Field is required"],
      },
      appSecret: {
        label: "app secret",
        subheader: "App secret",
        rules: [(value) => !!value || "Field is required"],
      },
      consumerKey: {
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
      const secrets = {};
      this.values[input] = data;

      Object.keys(this.fields).forEach((fieldName) => {
        this.fields[fieldName].rules.forEach((rule) => {
          const result = rule(this.values[fieldName]);

          if (typeof result === "string") {
            this.errors[fieldName] = [result];
            errors[fieldName] = result;
          } else {
            this.errors[fieldName] = [];
          }
        });
      });

      if (this.values.appKey) {
        secrets.app_key = this.values.appKey;
      }
      if (this.values.appSecret) {
        secrets.app_secret = this.values.appSecret;
      }
      if (this.values.consumerKey) {
        secrets.consumer_key = this.values.consumerKey;
      }
      if (this.values.endpoint) {
        secrets.endpoint = this.values.endpoint;
      }

      this.$emit(`change:secrets`, secrets);
      this.$emit(`passed`, Object.keys(errors).length === 0);
      console.log(errors);
    },
    getValue(fieldName) {
      switch (fieldName) {
        case "appKey":
          return this.secrets.app_key;
        case "appSecret":
          return this.secrets.app_secret;
        case "consumerKey":
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
