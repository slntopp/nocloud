import { createConnectTransport } from "@connectrpc/connect-web";
import { useStore } from "@/store/index";

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
        // On localhost the admin is not served next to the API. Price models
        // and other Connect calls must use the same proxy target as api.js.
        baseUrl:
          window.location.hostname === "localhost"
            ? "http://localhost:8624/https://api.nc2dev.support.by"
            : window.location.origin,
        useBinaryFormat: true,
        interceptors: [
          (next) => async (req) => {
            req.header.set(
              "Authorization",
              `Bearer ${rootGetters["auth/token"]}`
            );
            req.header.set("X-Requested-With", "XMLHttpRequest");
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
