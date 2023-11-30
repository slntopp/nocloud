import { computed } from "vue";
import { useStore } from "@/store";
import useCurrency from "@/hooks/useCurrency";

const useInstancePrices = (instance) => {
  const store = useStore();
  const { convertFrom, convertTo, rates, defaultCurrency, } = useCurrency();

  const namespace = computed(() =>
    store.getters["namespaces/all"]?.find(
      (n) => n.uuid === instance.access.namespace
    )
  );

  const account = computed(() => {
    if (!namespace.value) {
      return;
    }
    return store.getters["accounts/all"]?.find(
      (a) => a?.uuid === namespace.value.access.namespace
    );
  });

  const accountCurrency = computed(() => account.value?.currency);
  const accountRate = computed(() => {
    if (defaultCurrency.value === accountCurrency.value) {
      return 1;
    }
    return rates.value.find(
      (r) => r.to === accountCurrency.value && r.from === defaultCurrency.value
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
    rates
  };
};

export default useInstancePrices;
