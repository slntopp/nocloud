import api from "@/api.js";
export default {
  namespaced: true,
  state: {
    plugins: [],
    loading: false,
  },
  getters: {
    all(state) {
      return state.plugins;
    },
    isLoading(state) {
      return state.loading;
    },
  },
  mutations: {
    setPlugins(state, plugins) {
      state.plugins = plugins;
    },
  },
  actions: {
    fetch({ commit }) {
      api.settings.get(["plugins"]).then((res) => {
        const key = res["plugins"];
        if (key) commit("setPlugins", JSON.parse(key));
      });
    },
  },
};
