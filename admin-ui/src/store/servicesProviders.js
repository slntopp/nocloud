import api from "@/api.js";
import { createListCache } from "@/store/cache.js";

const spCache = createListCache({ invalidateOn: /^\/?sp(\/|\?|$)/ });

const load = async (commit, params) => {
  commit("setServicesProviders", []);
  commit("setLoading", true);
  try {
    const response = await api.servicesProviders.list(params);
    commit("setServicesProviders", response.pool);
    return response;
  } finally {
    commit("setLoading", false);
  }
};

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
    // Always goes to the server (reload button, after changes)
    fetch({ commit }, params) {
      return spCache.run(params, () => load(commit, params), { force: true });
    },
    // Reuses the loaded list if it's fresh and was loaded with the same params
    ensure({ commit }, params) {
      return spCache.run(params, () => load(commit, params));
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
