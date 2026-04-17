import api from "@/api.js";
import { createPromiseClient } from "@connectrpc/connect";
import { AccountsService } from "nocloud-proto/proto/es/registry/registry_connect";

export default {
  namespaced: true,
  state: {
    accounts: [],
    total: 0,
    loading: false,
    one: {},
    currentRequestId: null,
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
    async fetch({ commit, state }, params) {
      const requestId = Date.now();
      state.currentRequestId = requestId;

      if (!params.silent) commit("setLoading", true);

      try {
        const response = await api.post("accounts", params);

        if (requestId !== state.currentRequestId) {
          return;
        }

        commit("setAccounts", response.pool);
        commit("setTotal", response.count);
      } finally {
        if (requestId === state.currentRequestId) {
          commit("setLoading", false);
        }
      }
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
    accountsClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(AccountsService, rootGetters["app/transport"]);
    },
  },
};
