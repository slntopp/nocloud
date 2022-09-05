<template>
  <div>
    <!-- Secrets -->
    <v-card-title class="px-0 mb-3"> Secrets:</v-card-title>
    <v-row v-if="!editing">
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.app_key"
          label="app key"
          style="display: inline-block; width: 330px"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.app_secret"
          label="app secret"
          style="display: inline-block; width: 330px"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.consumer_key"
          label="consumer key"
          style="display: inline-block; width: 330px"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.secrets.endpoint"
          label="region"
          style="display: inline-block; width: 330px"
        />
      </v-col>
    </v-row>

    <slot></slot>
  </div>
</template>

<script>
export default {
  name: 'service-provider-ovh',
  props: {
    template: { type: Object, required: true },
    editing: { type: Boolean, required: true }
  }
}
</script>
