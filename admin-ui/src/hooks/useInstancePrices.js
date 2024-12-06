import { computed } from "vue";
import useCurrency from "@/hooks/useCurrency";

const useInstancePrices = (instance, account) => {
  const { convertFrom, convertTo, rates, defaultCurrency } = useCurrency();

  const accountCurrency = computed(() => account?.currency);
  const accountRate = computed(() => {
    if (defaultCurrency.value?.code == accountCurrency.value?.code) {
      return 1;
    }
    return rates.value.find(
      (r) =>
        r.to?.code == accountCurrency.value?.code &&
        r.from?.code == defaultCurrency.value?.code
    );
  });

  const toAccountPrice = (price) => {
    return convertTo(price, accountCurrency.value);
  };
  const fromAccountPrice = (price) => {
    return convertFrom(price, accountCurrency.value);
  };

  return {
    toAccountPrice,
    fromAccountPrice,
    account,
    accountCurrency,
    accountRate,
    rates,
  };
};

export default useInstancePrices;
