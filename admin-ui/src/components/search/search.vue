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
        rounded
        readonly
        v-on="on"
      >
      <template v-slot:label>
        <v-chip style="height:20px" class="mr-1" small color="gray" v-for="tag in getAllTags" :key="tag">
          {{ tag }}
        </v-chip>
      </template>
      </v-text-field>
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
      if (!this.searchParam) {
        return this.tags;
      }
      return [this.searchParam, ...this.tags];
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
      if(!this.searchMenuName){
        return  null
      }
      return () =>
        import(`@/components/search/menus/${this.searchMenuName}.vue`);
    },
  },
};
</script>
