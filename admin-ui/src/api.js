import Api from "nocloudjsrest";
import vuex from "@/store/index.js";
const api = new Api('http://localhost:8624/https://api.nc2dev.support.by/');

api.axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (
      error.response &&
      error.response?.data?.code === 7 &&
      !error.response?.config?.url?.includes("transactions") &&
      !error.response?.config?.url?.includes("services")
    ) {
      // console.log("credentials are not actual");
      vuex.dispatch("auth/logout");
    }
    return Promise.reject(error); // this is the important part
  }
);

export default api;
