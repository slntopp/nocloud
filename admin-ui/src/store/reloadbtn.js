export default {
  namespaced: true,
  state: {
    loading: false,
    onclick: {
      type: null,
      params: null,
      event:null
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
    setCallback(state, { type, params,event }) {
      state.onclick.type = type;
      state.onclick.params = params !== undefined ? params : null;
      state.onclick.event = event
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
      if(state.onclick.event){
        state.onclick.event()
        return
      }
      if (!state.onclick.type) return;
      commit("setLoading", true);
      dispatch(state.onclick.type, state.onclick.params, {
        root: true,
      });
      commit("setLoading", false);
    },
  },
};
