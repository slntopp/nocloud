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
    fetch({ commit }) {
      commit("setAddons", []);
      commit("setLoading", true);
      return new Promise((resolve, reject) => {
        api
          .get("/addons")
          .then((response) => {
            commit("setAddons", response.addons);
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          })
          .finally(() => {
            commit("setLoading", false);
          });
      });
    },
    fetchById({ commit }, id) {
      commit("setLoading", true);
      return new Promise((resolve, reject) => {
        api
          .get(`addons/${id}`)
          .then((response) => {
            commit("setOne", response);
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          })
          .finally(() => {
            commit("setLoading", false);
          });
      });
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
