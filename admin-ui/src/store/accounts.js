import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    accounts: [],
    loading: false,
  },
  mutations: {
    setAccounts(state, accounts) {
      state.accounts = accounts;
    },
    pushAccount(state, account) {
      const index = state.accounts.findIndex((a) => a.uuid === account.uuid);

      if (index !== -1) {
        state.accounts[index] = account;
      } else {
        state.accounts.push(account);
      }
    },
    setLoading(state, data) {
      state.loading = data;
    },
  },
  actions: {
    fetch({ commit }) {
      commit("setAccounts", []);
      commit("setLoading", true);
      return new Promise((resolve, reject) => {
        api.accounts
          .list()
          .then((response) => {
            commit("setAccounts", response.pool);
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
        api.accounts
          .get(id)
          .then((response) => {
            commit("pushAccount", response);
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
      return state.accounts;
    },
    isLoading(state) {
      return state.loading;
    },
  },
};
