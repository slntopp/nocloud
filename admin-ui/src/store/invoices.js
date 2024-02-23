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
    async fetch({ commit }) {
      commit("setInvoices", []);
      commit("setLoading", true);
      try {
        const response = await api.post("/billing/invoices");
        commit("setInvoices", response.pool);
      } finally {
        commit("setLoading", false);
      }
    },
    async fetchById({ commit }, id) {
      commit("setLoading", true);
      try {
        const response = await api.get(`/billing/invoices/${id}`);
        commit("setOne", response);
      } finally {
        commit("setLoading", false);
      }
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
