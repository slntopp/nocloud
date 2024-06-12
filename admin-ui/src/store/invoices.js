import { createPromiseClient } from "@connectrpc/connect";
import { BillingService } from "nocloud-proto/proto/es/billing/billing_connect";
import {
  GetInvoicesCountRequest,
  GetInvoicesRequest,
} from "nocloud-proto/proto/es/billing/billing_pb";

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
    async fetch({ commit, getters }, params) {
      commit("setInvoices", []);
      commit("setLoading", true);
      try {
        const response = await getters["invoicesClient"].getInvoices(
          GetInvoicesRequest.fromJson(params)
        );
        commit("setInvoices", response.toJson().pool);
      } finally {
        commit("setLoading", false);
      }
    },
    count({ getters }, params) {
      return getters["invoicesClient"].getInvoicesCount(
        GetInvoicesCountRequest.fromJson(params)
      );
    },
    async get({ commit, getters }, uuid) {
      commit("setLoading", true);
      try {
        const response = await getters["invoicesClient"].getInvoice({ uuid });
        commit("setOne", response.toJson());
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
    invoicesClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(BillingService, rootGetters["app/transport"]);
    },
  },
};
