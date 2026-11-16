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
    updateService(state, service) {
      const index = state.servicesProviders.findIndex(
        (serv) => serv.uuid === service.uuid
      );
      if (index === -1) {
        state.servicesProviders = [...state.servicesProviders, service];
      } else {
        state.servicesProviders = state.servicesProviders.map((serv) =>
          serv.uuid === service.uuid ? service : serv
        );
      }
    },
    setLoading(state, data) {
      state.loading = data;
    },
  },
  actions: {
    fetch({ commit }, params) {
      commit("setServicesProviders", []);
      commit("setLoading", true);
      return new Promise((resolve, reject) => {
        api.servicesProviders
          .list(params)
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
    fetchById({ commit }, id) {
      commit("setLoading", true);
      return new Promise((resolve, reject) => {
        api.servicesProviders
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
