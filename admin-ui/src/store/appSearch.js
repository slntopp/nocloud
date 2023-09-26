export default {
  namespaced: true,
  state: {
    searchParam: "",
    searchName: "",
    variants: {},
    customParams: {},
  },
  mutations: {
    setSearchParam(state, newSearchParam) {
      state.searchParam = newSearchParam;
    },
    setVariants(state, val) {
      const variants = {};
      Object.keys(val).forEach((key) => {
        const items = val[key].items?.map((i) => ({
          title: i.title || i,
          uuid: i.uuid || i,
        }));
        variants[key] = { key, ...val[key], items };
      });
      state.variants = variants;
    },
    setSearchName(state, val) {
      state.searchName = val;
    },
    resetSearch(state) {
      state.searchName = "";
    },
    pushVariant(state, { key, value }) {
      state.variants = { ...state.variants, [key]: value };
    },
    resetSearchParams(state) {
      state.searchParam = "";
      state.variants = {};
      state.customParams = {};
    },
    setCustomParams(state, params) {
      state.customParams = params;
    },
    setCustomParam(state, { key, value }) {
      if ((key === "searchParam" || !key) && !value.value) {
        return;
      }
      state.customParams = {
        ...state.customParams,
        [key]: !value.isArray
          ? value
          : [...(state.customParams[key] || []), value],
      };
    },
    deleteCustomParam(state, { key, value, isArray }) {
      if (value && isArray) {
        state.customParams[key] = state.customParams[key].filter(
          (v) => v.value !== value
        );
      } else {
        delete state.customParams[key];
        state.customParams = { ...state.customParams };
      }
    },
  },
  getters: {
    param(state) {
      return state.searchParam;
    },
    customSearchParam(state) {
      return state.customParams.searchParam?.value;
    },
    variants(state) {
      const variants = { ...state.variants };
      if (Object.keys(variants).length) {
        variants["searchParam"] = { title: "Other", key: "searchParam" };
      }
      return variants;
    },
    searchName(state) {
      return state.searchName + "_search";
    },
    customParams(state) {
      const params = { ...state.customParams };
      delete params["searchParam"];
      return params;
    },
  },
};
