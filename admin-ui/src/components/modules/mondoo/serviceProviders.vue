<template>
  <div class="sp-mondoo">
    <v-row class="align-start">
      <v-col cols="4">
        <v-list-subheader>Service account</v-list-subheader>
      </v-col>
      <v-col cols="8">
        <v-textarea
          :model-value="accountText"
          label="Service account JSON"
          hint="Paste the credentials file of a Mondoo service account for the space (JSON or the base64 form the console downloads). Editor role is needed to create integrations; Viewer is enough with import mode 'only'."
          persistent-hint
          rows="6"
          auto-grow
          spellcheck="false"
          :error-messages="errors"
          @update:model-value="changeAccount"
        />
      </v-col>
    </v-row>

    <v-row v-if="account" class="align-center">
      <v-col cols="4">
        <v-list-subheader>Space</v-list-subheader>
      </v-col>
      <v-col cols="8">
        <v-text-field readonly :model-value="space" label="space mrn" />
      </v-col>
    </v-row>

    <v-row class="align-center">
      <v-col cols="4">
        <v-list-subheader>Existing integrations</v-list-subheader>
      </v-col>
      <v-col cols="8">
        <v-select
          :model-value="secrets.import_mode || ''"
          :items="importModes"
          item-title="title"
          item-value="value"
          label="import mode"
          :hint="importHint"
          persistent-hint
          @update:model-value="changeImportMode"
        />
      </v-col>
    </v-row>

    <v-row class="align-center">
      <v-col cols="4">
        <v-list-subheader>SSH logins from customers</v-list-subheader>
      </v-col>
      <v-col cols="8">
        <v-switch
          :model-value="!!secrets.ssh_login_key"
          color="primary"
          label="accept a login for one scan or fix"
          hint="For machines the scan runner's key no longer opens. Turning this on creates the key such logins are encrypted with; the Ansible service reads it from this provider. Add INSTANCE_LOGIN: ssh_login.box to the vars of the SCAN and REMEDIATE hooks."
          persistent-hint
          @update:model-value="changeLogins"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script>
import { parseServiceAccount, importModes } from "./serviceAccount";

export default {
  name: "servicesProviders-create-mondoo",
  props: {
    secrets: { type: Object, default: () => ({}) },
    passed: { type: Boolean, default: false },
  },
  emits: ["change:secrets", "passed"],
  data: () => ({ text: null, errors: [], importModes }),
  computed: {
    account() {
      return parseServiceAccount(this.secrets.service_account).account;
    },
    accountText() {
      if (this.text !== null) return this.text;
      return this.account ? JSON.stringify(this.account, null, 2) : "";
    },
    space() {
      return this.account?.space_mrn || this.account?.parent_mrn || "";
    },
    importHint() {
      return this.importModes.find(
        (m) => m.value === (this.secrets.import_mode || "")
      )?.hint;
    },
  },
  mounted() {
    // Editing an existing provider: what is stored already passed the driver's test.
    if (this.account) this.$emit("passed", true);
  },
  methods: {
    changeAccount(value) {
      this.text = value;
      const { account, error } = parseServiceAccount(value);
      this.errors = error ? [error] : [];
      if (account) {
        this.$emit("change:secrets", {
          ...this.secrets,
          service_account: account,
        });
      }
      this.$emit("passed", !!account);
    },
    changeLogins(on) {
      const secrets = { ...this.secrets };
      if (on) {
        // 32 random bytes, base64: the key the driver and the Ansible
        // service share through this provider's secrets.
        const bytes = crypto.getRandomValues(new Uint8Array(32));
        secrets.ssh_login_key = btoa(String.fromCharCode(...bytes));
      } else delete secrets.ssh_login_key;
      this.$emit("change:secrets", secrets);
      this.$emit("passed", !!this.account);
    },
    changeImportMode(value) {
      const secrets = { ...this.secrets };
      if (value) secrets.import_mode = value;
      else delete secrets.import_mode;
      this.$emit("change:secrets", secrets);
      this.$emit("passed", !!this.account);
    },
  },
};
</script>
