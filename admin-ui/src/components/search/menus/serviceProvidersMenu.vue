<template>
  <v-card class="px-5" color="background-light">
    <v-card-title>service providers search</v-card-title>
    <slot></slot>
    <v-select
      item-color="dark"
      class="mt-5"
      multiple
      label="service provider type"
      v-model="selectedTypes"
      @change="setAdvancedParams"
      :items="spTypes"
    ></v-select>
  </v-card>
</template>

<script>
export default {
  name: "service-provider-search-menu",
  data() {
    return { spTypes: ["all"], selectedTypes: [] };
  },
  created() {
    this.spTypes.push(...this.getBaseSptypes());
  },
  methods: {
    getBaseSptypes() {
      const rightTypes = [];

      const types = require.context(
        "@/components/modules/",
        true,
        /serviceProviders\.vue$/
      );

      types.keys().forEach((key) => {
        const matched = key.match(
          /\.\/([A-Za-z0-9-_,\s]*)\/serviceProviders\.vue/i
        );

        const type = matched[1];

        if (matched && matched.length > 1) {
          rightTypes.push(type);
        }
      });

      return rightTypes;
    },
    setAdvancedParams(types) {
      this.$store.commit("appSearch/setAdvancedParams", { types });
      this.$store.commit("appSearch/setTags", types);
    },
  },
};
</script>
