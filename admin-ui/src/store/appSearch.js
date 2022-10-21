export default {
  namespaced: true,
  state: {
    searchParam: "",
    isAdvancedSearch: false,
    advancedSearchParams: null,
    searchMenuName:''
  },
  mutations: {
    setSearchParam(state, newSearchParam) {
      state.searchParam = newSearchParam;
    },
    setAdvancedSearch(state,searchMenuName){
      state.isAdvancedSearch=true
      state.searchMenuName=searchMenuName
    },
    resetSearchParams(state) {
      state.isAdvancedSearch = false;
      state.searchParam = "";
      state.advancedSearchParams = null;
    },
  },
  getters: {
    param(state) {
      return state.searchParam;
    },
    isAdvancedSearch(state){
      return state.isAdvancedSearch && state.searchMenuName
    },
    advancedParams(state){
      return state.advancedSearchParams
    },
    searchMenuName(state){
      return state.searchMenuName
    }
  },
};
