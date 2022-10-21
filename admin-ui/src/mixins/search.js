const searchMixin = {
  mounted(){
    this.$store.commit("appSearch/resetSearchParams");
  }
};

export default searchMixin;
