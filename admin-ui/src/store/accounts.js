import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    accounts: [],
    total: 0,
    loading: false,
    one: {},
  },
  mutations: {
    setAccounts(state, accounts) {
      state.accounts = accounts;
    },
    setTotal(state, total) {
      state.total = +total;
    },
    setOne(state, account) {
      state.one = account;
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
    fetch({ commit }, params) {
      commit("setAccounts", []);
      if (!params.silent) {
        commit("setLoading", true);
      }

      return new Promise((resolve, reject) => {
        api
          .post("accounts", params)
          .then((response) => {
            commit("setAccounts", response.pool);
            commit("setTotal", response.count);
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
      return state.accounts;
    },
    one(state) {
      return state.one;
    },
    total(state) {
      return state.total;
    },
    isLoading(state) {
      return state.loading;
    },
  },
};
