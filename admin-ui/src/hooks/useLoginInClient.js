import { useStore } from "@/store";

const useLoginInClient = () => {
  const store = useStore();

  const loginHandler = ({ accountUuid, instanceId, type }) => {
      console.log(accountUuid,instanceId,type)
    store
      .dispatch("auth/loginToApp", { uuid: accountUuid, type: "whmcs" })
      .then(({ token }) => {
        store.dispatch("auth/getAppURL").then((res) => {
          const win = window.open(JSON.parse(res.app).url);

          window.addEventListener("message", () => {
            win.postMessage({ token, uuid: instanceId, type: type }, "*");
          });
        });
      })
      .catch((e) => {
        store.commit("snackbar/showSnackbarError", {
          message: e.response?.data?.message || "Error during login",
        });
      });
  };

  return { loginHandler };
};

export default useLoginInClient