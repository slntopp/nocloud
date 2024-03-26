import { createConnectTransport } from "@connectrpc/connect-web";
import { useStore } from "./index";

export default {
  namespaced: true,
  state: {
    theme: "dark",
    chatClicks: 0
  },
  mutations: {
    setTheme(state, theme = "dark") {
      state.theme = theme;
    },
    setChatClicks(state, value) {
      state.chatClicks += value;
    }
  },
  getters: {
    theme: (state) => state.theme,
    chatClicks: (state) => state.chatClicks,
    transport(state, getters, rootState, rootGetters) {
      const transport = createConnectTransport({
        baseUrl: window.location.origin,
        useBinaryFormat: true,
        interceptors: [
          (next) => async (req) => {
            req.header.set(
              "Authorization",
              `Bearer ${rootGetters["auth/token"]}`
            );
            return next(req);
          },

          (next) => async (req) => {
            try {
              return await next(req);
            } catch (err) {
              if (
                err.response &&
                err.response?.data?.code === 7 &&
                !err.response?.config?.url?.includes("transactions") &&
                !err.response?.config?.url?.includes("services")
              ) {
                // console.log("credentials are not actual");
                const store = useStore();
                store.dispatch("auth/logout");
              }
              return Promise.reject(err);
            }
          },
        ],
      });

      return transport;
    },
  },
};
