import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    servicesProviders: [],
    loading: false,
  },
  getters: {
    all(state) {
      return state.servicesProviders;
    },
    isLoading(state) {
      return state.loading;
    },
  },
  mutations: {
    setServicesProviders(state, servicesProviders) {
      state.servicesProviders = servicesProviders;
    },
    setLoading(state, data) {
      state.loading = data;
    },
  },
  actions: {
    fetch({ commit }) {
      commit("setLoading", true);
      return new Promise((resolve, reject) => {
        api.servicesProviders
          .list()
          .then((response) => {
            commit("setServicesProviders", response.pool);
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
};
