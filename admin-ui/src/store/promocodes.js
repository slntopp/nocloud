import { createPromiseClient } from "@connectrpc/connect";
import { PromocodesService } from "nocloud-proto/proto/es/billing/billing_connect";
import {
  CountPromocodesRequest,
  ListPromocodesRequest,
  Promocode,
} from "nocloud-proto/proto/es/billing/promocodes/promocodes_pb";

export default {
  namespaced: true,
  state: {
    promocodes: [],
    one: null,
    loading: false,
  },
  mutations: {
    setPromocodes(state, promocodes) {
      state.promocodes = promocodes;
    },
    setOne(state, promocode) {
      state.one = promocode;
    },
    setLoading(state, data) {
      state.loading = data;
    },
  },
  actions: {
    async fetch({ commit, getters }, options) {
      commit("setPromocodes", []);
      commit("setLoading", true);
      try {
        const response = await getters.promocodesClient.list(
          ListPromocodesRequest.fromJson(options)
        );
        commit("setPromocodes", response.toJson().promocodes);
      } finally {
        commit("setLoading", false);
      }
    },
    async get({ commit, getters }, uuid) {
      commit("setOne", null);
      commit("setLoading", true);
      try {
        const response = await getters.promocodesClient.get(
          Promocode.fromJson({ uuid })
        );
        commit("setOne", response.toJson());
      } finally {
        commit("setLoading", false);
      }
    },
    async count({ getters }, options) {
      return getters.promocodesClient.count(
        CountPromocodesRequest.fromJson(options)
      );
    },
  },
  getters: {
    promocodesClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(
        PromocodesService,
        rootGetters["app/transport"]
      );
    },
    all(state) {
      return state.promocodes;
    },
    one(state) {
      return state.one;
    },
    isLoading(state) {
      return state.loading;
    },
  },
};
