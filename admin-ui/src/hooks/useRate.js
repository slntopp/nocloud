import api from "@/api";
import { computed, onMounted, ref, watch } from "vue";
import { useStore } from "@/store";

const useRate = (currency = ref("")) => {
  const store = useStore();

  const rate = ref(0);

  const defaultCurrency = computed(() => {
    return store.getters["currencies/default"];
  });

  onMounted(() => {
    fetchRate();
  });

  const fetchRate = async (fetchedCurrency) => {
    if (!fetchedCurrency) {
      fetchedCurrency = currency.value;
    }

    if (!fetchedCurrency || !defaultCurrency.value) {
      return;
    }
    if (fetchedCurrency === defaultCurrency.value) {
      rate.value = 1;
      return 1;
    }
    const res = await api.get(
      `/billing/currencies/rates/${fetchedCurrency}/${defaultCurrency.value}`
    );
    rate.value = res.rate;
    return res.rate;
  };

  watch(currency, () => {
    fetchRate();
  });

  watch(defaultCurrency, () => {
    fetchRate();
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
    fetchRate
  };
};

export default useRate;
