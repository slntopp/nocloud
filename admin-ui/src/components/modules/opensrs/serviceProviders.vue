<template>
  <div class="sp-openSrs">
    <v-row
      v-for="field in Object.keys(fields)"
      :key="field"
      :align="!fields[field].isJSON ? 'center' : null"
    >
      <v-col cols="4">
        <v-subheader>
          {{ fields[field].subheader || field }}

          <v-tooltip
            v-if="field == 'host' && hostWarning"
            bottom
            color="warning"
          >
            <template v-slot:activator="{ on, attrs }">
              <v-icon class="ml-2" color="warning" v-bind="attrs" v-on="on">
                mdi-alert-outline
              </v-icon>
            </template>

            <span
              >Non-standard RPC path: "{{ hostWarning }}" instead of
              "/RPC2"</span
            >
          </v-tooltip>
        </v-subheader>
      </v-col>

      <v-col cols="8">
        <v-text-field
          @change="(data) => changeHandler(field, data)"
          :value="getValue(field)"
          :label="fields[field].label"
          :rules="fields[field].rules"
          :error-messages="errors[field]"
          :type="fields[field].type"
          v-bind="fields[field].bind || {}"
          v-if="!fields[field].isJSON"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script>
export default {
  name: "servicesProviders-create-openSrs",
  props: {
    passed: {
      type: Boolean,
      default: false,
    },
    secrets: {
      type: Object,
    },
  },
  data: () => ({
    hostWarning: false,
    errors: {
      host: [],
      username: [],
      api_key: [],
    },
    fields: {
      host: {
        type: "text",
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
        label: "example.com",
      },
      username: {
        type: "text",
        subheader: "Username(Login)",
        rules: [(value) => !!value || "Field is required"],
        label: "username",
      },
      api_key: {
        type: "text",
        subheader: "Api key",
        rules: [(value) => !!value || "Field is required"],
        label: "api key",
      },
    },
  }),
  methods: {
    changeHandler(input, data) {
      const errors = {};
      const secrets = {};

      for (const key of Object.keys(this.secrets)) {
        secrets[key] = this.secrets[key];
      }

      secrets[input] = data;

      const result = {
        secrets,
      };

      this.$emit(`change:secrets`, secrets);
      this.$emit(`change:full`, result);
      this.$emit(`passed`, Object.keys(errors).length == 0);

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
  created() {
    this.host = this.secrets.host;
    this.user = this.secrets.username;
    this.api_key = this.secrets.api_key;
  },
  watch: {
    host(newVal) {
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

<style></style>
