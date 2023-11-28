import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    currenciesList: ["NCU", "USD", "EUR", "BYN", "PLN"],
    currencies: [],
    currency: {},
    defaultCurrency: "",
    loading: false,
  },
  getters: {
    all(state) {
      return state.currenciesList;
    },
    one(state) {
      return state.currency;
    },
    rates(state) {
      return state.currencies;
    },
    default(state) {
      return state.defaultCurrency;
    },
    isLoading(state) {
      return state.loading;
    },
  },
  mutations: {
    setCurrencies(state, currencies) {
      state.currenciesList = currencies;
    },
    setCurrency(state, currency) {
      state.currency = currency;
    },
    setRates(state, rates) {
      state.currencies = rates.map((el) => ({
        ...el,
        id: `${el.from} ${el.to}`,
      }));
    },
    setDefault(state, currencies) {
      const currency = currencies.find(
        (el) => el.rate === 1 && [el.from, el.to].includes("NCU")
      );

      if (!currency) return;
      state.defaultCurrency =
        currency.from === "NCU" ? currency.to : currency.from;
    },
    setLoading(state, data) {
      state.loading = data;
    },
    updateCurrency(state, newCurrency) {
      state.currency = state.currency.map((currency) =>
        newCurrency.id === currency.id ? newCurrency : currency
      );
    },
  },
  actions: {
    fetch({ commit, state }, options) {
      if (state.loading) return;
      if (!options?.silent) {
        commit("setRates", []);
        commit("setDefault", []);
        commit("setCurrencies", []);
        commit("setLoading", true);
      }

      return new Promise((resolve, reject) => {
        api
          .get("/billing/currencies")
          .then((response) => {
            commit("setCurrencies", response.currencies);
            return api.get("/billing/currencies/rates");
          })
          .then((response) => {
            commit("setRates", response.rates);
            commit("setDefault", response.rates);
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          })
          .finally(() => {
            commit("setLoading", false);
          });
      });
    },
    fetchById({ commit }, { from, to }) {
      commit("setLoading", true);

      return new Promise((resolve, reject) => {
        api
          .get(`/billing/currencies/${from}/${to}`)
          .then((response) => {
            commit("updateCurrency", response);
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          })
          .finally(() => {
            commit("setLoading", false);
          });
      });
    },
    fetchItem({ commit }, { from, to }) {
      commit("setLoading", true);

      return new Promise((resolve, reject) => {
        api
          .get(`/billing/currencies/${from}/${to}`)
          .then((response) => {
            commit("setCurrency", response);
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          })
          .finally(() => {
            commit("setLoading", false);
          });
      });
    },
  },
};
