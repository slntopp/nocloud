import { createPromiseClient } from "@connectrpc/connect";
import { CurrencyService } from "nocloud-proto/proto/es/billing/billing_connect";
import {
  GetCurrenciesRequest,
  GetExchangeRatesRequest,
} from "nocloud-proto/proto/es/billing/billing_pb";
import { createListCache } from "@/store/cache.js";

// currencies change through the connect client, which the cache doesn't see:
// the currencies page calls fetch() after its changes
const currenciesCache = createListCache();

const load = async (commit, getters, options) => {
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
    commit("setDefault", currencies);
    const { rates } = await getters["currencyClient"].getExchangeRates(
      new GetExchangeRatesRequest()
    );
    commit("setRates", rates);
  } finally {
    commit("setLoading", false);
  }
};

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
      const currency = currencies.find((el) => el.default);

      if (!currency) return;
      state.defaultCurrency = currency;
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
    // Always goes to the server (reload button, after changes)
    fetch({ commit, getters }, options) {
      return currenciesCache.run(
        null,
        () => load(commit, getters, options),
        { force: true }
      );
    },
    // Reuses the loaded currencies and rates if they're fresh
    ensure({ commit, getters }) {
      return currenciesCache.run(null, () => load(commit, getters));
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
