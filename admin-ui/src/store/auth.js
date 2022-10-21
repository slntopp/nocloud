import api from "@/api.js";
import Cookies from "js-cookie";
import router from "@/router";

const COOKIES_NAME = "noCloud-token";

export default {
  namespaced: true,
  state: {
    token: "",
    userdata: {},
  },
  mutations: {
    setToken(state, token) {
      state.token = token;
    },
    setUserdata(state, data) {
      state.userdata = data;
    },
  },
  actions: {
    login({ commit }, { login, password, type }) {
      return new Promise((resolve, reject) => {
        api
          .authorizeWithType(login, password, type, true)
          .then((response) => {
            api.applyToken(response.token);
            Cookies.set(COOKIES_NAME, response.token);
            commit("setToken", response.token);
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
  },
};
