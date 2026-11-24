<template>
  <div class="pa-4">
    <v-row v-if="!editing">
      <v-col cols="12" md="6">
        <v-text-field readonly :model-value="account?.mrn" label="service account" />
      </v-col>
      <v-col cols="12" md="6">
        <v-text-field
          readonly
          :model-value="account?.space_mrn || account?.parent_mrn"
          label="space"
        />
      </v-col>
      <v-col cols="12" md="6">
        <v-text-field readonly :model-value="account?.api_endpoint" label="api endpoint" />
      </v-col>
      <v-col cols="12" md="6">
        <v-text-field readonly :model-value="importMode" label="import mode" />
      </v-col>
      <v-col cols="12" md="6">
        <v-text-field
          readonly
          :model-value="template.secrets?.ssh_login_key ? 'accepted' : 'not accepted'"
          label="SSH logins from customers"
        />
      </v-col>
    </v-row>
    <slot v-else></slot>
  </div>
</template>

<script>
import { parseServiceAccount, importModes } from "./serviceAccount";

export default {
  props: { template: { type: Object, required: true } },
  data: () => ({ editing: false }),
  computed: {
    account() {
      return parseServiceAccount(this.template.secrets?.service_account).account;
    },
    importMode() {
      const mode = this.template.secrets?.import_mode || "";
      return importModes.find((m) => m.value === mode)?.title ?? mode;
    },
  },
};
</script>
