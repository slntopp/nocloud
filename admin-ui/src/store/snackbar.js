export default {
  namespaced: true,
  state: {
    visibility: false,
    message: "",
    timeout: 3000,
    route: {},
    color: "",
    buttonColor: "primary",
  },
  mutations: {
    showSnackbar(
      state,
      {
        message,
        timeout = 3000,
        route = {},
        color = "",
        buttonColor = "primary",
      }
    ) {
      state.message = message;
      state.timeout = timeout;
      state.route = route;
      state.visibility = true;
      state.color = color;
      state.buttonColor = buttonColor;
    },
    hideSnackbar(state) {
      state.visibility = false;
    },
    showSnackbarError(_, { message, timeout }) {
      const opts = {
        message,
        timeout,
        color: "red darken-3",
        buttonColor: "white",
      };
      this.commit("snackbar/showSnackbar", opts);
    },
    showSnackbarSuccess(_, { message, timeout }) {
      const opts = {
        message,
        timeout,
        color: "green darken-3",
        buttonColor: "white",
      };
      this.commit("snackbar/showSnackbar", opts);
    },
    setVisibility(state, val) {
      state.visibility = val;
    },
  },
  getters: {
    visibility(state) {
      return state.visibility;
    },
    message(state) {
      return state.message;
    },
    timeout(state) {
      return state.timeout;
    },
    route(state) {
      return state.route;
    },
    color(state) {
      return state.color;
    },
    buttonColor(state) {
      return state.buttonColor;
    },
  },
};
