import Api from "nocloudjsrest";
import vuex from "@/store/index.js";
const api = new Api();

// const api = new Api(
//   "http://localhost/https://api.nocloud.ione-cloud.net",
//   8624
// );

api.axios.interceptors.response.use(
  function (response) {
    return response;
  },
  function (error) {
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
