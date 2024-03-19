import { createPromiseClient } from "@connectrpc/connect";
import { ChatsAPI } from "core-chatting/plugin/src/connect/cc/cc_connect";
import { Empty } from "core-chatting/plugin/src/connect/cc/cc_pb";

export default {
  namespaced: true,
  state: {
    chats: [],
    loding: false,
  },
  mutations: {
    setChats(state, value) {
      state.chats = value;
    },
    setLoading(state, value) {
      state.loading = value;
    },
  },
  getters: {
    all(state) {
      return state.chats;
    },
    chatsClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(ChatsAPI, rootGetters["app/transport"]);
    },
  },
  actions: {
    async fetch({ getters, commit }) {
      commit("setLoading", true);
      try {
        const data = await getters["chatsClient"].list(Empty.fromJson({}));
        commit("setChats", data.chats);
      } finally {
        commit("setLoading", false);
      }
    },
  },
};
