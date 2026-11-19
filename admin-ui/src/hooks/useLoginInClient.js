import { useStore } from "@/store";

/** The client cabinet's address from the `app` setting, or "" until it is known. */
const appUrlOf = (setting) => {
  try {
    return JSON.parse(setting?.app ?? "{}").url ?? "";
  } catch {
    return "";
  }
};

const useLoginInClient = () => {
  const store = useStore();

  /**
   * Opens the client's cabinet signed in as the client. The window opens at once, on the click
   * (later the browser may block it), so the cabinet loads while the client's token is issued. The
   * token goes to that window once both are there: the token, and the cabinet's "ready".
   */
  const loginHandler = ({ accountUuid, instanceId, chatId, type }) => {
    const knownUrl = appUrlOf(store.state.auth.appURL);
    const win = window.open(knownUrl || "", "_blank");
    let token = "";
    let ready = false;

    const send = () => {
      if (!token || !ready || !win || win.closed) return;
      window.removeEventListener("message", onMessage);
      win.postMessage({ token, uuid: instanceId, chatId, type }, "*");
    };
    function onMessage(event) {
      if (event.source !== win || event.data !== "ready") return;
      ready = true;
      send();
    }
    window.addEventListener("message", onMessage);

    Promise.all([
      store.dispatch("auth/loginToApp", { uuid: accountUuid, type: "whmcs" }),
      store.dispatch("auth/getAppURL"),
    ])
      .then(([response, setting]) => {
        if (!knownUrl && win) {
          win.location = appUrlOf(setting);
        }
        token = response.token;
        send();
      })
      .catch((e) => {
        window.removeEventListener("message", onMessage);
        win?.close();
        store.commit("snackbar/showSnackbarError", {
          message: e.response?.data?.message || "Error during login",
        });
      });
  };

  return { loginHandler };
};

export default useLoginInClient;
