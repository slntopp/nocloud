const snackbar = {
  methods: {
    showSnackbarSuccess({ message, timeout }) {
      this.$store.commit("snackbar/showSnackbarSuccess", { message, timeout });
    },
    showSnackbarError({ message, timeout }) {
      this.$store.commit("snackbar/showSnackbarError", { message, timeout });
    },
    showSnackbar({
      message,
      timeout = 3000,
      route = {},
      color = "",
      buttonColor = "primary",
    }) {
      this.$store.commit("snackbar/showSnackbar", {
        message,
        timeout,
        route,
        color,
        buttonColor,
      });
    },
    hideSnackbar() {
      this.$store.commit("snackbar/hideSnackbar");
    },
  },
};

export default snackbar;