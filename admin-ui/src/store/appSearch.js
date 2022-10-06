export default {
  namespaced: true,
  state: {
    searchParam: "",
  },
  mutations: {
    setSearchParam(state, newSearchParam) {
      state.searchParam = newSearchParam;
    },
    resetSearchParam(state) {
      state.searchParam = "";
    },
  },
  getters: {
    param(state) {
      return state.searchParam;
    },
  },
};
