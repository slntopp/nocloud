import api from "@/api.js";
import Cookies from "js-cookie";
import router from "@/router";

const COOKIES_NAME = "noCloud-token";

export default {
  namespaced: true,
  state: {
    token: "",
    userdata: {},
    appURL: ""
  },
  mutations: {
    setToken(state, token) {
      const expires = new Date(Date.now() + 7776e6);

      state.token = token;
      Cookies.set(COOKIES_NAME, token, { expires });
      // document.cookie = `${COOKIES_NAME}=${token}; path=/; expires=${expires}`;
    },
    setUserdata(state, data) {
      state.userdata = data;
    },
    setAppURL(state, data) {
      state.appURL = data;
    }
  },
  actions: {
    login({ commit }, { login, password, type }) {
      return new Promise((resolve, reject) => {
        api.authorizeCustom({
            auth: { data: [login, password], type },
            exp: Math.round((Date.now() + 7776e6) / 1000),
            root_claim: true
          })
          .then((response) => {
            api.applyToken(response.token);
            commit("setToken", response.token);
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          });
      });
    },
    getAppURL({ state, commit }) {
      return new Promise((resolve, reject) => {
        if (state.appURL !== "") {
          resolve(state.appURL);
          return;
        }

        api.settings.get(["app"])
          .then((response) => {
            commit("setAppURL", response)
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          });
      });
    },
    loginToApp(_, { type, uuid }) {
      return new Promise((resolve, reject) => {
        api.authorizeCustom({ auth: { type, data: [] }, uuid })
          .then((response) => {
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          });
      });
    },
    logout({ commit }) {
      commit("setToken", "");
      Cookies.remove(COOKIES_NAME);
      router.push({ name: "Login" });
    },

    load({ commit }) {
      const token = Cookies.get(COOKIES_NAME);
      if (token) {
        api.axios.defaults.headers.common["Authorization"] = "Bearer " + token;
        commit("setToken", token);
      }
    },

    fetchUserData({ commit }) {
      commit;
      return new Promise((resolve, reject) => {
        api.accounts
          .get("me")
          .then((response) => {
            commit("setUserdata", response);
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          });
      });
    },
  },
  getters: {
    isLoggedIn(state) {
      return state.token.length > 0;
    },
    userdata(state) {
      return state.userdata;
    },
    token(state) {
      return state.token;
    },
  },
};
