import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    services: [],
    loading: false,
  },
  getters: {
    all(state) {
      return state.services;
    },
    isLoading(state) {
      return state.loading;
    },
  },
  mutations: {
    setServices(state, services) {
      state.services = services;
    },
    setLoading(state, data) {
      state.loading = data;
    },
    updateService(state, service) {
      state.services = state.services.map((serv) =>
        serv.uuid === service.uuid ? service : serv
      );
    },
  },
  actions: {
    fetch({ commit }) {
      commit("setLoading", true);
      return new Promise((resolve, reject) => {
        api.services
          .list()
          .then((response) => {
            commit("setServices", response.pool);
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
        api.services
          .get(id)
          .then((response) => {
            commit("updateService", response);
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
