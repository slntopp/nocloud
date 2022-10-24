export default {
  namespaced: true,
  state: {
    searchParam: "",
    isAdvancedSearch: false,
    advancedSearchParams: null,
    searchMenuName: "",
    tags: [],
  },
  mutations: {
    setSearchParam(state, newSearchParam) {
      state.searchParam = newSearchParam;
    },
    setAdvancedSearch(state, searchMenuName) {
      state.searchMenuName = searchMenuName;
      state.isAdvancedSearch = true;
    },
    setAdvancedParams(state, params) {
      state.advancedSearchParams = params;
    },
    setTags(state, tags) {
      state.tags = [...tags];
    },
    resetSearchParams(state) {
      state.isAdvancedSearch = false;
      state.searchParam = "";
      state.advancedSearchParams = null;
      state.searchMenuName = "";
      state.tags = [];
    },
  },
  getters: {
    param(state) {
      return state.searchParam;
    },
    isAdvancedSearch(state) {
      return state.isAdvancedSearch && state.searchMenuName;
    },
    advancedParams(state) {
      return { advanced: state.advancedSearchParams, param: state.searchParam };
    },
    searchMenuName(state) {
      return state.searchMenuName;
    },
    getTags(state) {
      return state.tags;
    },
  },
};
