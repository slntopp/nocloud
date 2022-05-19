import Api from "nocloudjsrest";
import vuex from "@/store/index.js";
const api = new Api();

api.axios.interceptors.response.use(
  function (response) {
    return response;
  },
  function (error) {
    if (error.response && error.response?.data?.code === 7) {
      // console.log("credentials are not actual");
      vuex.dispatch("auth/logout");
    }
    return Promise.reject(error); // this is the important part
  }
);

export default api;
