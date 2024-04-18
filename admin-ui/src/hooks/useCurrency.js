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

    if (currency?.title === defaultCurrency.value?.title) {
      rate = 1;
    } else {
      rate = rates.value.find(
        (r) =>
          r.to.title === defaultCurrency.value?.title &&
          r.from.title === currency?.title
      )?.rate;
    }

    return rate ? (price * rate).toFixed(2) : 0;
  };

  const convertTo = (price, currency) => {
    let rate;

    if (currency?.title === defaultCurrency.value?.title) {
      rate = 1;
    } else {
      rate = rates.value.find(
        (r) =>
          r.to.title === currency?.title &&
          r.from.title === defaultCurrency.value?.title
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
