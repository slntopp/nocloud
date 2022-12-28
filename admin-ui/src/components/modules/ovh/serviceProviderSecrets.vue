<template>
  <div class="pa-4">
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

    <slot v-else></slot>
  </div>
</template>

<script>
export default {
  props: { template: { type: Object, required: true } },
  data: () => ({ editing: false })
}
</script>
