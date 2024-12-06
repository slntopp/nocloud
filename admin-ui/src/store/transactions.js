import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    transactions: [],
    transaction: [],
    loading: false,
    count: 0,
    itemPerPage: 10,
    page: 1,
    filter: {
      field: "",
      sort: "",
    },
  },
  getters: {
    all(state, getters, rootState, rootGetters) {
      if (
        rootGetters["currencies/rates"]?.length &&
        state.transactions.length
      ) {
        return state.transactions.map((t) => {
          if (
            t.currency?.code === rootGetters["currencies/default"]?.code ||
            t.currency?.code == "NCU"
          ) {
            return t;
          }
          const rate = rootGetters["currencies/rates"].find(
            (r) =>
              r.from?.code === t.currency?.code &&
              r.to?.code === rootGetters["currencies/default"]?.code
          )?.rate;
          t.total = t.total * rate;
          return t;
        });
      }
      return state.transactions;
    },
    one(state) {
      return state.transaction;
    },
    isLoading(state) {
      return state.loading;
    },
    count(state) {
      return +state.count;
    },
    page(state) {
      return +state.page;
    },
    types() {
      return [
        { key: "transaction top-up", title: "Top-up transaction" },
        { key: "transaction payment", title: "Debit transaction" },
        { key: "system", title: "Debit transaction (System)" },
        {
          key: "transaction account-balance",
          title: "Set balance transaction",
        },
        { key: "invoice top-up", title: "Top-up invoice" },
        { key: "invoice payment", title: "Payment invoice" },
        { key: "invoice for service", title: "Invoice for service" },
        { key: "correct", title: "Correction" },
      ];
    },
  },
  mutations: {
    setTransactions(state, transactions) {
      state.transactions = transactions;
    },
    setTransaction(state, transaction) {
      state.transaction = transaction;
    },
    setLoading(state, data) {
      state.loading = data;
    },
    setCount(state, count) {
      state.count = count;
    },
    setPage(state, page) {
      state.page = page;
    },
    setFilter(state, filter) {
      state.filter = filter;
    },
    setItemPerPage(state, val) {
      state.itemPerPage = val;
    },
  },
  actions: {
    init({ commit }, data) {
      commit("setCount", 0);
      commit("setLoading", true);
      return api.transactions
        .count(data)
        .then((data) => {
          commit("setCount", data.total);
        })
        .finally(() => {
          commit("setLoading", false);
        });
    },
    fetch({ commit, state }, data) {
      commit("setTransactions", []);
      commit("setLoading", true);
      return api.transactions
        .list({
          limit: state.itemPerPage,
          page: state.page,
          field: state.filter.field || "proc",
          sort: state.filter.sort || "desc",
          ...data,
        })
        .then((data) => {
          commit("setTransactions", data.pool);
        })
        .finally(() => {
          commit("setLoading", false);
        });
    },
    changeFiltres({ commit, dispatch }, { options, data }) {
      commit("setPage", options.page);
      commit("setFilter", {
        field: options.sortBy[0],
        sort: options.sortDesc[0] ? "desc" : "asc",
      });
      commit("setItemPerPage", options.itemsPerPage);
      return dispatch("fetch", data);
    },
    fetchById({ commit }, params) {
      commit("setTransaction", []);
      commit("setLoading", true);

      return new Promise((resolve, reject) => {
        api.transactions
          .list(params)
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
