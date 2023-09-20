import { computed } from "vue";
import { useStore } from "@/store";

const useCurrency = () => {
  const store = useStore();

  const defaultCurrency = computed(() => {
    return store.getters["currencies/default"];
  });

  const rates = computed(() => {
    return store.getters["currencies/rates"];
  });

  const convertFrom = (price, currency) => {
    let rate;

    if (currency === defaultCurrency.value) {
      rate = 1;
    } else {
      rate = rates.value.find(
        (r) => r.to === defaultCurrency.value && r.from === currency
      )?.rate;
    }

    return rate ? (price / rate).toFixed(2) : 0;
  };

  const convertTo = (price, currency) => {
    let rate;

    if (currency === defaultCurrency.value) {
      rate = 1;
    } else {
      rate = rates.value.find(
        (r) => r.to === currency && r.from === defaultCurrency.value
      )?.rate;
    }

    return rate ? (price * rate).toFixed(2) : 0;
  };

  return {
    convertFrom,
    rates,
    defaultCurrency,
    convertTo,
  };
};

export default useCurrency;
