<template>
  <div>
    <h3 v-if="dense">Data:</h3>
    <v-card-title v-else class="px-0">Data:</v-card-title>
    <v-row align="center">
      <v-col v-for="key in dataKeys" :key="key">
        <v-text-field
          readonly
          style="display: inline-block; width: 200px"
          :label="dictionary[key] ?? key"
          :value="template.data[key]"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="login"
          style="display: inline-block; width: 200px"
          :value="template.state.meta.login"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="password"
          style="display: inline-block; width: 200px"
          :type="(isVisible) ? 'text' : 'password'"
          :value="template.state.meta.password"
          :append-icon="(isVisible) ? 'mdi-eye' : 'mdi-eye-off'"
          @click:append="isVisible = !isVisible"
        />
      </v-col>
    </v-row>

    <h3 v-if="dense">Resources:</h3>
    <v-card-title v-else class="px-0">Resources:</v-card-title>
    <v-row align="center">
      <v-col v-for="(item, key) in resources" :key="key">
        <v-text-field
          readonly
          style="display: inline-block; width: 200px"
          :label="dictionary[key] ?? key"
          :value="item"
        />
      </v-col>
      <v-col v-for="key in configKeys" :key="key">
        <v-text-field
          readonly
          style="display: inline-block; width: 200px"
          :label="dictionary[key] ?? key"
          :value="template.config[key]"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script>
export default {
  name: "instance-card",
  props: {
    template: { type: Object, required: true },
    dense: { type: Boolean },
  },
  data: () => ({
    isVisible: false,
    dictionary: {
      cpu: "CPU",
      ram: "RAM",
      os: "OS",
      vpsId: "id",
    },
    configKeys: ["datacenter", "os", "type"],
    dataKeys: ["vpsId", "creation", "expiration"],
  }),
  computed: {
    resources() {
      const { duration, planCode } = this.template.config
      const key = `${duration} ${planCode}`;

      return this.template.billingPlan.products[key].resources;
    }
  }
}
</script>
