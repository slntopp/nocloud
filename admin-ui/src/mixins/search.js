const searchMixin = {
  beforeDestroy() {
    this.$store.commit("appSearch/resetSearch");
  },
};

export default searchMixin;
