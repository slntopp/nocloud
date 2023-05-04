export default {
  namespaced: true,
  state: {
    searchParam: "",
    variants: {},
    customParams: {},
  },
  mutations: {
    setSearchParam(state, newSearchParam) {
      state.searchParam = newSearchParam;
    },
    setVariants(state, val) {
      state.variants = val;
    },
    resetSearchParams(state) {
      state.searchParam = "";
      state.variants = {};
      state.customParams = {};
    },
    setCustomParam(state, { key, value }) {
      // state.customParams[key] = value;
      state.customParams={...state.customParams,[key]:value}
    },
    deleteCustomParam(state, key) {
      delete state.customParams[key];
      state.customParams={...state.customParams}
    },
  },
  getters: {
    param(state) {
      return state.searchParam;
    },
    variants(state) {
      return state.variants;
    },
    customParams(state) {
      return state.customParams;
    },
  },
};
