import { createPromiseClient } from "@connectrpc/connect";
import { AddonsService } from "nocloud-proto/proto/es/billing/billing_connect";
import {
  CountAddonsRequest,
  ListAddonsRequest,
} from "nocloud-proto/proto/es/billing/addons/addons_pb";

export default {
  namespaced: true,
  state: {
    addons: [],
    one: null,
    loading: false,

    currentFetchRequestId: 0,
    currentCountRequestId: 0,
  },
  mutations: {
    setAddons(state, addons) {
      state.addons = addons;
    },
    setOne(state, addon) {
      state.one = addon;
    },
    setLoading(state, data) {
      state.loading = data;
    },
  },
  actions: {
    async fetch({ commit, state, getters }, options) {
      const id = ++state.currentFetchRequestId;
      commit("setLoading", true);

      try {
        const response = await getters.addonsClient.list(
          ListAddonsRequest.fromJson(options),
        );
        if (id !== state.currentFetchRequestId) return;

        commit("setAddons", response.addons);
      } finally {
        if (id === state.currentFetchRequestId) commit("setLoading", false);
      }
    },

    async count({ state, getters }, options) {
      const id = ++state.currentCountRequestId;
      try {
        const response = await getters.addonsClient.count(
          CountAddonsRequest.fromJson(options),
        );
        if (id !== state.currentCountRequestId) return null;

        return response;
      } catch (e) {
        return null;
      }
    },
    async fetchById({ commit, getters }, id) {
      commit("setLoading", true);
      try {
        const response = await getters.addonsClient.get({ uuid: id });
        commit("setOne", response.toJson());
      } finally {
        commit("setLoading", false);
      }
    },
  },
  getters: {
    addonsClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(AddonsService, rootGetters["app/transport"]);
    },
    all(state) {
      return state.addons;
    },
    one(state) {
      return state.one;
    },
    isLoading(state) {
      return state.loading;
    },
  },
};
