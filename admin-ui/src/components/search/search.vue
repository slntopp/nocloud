<template>
  <v-text-field
    v-if="!isAdvancedSearch"
    hide-details
    prepend-inner-icon="mdi-magnify"
    placeholder="Search..."
    single-line
    background-color="background-light"
    dence
    v-model="searchParam"
    rounded
  ></v-text-field>
  <v-menu v-else :close-on-content-click="false">
    <template v-slot:activator="{ on }">
      <v-text-field
        v-if="!isAdvancesSearch"
        hide-details
        prepend-inner-icon="mdi-magnify"
        placeholder="Search..."
        single-line
        background-color="background-light"
        dence
        rounded
        readonly
        v-on="on"
      ></v-text-field>
    </template>
    <component :is="searchMenuComponent" />
  </v-menu>
</template>

<script>
export default {
  name: "app-search",
  computed: {
    searchParam: {
      get() {
        return this.$store.getters["appSearch/param"];
      },
      set(newValue) {
        this.$store.commit("appSearch/setSearchParam", newValue);
      },
    },
    isAdvancedSearch() {
      return this.$store.getters["appSearch/isAdvancedSearch"];
    },
    searchMenuName() {
      return this.$store.getters["appSearch/searchMenuName"];
    },
    searchMenuComponent() {
      return () =>
        import(`@/components/search/menus/${this.searchMenuName}.vue`);
    },
  },
};
</script>
