<template>
  <div class="pa-4">
    <v-row v-if="!editing">
      <v-col cols="12" md="4">
        <v-text-field readonly :value="secrets.host" label="host" />
      </v-col>
      <template v-if="secrets.token_id">
        <v-col cols="12" md="4">
          <v-text-field readonly :value="secrets.token_id" label="token_id" />
        </v-col>
        <v-col cols="12" md="4">
          <password-text-field :value="secrets.token_secret" label="token_secret" />
        </v-col>
      </template>
      <template v-else>
        <v-col cols="12" md="4">
          <v-text-field readonly :value="secrets.user" label="user" />
        </v-col>
        <v-col cols="12" md="4">
          <password-text-field :value="secrets.pass" label="pass" />
        </v-col>
      </template>
      <v-col cols="12" md="4">
        <v-text-field readonly :value="secrets.insecure ? 'yes' : 'no'" label="insecure (skip TLS verify)" />
      </v-col>
      <v-col cols="12" md="4">
        <v-text-field readonly :value="secrets.ca_cert ? 'set' : '-'" label="ca_cert" />
      </v-col>
      <v-col cols="12" v-if="!secrets.user">
        <v-alert dense text type="warning">
          VNC console needs user + pass: PVE vncwebsocket rejects API tokens.
        </v-alert>
      </v-col>
    </v-row>
    <slot v-else></slot>
  </div>
</template>

<script>
import PasswordTextField from "@/components/ui/passwordTextField.vue";

export default {
  name: "service-provider-secrets-proxmox",
  components: { PasswordTextField },
  props: { template: { type: Object, required: true } },
  data: () => ({ editing: false }),
  computed: {
    secrets() {
      return this.template.secrets || {};
    },
  },
};
</script>
