import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    invoices: [],
    one: null,
    loading: false,
  },
  mutations: {
    setInvoices(state, invoices) {
      state.invoices = invoices;
    },
    setOne(state, invoice) {
      state.one = invoice;
    },
    setLoading(state, data) {
      state.loading = data;
    },
  },
  actions: {
    fetch({ commit }) {
      commit("setInvoices", []);
      commit("setLoading", true);
      return new Promise((resolve, reject) => {
        api
          .get("/billing/invoices")
          .then((response) => {
            commit("setInvoices", response.pool);
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
        api
          .get(`/billing/invoices/${id}`)
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
      return state.invoices;
    },
    one(state) {
      return state.one;
    },
    isLoading(state) {
      return state.loading;
    },
  },
};
