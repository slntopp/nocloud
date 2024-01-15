<template>
  <div class="sp-ione">
    <v-row
      v-for="field in fieldKeys"
      :key="field"
      :align="!fields[field].isJSON ? 'center' : null"
    >
      <v-col cols="4">
        <subheader-with-info :infoText="`SPInfo.ione.${field}`">
          {{ fields[field].subheader || field }}
        </subheader-with-info>
      </v-col>

      <v-col cols="8">
        <v-switch
          v-if="fields[field]?.type === 'bool'"
          v-model="fields[field].bind"
          :label="fields[field].label"
          @change="(data) => changeHandler(field, data)"
        ></v-switch>
        <v-text-field
          @change="(data) => changeHandler(field, data)"
          :value="getValue(field)"
          :label="fields[field].label"
          :rules="fields[field].rules"
          :error-messages="errors[field]"
          :type="fields[field].type"
          v-bind="fields[field].bind || {}"
          v-if="!fields[field].isJSON && fields[field].type !== 'bool'"
        />
        <json-editor
          v-if="fields[field].isJSON"
          :json="getValue(field)"
          @changeValue="(data) => changeHandler(field, data)"
        />
      </v-col>
    </v-row>
    <!-- Vlans key -->
    <v-row class="ml-2">
      <v-col cols="4" v-for="field in vlansKeys()" :key="field">
        <v-select
          v-if="fields[field].type === 'select'"
          :items="vlansKeysItems"
          v-bind="fields[field].bind || {}"
          :value="getValue(field)"
          :placeholder="fields[field].label"
          @change="(data) => changeHandler(field, data)"
        />
        <v-text-field
          v-else
          :placeholder="fields[field].label"
          @change="(data) => changeHandler(field, data)"
          :value="getValue(field)"
          :label="fields[field].subheader"
          :rules="fields[field].rules"
          :error-messages="errors[field]"
          :type="fields[field].type"
          v-bind="fields[field].bind || {}"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script>
import JsonEditor from "@/components/JsonEditor.vue";
import subheaderWithInfo from "@/components/ui/subheaderWithInfo.vue";

function isJSON(str) {
  try {
    JSON.parse(str);
    return true;
  } catch {
    return false;
  }
}

// const domainRegex = /^((https?:\/\/)|(www.))(?:(\.?[a-zA-Z0-9-]+){1,}|(\d+\.\d+.\d+.\d+))(\.[a-zA-Z]{2,})?(:\d{4})?\/?$/;
// const domainRegex = /^(https?):\/\/(((?!-))(xn--|_{1,1})?[a-z0-9-]{0,61}[a-z0-9]{1,1}\.)*(xn--)?([a-z0-9][a-z0-9-]{0,60}|[a-z0-9-]{1,30}\.[a-z]{2,})$/
export default {
  components: { JsonEditor, subheaderWithInfo },
  name: "servicesProviders-create-ione",
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
    vlansKeysItems: ["vcenter"],
    hostWarning: false,
    errors: {
      host: [],
      user: [],
      pass: [],
      group: [],
      size: [],
      start: [],
      sched: [],
      sched_ds: [],
      public_ip_pool: [],
      private_vnet_tmpl: [],
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
      user: {
        type: "text",
        subheader: "user(Login)",
        rules: [(value) => !!value || "Field is required"],
        label: "user",
      },
      pass: {
        type: "pass",
        subheader: "pass or Token",
        rules: [(value) => !!value || "Field is required"],
        label: "pass",
      },
      group: {
        type: "number",
        subheader: "Group",
        rules: [
          (value) => !!value || "Field is required",
          (value) => value == Number(value) || "wrong group id",
        ],
        label: "100",
        bind: {
          min: 0,
        },
      },
      vlansKey: {
        type: "select",
        subheader: "Vlans key",
        label: "vlans key",
        rules: [() => true],
      },
      start: {
        type: "number",
        subheader: "Start",
        label: "number between 1 and 4096",
        rules: [(value) => +value <= 4096 || "Vlаns cant be more thna 4096"],
        bind: {
          min: 0,
        },
      },
      size: {
        type: "number",
        subheader: "Size",
        label: "number between 1 and 4096",
        rules: [(value) => +value <= 4096 || "Vlаns cant be more thna 4096"],
        bind: {
          min: 0,
        },
      },
      sched: {
        type: "text",
        subheader: "schedr rules",
        isJSON: true,
        rules: [
          (value) => !!value || "Field is required",
          (value) => isJSON(value) || "is not valid JSON",
        ],
        label: "JSON",
      },
      sched_ds: {
        type: "text",
        subheader: "DataStore schedr rules",
        isJSON: true,
        rules: [
          (value) => !!value || "Field is required",
          (value) => isJSON(value) || "is not valid JSON",
        ],
        label: "JSON",
      },
      public_ip_pool: {
        type: "number",
        subheader: "Public IPs Pool ID",
        rules: [
          (value) => !!value || value === 0 || "Field is required",
          (value) => value == 0 || !!Number(value) || "Field must be number",
        ],
        label: "pip",
        bind: {
          min: 0,
        },
      },
      private_vnet_tmpl: {
        type: "number",
        subheader: "Private Networks Template ID",
        rules: [
          (value) => !!value || value === 0 || "Field is required",
          (value) => value == 0 || !!Number(value) || "Field must be number",
        ],
        label: "pvp",
        bind: {
          min: 0,
        },
      },
      private_vnet_ban: {
        type: "bool",
        subheader: "Private Net Functions",
        rules: [() => true],
      },
    },
  }),
  methods: {
    requiredField(value) {
      return !!value || "Field is required";
    },
    isDomain(value) {
      const reg =
        /^(https?):\/\/(((?!-))(xn--|_{1,1})?[a-z0-9-]{0,61}[a-z0-9]{1,1}\.)*(xn--)?([a-z0-9][a-z0-9-]{0,60}|[a-z0-9-]{1,30}\.[a-z]{2,})$/;
      return !!value.match(reg) || "Is not valid domain";
    },
    isNumber(value) {
      return value === Number(value) || "Is not valid domain";
    },
    changeHandler(input, data) {
      let errors = {};
      const secrets = {};
      const vars = {};

      for (const secretKey of this.secretsKeys()) {
        secrets[secretKey] =
          secretKey === input ? data : this.getValue(secretKey);
      }

      secrets.group = "group" === input ? +data : +this.getValue("group");

      if (this.getValue("vlansKey") || input === "vlansKey") {
        if (input === "vlansKey") {
          secrets.vlans = {
            [data]: {
              start: +(this.getValue("start") ?? 0),
              size: +(this.getValue("size") ?? 0),
            },
          };
        } else {
          secrets.vlans = {
            [this.getValue("vlansKey")]: {
              start: +(this.getValue("start") ?? 0),
              size: +(this.getValue("size") ?? 0),
            },
          };
        }
      }

      for (const varKey of this.jsonVarsKeys()) {
        this.setVarsValueJSON(vars, varKey, errors, varKey === input, data);
      }

      const defaultVars = [
        "public_ip_pool",
        "private_vnet_tmpl",
        "private_vnet_ban",
      ];
      for (const varKey of defaultVars) {
        this.setVarsValueDefault(vars, varKey, varKey === input, data);
      }

      const result = {
        secrets,
        vars,
      };

      this.$emit(`change:secrets`, secrets);
      this.$emit(`change:vars`, vars);
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

    setVarsValueDefault(vars, fieldName, isChange, data) {
      vars[fieldName] = {
        value: {
          default: isChange ? JSON.parse(data) : this.getValue(fieldName),
        },
      };
    },

    setVarsValueJSON(vars, fieldName, errors, isChange, data) {
      if (isJSON(JSON.stringify(this.getValue(fieldName)))) {
        vars[fieldName] = { value: isChange ? data : this.getValue(fieldName) };
        delete errors[fieldName];
      } else {
        errors[fieldName] = ["is not valid JSON"];
      }
    },

    getValue(fieldName) {
      if (this.isVlansKey(fieldName)) {
        const vlansKey = Object.keys(this.secrets.vlansKey ?? {})[0];

        if (fieldName === "vlansKey") {
          return vlansKey ?? "";
        } else if ("start" === fieldName) {
          return vlansKey ? this.secrets.vlansKey[vlansKey]?.start : 0;
        } else {
          return vlansKey ? this.secrets.vlansKey[vlansKey]?.size : 0;
        }
      }

      if (this.secretsKeys().includes(fieldName) || fieldName === "group") {
        return this.secrets[fieldName];
      }

      if (this.jsonVarsKeys().includes(fieldName)) {
        if (typeof (this.vars[fieldName]?.value ?? {}) === "object") {
          return this.vars[fieldName]?.value ?? {};
        }
        return {};
      }

      switch (fieldName) {
        case "public_ip_pool":
          return this.vars.public_ip_pool?.value?.default ?? 0;
        case "private_vnet_tmpl":
          return this.vars.private_vnet_tmpl?.value?.default ?? 0;
        case "private_vnet_ban":
          return this.vars.private_vnet_ban?.default ?? false;
        default:
          return "";
      }
    },
    isVlansKey(field) {
      return this.vlansKeys().includes(field);
    },
    vlansKeys() {
      return ["vlansKey", "start", "size"];
    },
    secretsKeys() {
      return ["host", "user", "pass"];
    },
    jsonVarsKeys() {
      return ["sched_ds", "sched"];
    },
  },
  mounted() {
    if (!this.vars.console?.value) {
      this.$emit("change:vars", {
        ...this.vars,
        console: { value: { default: "vnc" } },
      });
    }
  },
  computed: {
    fieldKeys() {
      return Object.keys(this.fields).filter((key) => !this.isVlansKey(key));
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

<style></style>
