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

    if (currency?.code == defaultCurrency.value?.code) {
      rate = 1;
    } else {
      rate = rates.value.find(
        (r) =>
          [defaultCurrency.value?.code, "NCU"].includes(r.to.code) &&
          r.from.code == currency?.code
      )?.rate;
    }

    return rate ? (price * rate).toFixed(2) : 0;
  };

  const convertTo = (price, currency) => {
    let rate;

    if (currency?.code === defaultCurrency.value?.code) {
      rate = 1;
    } else {
      rate = rates.value.find(
        (r) =>
          r.to.code == currency?.code &&
          [defaultCurrency.value?.code, "NCU"].includes(r.from.code)
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
