<template>
  <div class="pa-4">
    <v-row v-if="!editing">
      <v-col>
        <v-text-field readonly :value="template.secrets.host" label="host" />
      </v-col>
      <v-col>
        <v-text-field readonly :value="template.secrets.token_id" label="token_id" />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.token_secret"
          label="token_secret"
          :type="showPassword ? 'text' : 'password'"
          :append-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
          @click:append="() => (showPassword = !showPassword)"
        />
      </v-col>
    </v-row>
    <slot v-else></slot>
  </div>
</template>

<script>
export default {
  props: { template: { type: Object, required: true } },
  data: () => ({ editing: false, showPassword: false }),
};
</script>
