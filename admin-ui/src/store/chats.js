import { createPromiseClient } from "@connectrpc/connect";
import {
  ChatsAPI,
  StreamService,
} from "core-chatting/plugin/src/connect/cc/cc_connect";
import {
  ListChatsRequest,
  StreamRequest,
} from "core-chatting/plugin/src/connect/cc/cc_pb";
import {
  endOfMonth,
  endOfWeek,
  startOfMonth,
  startOfWeek,
  startOfDay,
  endOfDay,
} from "date-fns";

export default {
  namespaced: true,
  state: {
    dayChats: [],
    weekChats: [],
    monthChats: [],
    loading: false,
  },
  mutations: {
    setDayChats(state, value) {
      state.dayChats = value;
    },
    setWeekChats(state, value) {
      state.weekChats = value;
    },
    setMonthChats(state, value) {
      state.monthChats = value;
    },
    setLoading(state, value) {
      state.loading = value;
    },
    replaceChat(state, value) {
      state.dayChats = state.dayChats.map((chat) =>
        chat.uuid === value.uuid ? value : chat
      );
    },
    pushChat(state, value) {
      state.dayChats = [...state.dayChats, value];
    },
  },
  getters: {
    dayChats(state) {
      return state.dayChats;
    },
    weekChats(state) {
      return state.weekChats;
    },
    monthChats(state) {
      return state.monthChats;
    },
    chatsClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(ChatsAPI, rootGetters["app/transport"]);
    },
    chatsStreamClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(StreamService, rootGetters["app/transport"]);
    },
    unreadChatsCount(state) {
      return state.dayChats.filter(
        (chat) => chat.meta.unread > 0 && [0, 1, 5, 6, 8].includes(chat.status)
      ).length;
    },
  },
  actions: {
    async fetch({ getters, commit, state }) {
      commit("setLoading", true);
      try {
        const baseReqParams = {
          limit: 1000,
          page: 1,
          field: "updated",
          sort: "desc",
        };
        const [day, week, month] = await Promise.all([
          getters["chatsClient"].list(
            ListChatsRequest.fromJson({
              ...baseReqParams,
              filters: {
                created: {
                  from: startOfDay(new Date()).getTime(),
                  to: endOfDay(new Date()).getTime(),
                },
              },
            })
          ),
          getters["chatsClient"].list(
            ListChatsRequest.fromJson({
              ...baseReqParams,
              filters: {
                created: {
                  from: startOfWeek(new Date()).getTime(),
                  to: endOfWeek(new Date()).getTime(),
                },
              },
            })
          ),
          getters["chatsClient"].list(
            ListChatsRequest.fromJson({
              ...baseReqParams,
              filters: {
                created: {
                  from: startOfMonth(new Date()).getTime(),
                  to: endOfMonth(new Date()).getTime(),
                },
              },
            })
          ),
        ]);

        commit("setDayChats", day.toJson()?.pool || []);
        commit("setWeekChats", week.toJson()?.pool || []);
        commit("setMonthChats", month.toJson()?.pool || []);

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
