import { computed, onMounted, ref, watch } from "vue";
import { useStore } from "@/store";

const useRate = (currency = ref()) => {
  const store = useStore();

  const rate = ref(0);

  const defaultCurrency = computed(() => {
    return store.getters["currencies/default"];
  });

  onMounted(() => {
    setRate();
  });

  const setRate = async (fetchedCurrency) => {
    if (!fetchedCurrency) {
      fetchedCurrency = currency.value;
    }

    rate.value = store.getters["currencies/rates"].find(
      (rate) =>
        rate.from.title === fetchedCurrency.title &&
        rate.to.title === defaultCurrency.value.title
    ).rate;
    return rate.value;
  };

  watch(currency, () => {
    setRate();
  });

  watch(defaultCurrency, () => {
    setRate();
  });

  const convertFrom = (price, convertedRate) => {
    if (!convertedRate) {
      convertedRate = rate.value;
    }
    return convertedRate ? (price / convertedRate).toFixed(2) : 0;
  };
  const convertTo = (price, convertedRate) => {
    if (!convertedRate) {
      convertedRate = rate.value;
    }
    return convertedRate ? (price * convertedRate).toFixed(2) : 0;
  };

  return {
    rate,
    convertTo,
    convertFrom,
    setRate,
  };
};

export default useRate;
