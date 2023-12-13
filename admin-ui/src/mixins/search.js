const searchMixin = ({name,defaultLayout}) => ({
  beforeDestroy() {
    this.$store.commit("appSearch/setSearchName", "");
    this.$store.commit("appSearch/setFields", []);
    this.$store.commit("appSearch/setDefaultLayout", null);
  },
  mounted() {
    this.$store.commit("appSearch/setSearchName", name);
    this.$store.commit("appSearch/setDefaultLayout", defaultLayout);
  },
});

export default searchMixin;
