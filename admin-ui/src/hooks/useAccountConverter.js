import api from "@/api";
import { computed, ref, watch } from "vue";
import { useStore } from "@/store";

const useAccountConverter = (instance) => {
  const store = useStore();

  const accountRate = ref(0);

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

  const accountCurrency = computed(() => account.value.currency);

  const defaultCurrency = computed(() => {
    return store.getters["currencies/default"];
  });

  const toAccountPrice = (price) => {
    return accountRate.value ? (price / accountRate.value).toFixed(2) : 0;
  };
  const fromAccountPrice = (price) => {
    return accountRate.value ? (price * accountRate.value).toFixed(2) : 0;
  };

  const fetchAccountRate = async () => {
    if (!account.value.currency || !defaultCurrency.value) {
      return;
    }
    if (account.value.currency === defaultCurrency.value) {
      accountRate.value = 1;
      return;
    }
    const res = await api.get(
      `/billing/currencies/rates/${account.value.currency}/${defaultCurrency.value}`
    );
    accountRate.value = res.rate;
  };

  watch(accountCurrency, () => {
    fetchAccountRate();
  });

  watch(defaultCurrency, () => {
    fetchAccountRate();
  });

  return {
    toAccountPrice,
    fromAccountPrice,
    fetchAccountRate,
    account,
    accountCurrency,
    accountRate,
  };
};

export default useAccountConverter;
