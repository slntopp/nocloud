import { createPromiseClient } from "@connectrpc/connect";
import { BillingService } from "nocloud-proto/proto/es/billing/billing_connect";
import {
  CreateInvoiceRequest,
  GetInvoicesCountRequest,
  GetInvoicesRequest,
} from "nocloud-proto/proto/es/billing/billing_pb";

export default {
  namespaced: true,
  state: {
    invoices: [],
    one: null,
    loading: false,
    currentFetchRequestId: null,
    currentCountRequestId: null,
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
    incrementFetchRequestId(state) {
      state.currentFetchRequestId++;
    },
    incrementCountRequestId(state) {
      state.currentCountRequestId++;
    },
  },
  actions: {
    async fetch({ commit, state, getters }, params) {
      commit("incrementFetchRequestId");
      const requestId = state.currentFetchRequestId;

      commit("setLoading", true);
      try {
        const response = await getters["invoicesClient"].getInvoices(
          GetInvoicesRequest.fromJson(params),
        );

        if (requestId !== state.currentFetchRequestId) return;

        commit("setInvoices", response.toJson().pool);
      } finally {
        if (requestId === state.currentFetchRequestId) {
          commit("setLoading", false);
        }
      }
    },
    async count({ state, getters, commit }, params) {
      commit("incrementCountRequestId");
      const requestId = state.currentCountRequestId;
      const response = await getters["invoicesClient"].getInvoicesCount(
        GetInvoicesCountRequest.fromJson(params),
      );

      if (requestId !== state.currentCountRequestId) return null;
      return response;
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
    copy({ getters }, invoice) {
      delete invoice.meta.whmcs_invoice_id;
      const data = {
        items: invoice.items,
        total: invoice.total,
        account: invoice.account,
        type: invoice.type,
        deadline: Math.round(Date.now() / 1000 + 86400 * 30),
        meta: invoice.meta,
        status: "DRAFT",
        currency: invoice.currency,
      };

      return getters["invoicesClient"].createInvoice(
        CreateInvoiceRequest.fromJson({
          invoice: data,
          isSendEmail: false,
        }),
      );
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
