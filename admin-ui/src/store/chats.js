import { createPromiseClient } from "@connectrpc/connect";
import {
  ChatsAPI,
  StreamService,
} from "core-chatting/plugin/src/connect/cc/cc_connect";
import {
  Empty,
  StreamRequest,
} from "core-chatting/plugin/src/connect/cc/cc_pb";

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
    replaceChat(state, value) {
      state.chats = state.chats.map((chat) =>
        chat.uuid === value.uuid ? value : chat
      );
    },
    pushChat(state, value) {
      state.chats = [...state.chats, value];
    },
  },
  getters: {
    all(state) {
      return state.chats;
    },
    chatsClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(ChatsAPI, rootGetters["app/transport"]);
    },
    chatsStreamClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(StreamService, rootGetters["app/transport"]);
    },
    unreadChatsCount(state) {
      return state.chats.filter(
        (chat) => chat.meta.unread > 0 && [0, 1, 5, 6, 8].includes(chat.status)
      ).length;
    },
  },
  actions: {
    async fetch({ getters, commit, state }) {
      commit("setLoading", true);
      try {
        const data = await getters["chatsClient"].list(Empty.fromJson({}));
        commit("setChats", data.chats);

        for await (const { type, item } of getters["chatsStreamClient"].stream(
          new StreamRequest()
        )) {
          switch (type) {
            case 4: {
              commit("replaceChat", {
                ...item.value,
                meta: { ...item.value.meta, unread: 0 },
              });
              break;
            }
            case 5: {
              const chat = state.chats.find(
                (chat) => chat.uuid === item.value.chat
              );
              commit("replaceChat", {
                ...chat,
                meta: { ...chat.meta, unread: 1 },
              });
              break;
            }
            case 1: {
              commit("pushChat", item.value);
              break;
            }
          }
        }
      } finally {
        commit("setLoading", false);
      }
    },
  },
};
