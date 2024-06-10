import { createPromiseClient } from "@connectrpc/connect";
import { CurrencyService } from "nocloud-proto/proto/es/billing/billing_connect";
import {
  GetCurrenciesRequest,
  GetExchangeRatesRequest,
} from "nocloud-proto/proto/es/billing/billing_pb";

export default {
  namespaced: true,
  state: {
    currenciesList: [{ id: 1, title: "NCU" }],
    currencies: [],
    currency: {},
    defaultCurrency: "",
    loading: false,
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
        id: `${el.from.id} ${el.to.id}`,
      }));
    },
    setDefault(state, currencies) {
      const currency = currencies.find(
        (el) => el.rate === 1 && [el.from.title, el.to.title].includes("NCU")
      );

      if (!currency) return;
      state.defaultCurrency =
        currency.from.title === "NCU" ? currency.to : currency.from;
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
    async fetch({ commit, state, getters }, options) {
      if (state.loading) return;
      if (!options?.silent) {
        commit("setRates", []);
        commit("setDefault", []);
        commit("setCurrencies", []);
        commit("setLoading", true);
      }

      try {
        const { currencies } = await getters["currencyClient"].getCurrencies(
          new GetCurrenciesRequest()
        );
        commit("setCurrencies", currencies);
        const { rates } = await getters["currencyClient"].getExchangeRates(
          new GetExchangeRatesRequest()
        );
        commit("setRates", rates);
        commit("setDefault", rates);
      } finally {
        commit("setLoading", false);
      }
    },
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
    currencyClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(CurrencyService, rootGetters["app/transport"]);
    },
  },
};
