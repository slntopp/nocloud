<template>
  <v-text-field
    hide-details
    prepend-inner-icon="mdi-magnify"
    placeholder="Search..."
    single-line
    background-color="background-light"
    dence
    v-model="searchParam"
    rounded
  ></v-text-field>
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
  },
};
</script>
