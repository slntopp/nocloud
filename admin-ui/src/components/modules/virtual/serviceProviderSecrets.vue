<template>
  <div class="pa-4">
    <v-row v-if="!editing">
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.group"
          label="group"
          style="display: inline-block; width: 330px"
        >
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.host"
          label="host"
          style="display: inline-block; width: 330px"
        >
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.user"
          label="user"
          style="display: inline-block; width: 330px"
        >
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.pass"
          label="password"
          style="display: inline-block; width: 330px"
          :type="showPassword ? 'text' : 'password'"
          :append-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
          @click:append="() => (showPassword = !showPassword)"
        >
        </v-text-field>
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
