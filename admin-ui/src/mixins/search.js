const searchMixin = {
  mounted() {
    this.$store.commit("appSearch/resetSearchParams");
  },
  beforeDestroy() {
    this.$store.commit("appSearch/setAdvancedSearch", null);
  },
  methods: {
    setAdvancedSearch(name) {
      this.$store.commit("appSearch/setAdvancedSearch", name);
    },
  },
};

export default searchMixin;
