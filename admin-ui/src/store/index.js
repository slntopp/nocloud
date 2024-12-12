import Vue from "vue";
import Vuex from "vuex";

// modules
import app from "./app";
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
import showcases from "./showcases";
import chats from "./chats";
import addons from "./addons";
import invoices from "./invoices";
import descriptions from "./descriptions";
import instances from "./instances";
import promocodes from "./promocodes";

Vue.use(Vuex);

const store = new Vuex.Store({
  state: {},
  mutations: {},
  actions: {},
  modules: {
    app,
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
    showcases,
    chats,
    addons,
    invoices,
    descriptions,
    instances,
    promocodes,
  },
});

export default store;

export const useStore = () => store;
