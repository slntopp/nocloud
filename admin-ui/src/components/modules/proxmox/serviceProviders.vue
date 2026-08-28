<template>
  <div class="sp-proxmox">
    <v-row v-for="field in fieldKeys" :key="field" align="center">
      <v-col cols="4">{{ fields[field].subheader }}</v-col>
      <v-col cols="8">
        <v-switch
          v-if="fields[field].type === 'bool'"
          v-model="fields[field].bind"
          :label="fields[field].label"
          @change="(data) => changeHandler(field, data)"
        />
        <v-text-field
          v-else
          :value="getValue(field)"
          :label="fields[field].label"
          :rules="fields[field].rules"
          :type="fields[field].type"
          @change="(data) => changeHandler(field, data)"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script>
export default {
  name: "servicesProviders-create-proxmox",
  props: {
    secrets: { type: Object, default: () => ({}) },
    vars: { type: Object, default: () => ({}) },
    passed: { type: Boolean, default: false },
  },
  data: () => ({
    fields: {
      host: {
        secret: true,
        subheader: "Host",
        type: "text",
        label: "https://pve.example.com:8006",
        rules: [(v) => !!v || "Required"],
      },
      token_id: {
        secret: true,
        subheader: "Token ID",
        type: "text",
        label: "user@pam!nocloud",
        rules: [(v) => !!v || "Required"],
      },
      token_secret: {
        secret: true,
        subheader: "Token secret",
        type: "password",
        label: "token secret",
        rules: [(v) => !!v || "Required"],
      },
      insecure: {
        secret: true,
        subheader: "Skip TLS verify",
        type: "bool",
        label: "insecure",
        bind: false,
      },
      storage: {
        secret: false,
        subheader: "Storage",
        type: "text",
        label: "local-lvm",
        rules: [(v) => !!v || "Required"],
      },
      bridge: {
        secret: false,
        subheader: "Bridge",
        type: "text",
        label: "vmbr0",
        rules: [(v) => !!v || "Required"],
      },
      default_node: {
        secret: false,
        subheader: "Default node (optional)",
        type: "text",
        label: "pve",
        rules: [],
      },
    },
  }),
  computed: {
    fieldKeys() {
      return Object.keys(this.fields);
    },
  },
  methods: {
    getValue(field) {
      const f = this.fields[field];
      if (f.secret) return this.secrets?.[field];
      return this.vars?.[field]?.value?.default;
    },
    changeHandler(field, data) {
      const f = this.fields[field];
      if (f.type === "bool") {
        data = !!data;
      }
      if (f.secret) {
        this.$emit("update:secrets", { ...this.secrets, [field]: data });
      } else {
        this.$emit("update:vars", {
          ...this.vars,
          [field]: { value: { default: data } },
        });
      }
      this.checkPass();
    },
    checkPass() {
      const ok =
        !!this.secrets?.host &&
        !!this.secrets?.token_id &&
        !!this.secrets?.token_secret &&
        !!this.vars?.storage?.value?.default &&
        !!this.vars?.bridge?.value?.default;
      this.$emit("update:passed", ok);
    },
  },
  mounted() {
    this.checkPass();
  },
};
</script>
