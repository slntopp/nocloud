export default {
  namespaced: true,
  state: {
    searchParam: "",
    isAdvancedSearch: false,
    variants: {},
  },
  mutations: {
    setSearchParam(state, newSearchParam) {
      state.searchParam = newSearchParam;
    },
    setAdvancedSearch(state, val) {
      state.isAdvancedSearch = val;
    },
    setVariants(state, val) {
      state.variants = val;
    },
    resetSearchParams(state) {
      state.isAdvancedSearch = false;
      state.searchParam = "";
      state.variants = [];
    },
  },
  getters: {
    param(state) {
      return state.searchParam;
    },
    isAdvancedSearch(state) {
      return state.isAdvancedSearch;
    },
    variants(state) {
      return state.variants;
    },
  },
};
