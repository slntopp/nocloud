export default {
  namespaced: true,
  state: {
    theme: "dark",
  },
  mutations: {
    setTheme(state, theme = "dark") {
      state.theme = theme;
    },
  },
  getters: {
    theme(state) {
      return state.theme;
    },
  },
};
