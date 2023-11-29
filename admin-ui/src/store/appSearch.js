export default {
  namespaced: true,
  state: {
    fields: [],
    currentLayout: "",
    searchName: "",
    filter: {},
    param: "",
  },
  mutations: {
    setFields(state, fields) {
      state.fields = fields;
    },
    pushFields(state, fields) {
      state.fields=[...state.fields,...fields]
    },
    setCurrentLayout(state, name) {
      state.currentLayout = name;
    },
    setSearchName(state, name) {
      state.searchName = name;
    },
    setFilter(state, filter) {
      state.filter = filter;
    },
    setParam(state, param) {
      state.param = param;
    },
    setFilterValue(state, { key, value }) {
      state.filter[key] = value;
    },
  },
  getters: {
    param(state) {
      return state.param;
    },
    fields(state) {
      return state.fields;
    },
    currentLayout(state) {
      return state.currentLayout;
    },
    filter(state) {
      return state.filter;
    },
    searchName(state) {
      return state.searchName;
    },
  },
};
