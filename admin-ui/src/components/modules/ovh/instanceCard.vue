<template>
  <div>
    <v-row align="center">

    </v-row>
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
          :type="isVisible ? 'text' : 'password'"
          :value="template.state.meta.password"
          :append-icon="isVisible ? 'mdi-eye' : 'mdi-eye-off'"
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
      <v-col v-for="k in configKeys" :key="k.key">
        <v-text-field
          readonly
          style="display: inline-block; width: 200px"
          :label="dictionary[k.key] ?? k.key"
          :value="getConfigValue(k.path)"
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
    configKeys: [
      { key: "datacenter", path: null },
      { key: "os", path: null },
      { key: "type", path: "type" },
    ],
    dataKeys: ["vpsId", "creation", "expiration"],

  }),
  mounted() {
    this.initPrices();
    this.initConfigsKeys();
    this.getBasePrices();
  },
  methods: {

    initConfigsKeys() {
      this.configKeys.forEach((k) => {
        if (!k.path) {
          k.path = this.getKeyFromConfiguration(k.key);
        }
      });
    },
    getKeyFromConfiguration(name) {
      for (const key of Object.keys(this.template.config.configuration))
        if (key.includes(name)) {
          return key;
        }
    },
    getConfigValue(path) {
      return (
        this.template.config[path] || this.template.config.configuration[path]
      );
    },


  },
  computed: {

    resources() {
      return this.tarrif.resources;
    },

  },
};
</script>
