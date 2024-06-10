import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    namespaces: [],
    one: null,
    loading: false,
  },
  mutations: {
    setNamespaces(state, namespaces) {
      state.namespaces = namespaces;
    },
    setLoading(state, data) {
      state.loading = data;
    },
    setOne(state, value) {
      state.one = value;
    },
  },
  actions: {
    async fetch({ commit }, options) {
      commit("setLoading", true);
      commit("setNamespaces", []);
      try {
        const response = await api.post("namespaces", options);
        commit("setNamespaces", response.pool);
        return response;
      } finally {
        commit("setLoading", false);
      }
    },
    async fetchById({ commit }, id) {
      commit("setLoading", true);
      commit("setOne", null);
      try {
        const response = await api.get(`namespaces/${id}`);
        commit("setOne", response);
        return response;
      } finally {
        commit("setLoading", false);
      }
    },
  },
  getters: {
    all(state) {
      return state.namespaces.map((n) => ({
        ...n,
        title: n.title.startsWith("NS_") ? n.title : `NS_${n.title}`,
      }));
    },
    isLoading(state) {
      return state.loading;
    },
    one(state) {
      return state.one;
    },
  },
};
