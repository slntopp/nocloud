import { computed } from "vue";
import { useStore } from "@/store";
import { Rounding } from "nocloud-proto/proto/es/billing/billing_pb";

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

    const endPrice = rate ? price * rate : 0;

    const precision = currency.precision || 0;
    const rounding = currency.rounding || "ROUND_HALF";

    if (endPrice == 0) {
      return 0;
    }

    if (endPrice < 0.01 && endPrice > -1) {
      return parseFloat(endPrice.toFixed(10));
    }

    if (Rounding.ROUND_HALF === Rounding[rounding]) {
      return +endPrice.toFixed(precision).toString();
    }

    const fn =
      Rounding[rounding] === Rounding.ROUND_DOWN ? Math.floor : Math.round;

    return fn(endPrice * Math.pow(10, precision)) / Math.pow(10, precision);
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
