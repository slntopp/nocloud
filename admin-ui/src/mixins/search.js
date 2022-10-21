const searchMixin = {
  mounted(){
    this.$store.commit("appSearch/resetSearchParams");
  },
  methods:{
    setAdvancedSearch(name){
      this.$store.commit('appSearch/setAdvancedSearch',name)
    }
  }
};

export default searchMixin;
