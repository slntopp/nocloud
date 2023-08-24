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

  const fetchRate = async () => {
    if (!currency.value || !defaultCurrency.value) {
      return;
    }
    if (currency.value === defaultCurrency.value) {
      rate.value = 1;
      return;
    }
    const res = await api.get(
      `/billing/currencies/rates/${currency.value}/${defaultCurrency.value}`
    );
    rate.value = res.rate;
  };

  watch(currency, () => {
    fetchRate();
  });

  watch(defaultCurrency, () => {
    fetchRate();
  });

  const convertFrom = (price) => {
    return rate.value ? (price / rate.value).toFixed(2) : 0;
  };
  const convertTo = (price) => {
    return rate.value ? (price * rate.value).toFixed(2) : 0;
  };

  return {
    rate,
    convertTo,
    convertFrom,
  };
};

export default useRate;
