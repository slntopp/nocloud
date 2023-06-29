import Vue from "vue";
import Vuex from "vuex";

// modules
import auth from "./auth";
import namespaces from "./namespaces";
import accounts from "./accounts";
import servicesProviders from "./servicesProviders";
import dns from "./dns";
import settings from "./settings";
import reloadBtn from "./reloadbtn";
import services from "./services";
import plans from "./plans";
import transactions from "./transactions";
import appSearch from "./appSearch";
import vnc from "./vnc";
import currencies from "./currencies";
import plugins from "./plugins";
import snackbar from "./snackbar";
import actions from "./actions";

Vue.use(Vuex);

const store = new Vuex.Store({
  state: {},
  mutations: {},
  actions: {},
  modules: {
    auth,
    namespaces,
    accounts,
    servicesProviders,
    dns,
    settings,
    services,
    reloadBtn,
    plans,
    transactions,
    appSearch,
    vnc,
    currencies,
    plugins,
    snackbar,
    actions,
  },
});

export default store;

export const useStore = () => store;
