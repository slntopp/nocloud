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
  <v-autocomplete
    no-filter
    @change="changeValue"
    :search-input.sync="inputValue"
    @update:search-input="changeSearchInput"
    :items="searchItems"
    item-text="title"
    item-value="key"
    v-else
  >
    <template v-slot:item="{ item }">
      <span v-if="!selectedGroupKey">{{
        `Search ${inputValue} in ${item.title}`
      }}</span>
      <span v-else>{{ item.title }}</span>
    </template>
  </v-autocomplete>
</template>

<script>
import { mapGetters } from "vuex";

export default {
  name: "app-search",
  data: () => ({ inputValue: "", selectedGroupKey: "" }),
  methods: {
    changeValue(e) {
      console.log(e, this.inputValue);
      this.selectedGroupKey = e;
    },
    changeSearchInput(e) {
      this.inputValue = e;
      console.log(e);
    },
  },
  computed: {
    ...mapGetters("appSearch", {
      isAdvancedSearch: "isAdvancedSearch",
      variants: "variants",
    }),
    searchParam: {
      get() {
        return this.$store.getters["appSearch/param"];
      },
      set(newValue) {
        this.$store.commit("appSearch/setSearchParam", newValue);
      },
    },
    searchItems() {
      return (
        this.variants[this.selectedGroupKey]?.items ||
        Object.keys(this.variants).map((key) => ({
          key,
          title: this.variants[key]?.title || key,
        }))
      );
    },
  },
};
</script>
