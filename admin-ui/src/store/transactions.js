import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    transactions: [],
    transaction: [],
    loading: false,
  },
  getters: {
    all(state) {
      return state.transactions;
    },
    one(state) {
      return state.transaction;
    },
    isLoading(state) {
      return state.loading;
    },
  },
  mutations: {
    setTransactions(state, transactions) {
      state.transactions = transactions.reduce(
        (acc, item) => [...acc, ...item.pool],
        []
      );
    },
    setTransaction(state, transaction) {
      state.transaction = transaction;
    },
    setLoading(state, data) {
      state.loading = data;
    },
  },
  actions: {
    fetch({ commit }, { accounts, service }) {
      commit("setLoading", true);

      return new Promise((resolve, reject) => {
        const promises = accounts.map((account) =>
          api.transactions.get({ account, service })
        );

        Promise.all(promises)
          .then((response) => {
            commit("setTransactions", response);
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
    fetchById({ commit }, params) {
      commit("setLoading", true);

      return new Promise((resolve, reject) => {
        api.transactions
          .get(params)
          .then((response) => {
            commit("setTransaction", response.pool);
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
