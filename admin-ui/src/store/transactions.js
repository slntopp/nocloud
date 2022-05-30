import api from '@/api.js';

export default {
  namespaced: true,
  state: {
    transactions: [],
    transaction: [],
    loading: false
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
    }
  },
  mutations: {
    setTransactions(state, transactions) {
      state.transactions = transactions
        .reduce((acc, item) => [...acc, ...item.pool], []);
    },
    setTransaction(state, transaction) {
      state.transaction = transaction;
    },
    setLoading(state, data) {
      state.loading = data;
    }
  },
  actions: {
    fetch({ commit }, accounts) {
      commit('setLoading', true);

      return new Promise((resolve, reject) => {
        const promises = accounts.map((id) =>
          api.get('/billing/transactions', {
            params: { account: id }
          })
        );

        Promise.all(promises)
          .then((response) => {
            commit('setTransactions', response);
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          })
          .finally(() => {
            commit('setLoading', false);
          });
      });
    },
    fetchById({ commit }, id) {
      commit('setLoading', true);

      return new Promise((resolve, reject) => {
        api.get('/billing/transactions', {
          params: { account: id }
        })
          .then((response) => {
            commit('setTransaction', response);
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          })
          .finally(() => {
            commit('setLoading', false);
          });
      });
    }
  },
};
