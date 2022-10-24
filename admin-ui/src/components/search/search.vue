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
        hide-details
        prepend-inner-icon="mdi-magnify"
        placeholder="Search..."
        single-line
        background-color="background-light"
        dence
        :value="getAllTags"
        rounded
        readonly
        v-on="on"
      ></v-text-field>
    </template>
    <component :is="searchMenuComponent">
      <v-text-field
        hide-details
        prepend-inner-icon="mdi-magnify"
        placeholder="Search..."
        single-line
        dence
        color="white"
        background-color="background-light"
        v-model="searchParam"
      ></v-text-field>
    </component>
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
    getAllTags() {
      return [this.searchParam, ...this.tags].join(", ");
    },
    isAdvancedSearch() {
      return this.$store.getters["appSearch/isAdvancedSearch"];
    },
    tags() {
      return this.$store.getters["appSearch/getTags"];
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
