import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    addons: [],
    loading: false,
  },
  mutations: {
    setAddons(state, addons) {
      state.addons = addons;
    },
    pushAddon(state, addon) {
      const index = state.addons.findIndex((a) => a.uuid === addon.uuid);

      if (index !== -1) {
        state.addons[index] = addon;
      } else {
        state.addons.push(addon);
      }
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
    // fetchById({ commit }, id) {
    //     commit("setLoading", true);
    //     return new Promise((resolve, reject) => {
    //         api.accounts
    //             .get(id)
    //             .then((response) => {
    //                 commit("pushAccount", response);
    //                 resolve(response);
    //             })
    //             .catch((error) => {
    //                 reject(error);
    //             })
    //             .finally(() => {
    //                 commit("setLoading", false);
    //             });
    //     });
    // },
  },
  getters: {
    all(state) {
      return state.addons;
    },
    isLoading(state) {
      return state.loading;
    },
  },
};
