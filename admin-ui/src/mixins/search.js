const searchMixin = {
  mounted(){
    this.$store.commit("appSearch/resetSearchParam");
  }
};

export default searchMixin;
