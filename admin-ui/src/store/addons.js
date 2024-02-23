import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    addons: [],
    one: null,
    loading: false,
  },
  mutations: {
    setAddons(state, addons) {
      state.addons = addons;
    },
    setOne(state, addon) {
      state.one = addon;
    },
    setLoading(state, data) {
      state.loading = data;
    },
  },
  actions: {
    async fetch({ commit }, options) {
      commit("setAddons", []);
      commit("setLoading", true);
      try {
        const response = await api.post("/billing/addons", options);

        commit("setAddons", response.addons);
      } finally {
        commit("setLoading", false);
      }
    },
    async count(_, options) {
      return api.post("/billing/count/addons", options);
    },
    async fetchById({ commit }, id) {
      commit("setLoading", true);
      try {
        const response = await api.get(`/billing/addons/${id}`);
        commit("setOne", response);
      } finally {
        commit("setLoading", false);
      }
    },
  },
  getters: {
    all(state) {
      return state.addons;
    },
    one(state) {
      return state.one;
    },
    isLoading(state) {
      return state.loading;
    },
  },
};
