const searchMixin = {
  beforeDestroy() {
    this.$store.commit("appSearch/resetSearchParams");
  },
};

export default searchMixin;
