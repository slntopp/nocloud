export default {
  namespaced: true,
  state: {
    loading: false,
    onclick: {
      type: null,
      params: null,
    },
    btnStates: {
      disabled: false,
      visible: true,
    },
  },
  getters: {
    isLoading(state) {
      return state.loading;
    },
    states(state) {
      return state.btnStates;
    },
  },
  mutations: {
    setCallback(state, { type, params }) {
      state.onclick.type = type;
      state.onclick.params = params !== undefined ? params : null;
    },
    clear(state) {
      state.onclick = { type: null, params: null };
    },
    setLoading(state, data) {
      state.loading = data;
    },
    setState(state, { stateName, value }) {
      this.state.btnStates[stateName] = value;
    },
  },
  actions: {
    async onclick({ state, commit, dispatch }) {
      if (!state.onclick.type) return;
      commit("setLoading", true);
      dispatch(state.onclick.type, state.onclick.params, {
        root: true,
      });
      commit("setLoading", false);
    },
  },
};
