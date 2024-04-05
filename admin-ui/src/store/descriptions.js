import { createPromiseClient } from "@connectrpc/connect";
import { DescriptionsService } from "nocloud-proto/proto/es/billing/billing_connect";

export default {
  namespaced: true,
  state: {},
  mutations: {},
  actions: {
    update({ getters }, data) {
      return getters["descriptionsClient"].update(data);
    },
    create({ getters }, data) {
      return getters["descriptionsClient"].create(data);
    },
    get({ getters }, uuid) {
      return getters["descriptionsClient"].get({ uuid });
    },
  },
  getters: {
    descriptionsClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(
        DescriptionsService,
        rootGetters["app/transport"]
      );
    },
  },
};
