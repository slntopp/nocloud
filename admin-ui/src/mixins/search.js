const searchMixin = (name) => ({
  beforeDestroy() {
    this.$store.commit("appSearch/setSearchName", "");
    this.$store.commit("appSearch/setFields", []);
  },
  mounted() {
    this.$store.commit("appSearch/setSearchName", name);
  },
});

export default searchMixin;
